package notification

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"bytes"
	
	stat "github.com/asiainfoLDP/datahub_commons/statistics"
	"github.com/asiainfoLDP/datahub_commons/log"
)

//=============================================================
//
//=============================================================

var dbUpgraders = []DatabaseUpgrader {
	newDatabaseUpgrader_1(),
}

const (
	DbPhase_Unkown    = -1
	DbPhase_Serving   = 0 // must be 0
	DbPhase_Upgrading = 1
)

var dbPhase = DbPhase_Unkown

func IsServing() bool {
	return dbPhase == DbPhase_Serving
}

// for ut, reallyNeedUpgrade is false
func TryToUpgradeDatabase(db *sql.DB, dbName string, reallyNeedUpgrade bool) error {
	
	if reallyNeedUpgrade {
	
		if len(dbUpgraders) == 0 {
			return errors.New("at least one db upgrader needed")
		}
		lastDbUpgrader := dbUpgraders[len(dbUpgraders) - 1]
		
		// create tables. (no table created in following _upgradeDatabase callings)
		
		err := lastDbUpgrader.TryToCreateTables(db)
		if err != nil {
			return err
		}
		
		// init version value stat as LatestDataVersion if it doesn't exist,
		// which means tables are just created. In this case, no upgrations are needed.
		
		// INSERT INTO DH_ITEM_STAT (STAT_KEY, STAT_VALUE) 
		// VALUES(dbName#version, LatestDataVersion) 
		// ON DUPLICATE KEY UPDATE STAT_VALUE=LatestDataVersion;
		dbVersionKey := stat.GetVersionKey(dbName)
		_, err = stat.SetStatIf(db, dbVersionKey, lastDbUpgrader.NewVersion(), 0)
		if err != nil && err != stat.ErrOldStatNotMatch {
			return err
		}
		
		current_version, err := stat.RetrieveStat(db, dbVersionKey)
		if err != nil {
			return err
		}
		
		// upgrade 
		
		if current_version != lastDbUpgrader.NewVersion() {
		
			log.DefaultLogger().Info("mysql start upgrading ...")
		
			dbPhase = DbPhase_Unkown
			
			for _, dbupgrader := range dbUpgraders {
				if err = _upgradeDatabase(db, dbName, dbupgrader); err != nil {
					return err
				}
			}
		}
	}
	
	dbPhase = DbPhase_Serving
	
	log.DefaultLogger().Info("mysql start serving ...")
	
	return nil
}

func _upgradeDatabase(db *sql.DB, dbName string, upgrader DatabaseUpgrader) error {
	dbVersionKey := stat.GetVersionKey(dbName)
	current_version, err := stat.RetrieveStat(db, dbVersionKey)
	if err != nil {
		return err
	}
	if current_version == 0 {
		current_version = 1
	}
	
	log.DefaultLogger().Info("TryToUpgradeDatabase current version: ", current_version) 
	
	if upgrader.NewVersion() <= current_version {
		return nil
	}
	if upgrader.OldVersion() != current_version {
		return fmt.Errorf("old version (%d) <= current version (%d)", upgrader.OldVersion(), current_version)
	}
	
	dbPhaseKey := stat.GetPhaseKey(dbName)
	phase, err := stat.SetStatIf(db, dbPhaseKey, DbPhase_Upgrading, DbPhase_Serving)

	log.DefaultLogger().Info("TryToUpgradeDatabase current phase: ", phase) 
	
	if err != nil {
		return err
	}
	
	// ...
	
	dbPhase = DbPhase_Upgrading
	
	err = upgrader.Upgrade(db)
	if err != nil {
		return err
	}
	
	// ...
	
	_, err = stat.SetStat(db, dbVersionKey, upgrader.NewVersion())
	if err != nil {
		return err
	}
	
	log.DefaultLogger().Info("TryToUpgradeDatabase new version: ", upgrader.NewVersion()) 
	
	time.Sleep(30 * time.Millisecond)
	
	_, err = stat.SetStatIf(db, dbPhaseKey, DbPhase_Serving, DbPhase_Upgrading)
	if err != nil {
		return err
	}
	
	return nil
}

type DatabaseUpgrader interface {
	OldVersion() int
	NewVersion() int
	Upgrade(db *sql.DB) error
	TryToCreateTables(db *sql.DB) error
}

type DatabaseUpgrader_Base struct {
	oldVersion int
	newVersion int
	
	currentTableCreationSqlFile string
}

func (upgrader DatabaseUpgrader_Base) OldVersion() int {
	return upgrader.oldVersion
}

func (upgrader DatabaseUpgrader_Base) NewVersion() int {
	return upgrader.newVersion
}

func (upgrader DatabaseUpgrader_Base) TryToCreateTables(db *sql.DB) error {
	
	if upgrader.currentTableCreationSqlFile == "" {
		return nil
	}
	
	data, err := ioutil.ReadFile(filepath.Join("_db", upgrader.currentTableCreationSqlFile))
	if err != nil {
		return err
	}
	
	sqls := bytes.SplitAfter(data, []byte("DEFAULT CHARSET=UTF8;"))
	sqls = sqls[:len(sqls)-1]
	for _, sql := range sqls {
		_, err = db.Exec(string(sql))
		if err != nil {
			return err
		}
	}
	
	return nil
}

//=============================================================
//
//=============================================================

const (
	Status_Unread = 0
	Status_Read   = 1
	Status_Either = 2
)

const (
	Level_Any = -1
	Level_General = 0
	Level_Request = 50
)

const (
	StatCategory_Unknown      = 0
	StatCategory_MessageType  = 1
	StatCategory_MessageLevel = 2
)

//=============================================================
//
//=============================================================

type Message struct {
	Message_id   int64     `json:"messageid"`
	Message_type string    `json:"type"`
	Level        int       `json:"level"`
	Status       int       `json:"status"`
	Time         time.Time `json:"time"`
	Receiver     string    `json:"receiver"`
	Sender       string    `json:"sender,omitempty"`
	Hints        string    `json:"Hints,omitempty"`
	Raw_data     string    `json:"-"` // raw_data
	Json_data    interface{} `json:"data"` // json.Unmarshal(Raw_data)
}

const (
	MessageTableName_ForBrowser = "DF_MESSAGE"
)

//=============================================================
//
//=============================================================

func CreateMessage(db *sql.DB, messageType, receiver, sender string, level int, hints, jsonData string) (int64, error) {
	return createMessage(db, MessageTableName_ForBrowser, messageType, receiver, sender, level, hints, jsonData)
}

func createMessage(db *sql.DB, tableName, messageType, receiver, sender string, level int, hints, jsonData string) (int64, error) {
	nowstr := time.Now().Format("2006-01-02 15:04:05.999999")
	sqlstr := fmt.Sprintf(`insert into %s (
						TYPE, STATUS, LEVEL, TIME,
						RECEIVER, SENDER, HINTS, DATA
						) values (
						'%s', %d, %d, '%s', 
						'%s', '%s', '%s', ?
						)`,
		tableName,
		messageType, Status_Unread, level, nowstr,
		receiver, sender, hints)
	result, err := db.Exec(sqlstr, jsonData)
	if err != nil {
		return 0, err
	}

	// assert result.RowsAffected () == 1
	//go func() {
	//	if tableName == MessageTableName_ForBrowser {
	//		UpdateUserMessageStats(db, receiver, messageType, 1)
	//	}
	//}()

	id, _ := result.LastInsertId()
	return id, nil
}

func DeleteUserMessage(db *sql.DB, currentUserName string, messageid int64) error {
	return deleteUserMessage(db, currentUserName, MessageTableName_ForBrowser, messageid)
}

func deleteUserMessage(db *sql.DB, currentUserName string, tableName string, messageid int64) error {
	sqlstr := fmt.Sprintf(`delete from %s
							where RECEIVER='%s' and MESSAGE_ID=%d
							`, tableName, currentUserName, messageid)
	result, err := db.Exec(sqlstr)
	if err != nil {
		return err
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return errors.New ("failed to delete")
	}

	return nil
}

func ModifyUserMessage(db *sql.DB, currentUserName string, messageid int64, action string) (bool, error) {
	return modifyUserMessage(db, currentUserName, MessageTableName_ForBrowser, messageid, action)
}

const (
	Action_SetRead   = "read"
	Action_SetUnread = "unread"
)

// now action can only be "read", "unread"
// return bool means handled or not
func modifyUserMessage(db *sql.DB, currentUserName string, tableName string, messageid int64, action string) (bool, error) {
	status := Status_Either
	if action == Action_SetRead {
		status = Status_Read
	} else if action == Action_SetUnread {
		status = Status_Unread
	}

	if status == Status_Either {
		return false, nil
	}

	tx, err := db.Begin()
	if err != nil {
		return true, err
	}

	sqlget := fmt.Sprintf(`select RECEIVER, STATUS from %s where MESSAGE_ID=%d`, tableName, messageid)
	old_status := Status_Either
	receiver := ""
	err = tx.QueryRow(sqlget).Scan(&receiver, &old_status)
	if err != nil {
		tx.Rollback()
		return true, err
	}

	if receiver != currentUserName {
		tx.Rollback()
		return true, errors.New("user can't modify message of other users")
	}

	if old_status == status {
		tx.Rollback()
		return true, nil
	}

	sqlupdate := fmt.Sprintf(`update %s set STATUS=%d where MESSAGE_ID=%d`, tableName, status, messageid)
	_, err = tx.Exec(sqlupdate)
	if err != nil {
		tx.Rollback()
		return true, err
	}

	tx.Commit()

	return true, nil
}

func ModifyMessageDataByID(db *sql.DB, messageid int64, jsonData string) error {
	return modifyMessageDataByID(db, MessageTableName_ForBrowser, messageid, jsonData)
}

func modifyMessageDataByID(db *sql.DB, tableName string, messageid int64, jsonData string) error {
	sqlupdate := fmt.Sprintf(`update %s set DATA=? where MESSAGE_ID=%d`, tableName, messageid)
	_, err := db.Exec(sqlupdate, jsonData)
	return err
}

func GetMessageByUserAndID(db *sql.DB, currentUserName string, messageid int64) (*Message, error) {
	return getMessageByUserAndID(db, currentUserName, MessageTableName_ForBrowser, messageid)
}

func getMessageByUserAndID(db *sql.DB, currentUserName string, tableName string, messageid int64) (*Message, error) {
	sqlwhere := fmt.Sprintf(" MESSAGE_ID=%d", messageid)
	if currentUserName != "" {
		sqlwhere = sqlwhere + fmt.Sprintf(" and RECEIVER='%s'", currentUserName)
	}
	
	messages, err := queryMessages(db, tableName, sqlwhere, 1, 0)
	if err != nil {
		return nil, err
	}
	
	if len(messages) == 0 {
		return nil, errors.New("not found")
	}
	
	return messages[0], nil
}

func GetUserMessages(db *sql.DB, username string, messagetype string, level int, status int, sender string, offset int64, limit int) (int64, []*Message, error) {
	return getUserMessages(db, username, MessageTableName_ForBrowser, messagetype, level, status, sender, offset, limit)
}

func validateOffsetAndLimit(count int64, offset *int64, limit *int) {
	if *limit < 1 {
		*limit = 1
	}
	if *offset >= count {
		*offset = count - int64(*limit)
	}
	if *offset < 0 {
		*offset = 0
	}
	if *offset + int64(*limit) > count {
		*limit = int(count - *offset)
	}
}

//func getUserMessages(db *sql.DB, username string, tableName string, messagetype string, status int, sender string, beforetime *time.Time, aftertime *time.Time) ([]*Message, error) {
func getUserMessages(db *sql.DB, username string, tableName string, messagetype string, level int, status int, sender string, offset int64, limit int) (int64, []*Message, error) {
	sqlwhere := fmt.Sprintf(" RECEIVER='%s'", username)
	if messagetype != "" {
		sqlwhere = sqlwhere + fmt.Sprintf(" and TYPE='%s'", messagetype)
	}
	if status != Status_Either {
		sqlwhere = sqlwhere + fmt.Sprintf(" and STATUS=%d", status)
	}
	if sender != "" {
		sqlwhere = sqlwhere + fmt.Sprintf(" and SENDER='%s'", sender)
	}
	//if aftertime != nil {
	//	sqlwhere = sqlwhere + fmt.Sprintf(" and TIME >= '%s' order by TIME asc", aftertime.Format("2006-01-02 15:04:05.999999"))
	//} else {
	//	if beforetime != nil {
	//		sqlwhere = sqlwhere + fmt.Sprintf(" and TIME <= '%s'", beforetime.Format("2006-01-02 15:04:05.999999"))
	//	}
	//	sqlwhere = sqlwhere + " order by TIME desc"
	//}
	if level >= 0 {
		sqlwhere = sqlwhere + fmt.Sprintf(" and LEVEL=%d", level)
	}
	sqlwhere = sqlwhere + " order by TIME desc"
	
	count, err := queryMessagesCount(db, tableName, sqlwhere)
	if err != nil {
		return 0, nil, err
	}
	if count == 0 {
		return 0, []*Message{}, nil
	}
	validateOffsetAndLimit(count, &offset, &limit)
	
	messages, err := queryMessages(db, tableName, sqlwhere, limit, offset)
	if err != nil {
		return 0, nil, err
	}
	
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		
		_ = json.Unmarshal([]byte(msg.Raw_data), &msg.Json_data)
		
		//msg.Receiver = ""

		//if messagetype != "" {
		//	msg.Message_type = ""
		//}

		//if status != Status_Either {
		//	msg.Status = Status_Either // for json to omit
		//}

		//if sender != "" {
		//	msg.Sender = ""
		//}
	}
	
	go func(){
		num := len(messages)
		for i := 0; i < num; i++ {
			msg := messages[i]
			if msg.Status == Status_Unread {
				modifyUserMessage(db, username, tableName, msg.Message_id, Action_SetRead)
				//UpdateUserMessageStats(db, username, msg.Message_type, -1)
			}
		}
	}()
	
	return count, messages, nil
}

func queryMessagesCount(db *sql.DB, tableName string, sqlWhere string) (int64, error) {
	count := int64(0)
	
	sql_str := fmt.Sprintf(`select COUNT(*) from %s where %s`, tableName, sqlWhere)
	err := db.QueryRow(sql_str).Scan(&count)
	
	return count, err
}

func queryMessages(db *sql.DB, tableName string, sqlWhere string, limit int, offset int64) ([]*Message, error) {
	offset_str := ""
	if offset > 0 {
		offset_str = fmt.Sprintf("offset %d", offset)
	}
	sql_str := fmt.Sprintf(`select 
					MESSAGE_ID,
					TYPE, STATUS, LEVEL, TIME,
					RECEIVER, SENDER, HINTS, DATA
					from %s 
					where %s
					limit %d
					%s
					`, 
					tableName, 
					sqlWhere, 
					limit, 
					offset_str)

	rows, err := db.Query(sql_str)
	switch {
	case err == sql.ErrNoRows:
		return []*Message{}, nil
	case err != nil:
		return nil, err
	}
	defer rows.Close()

	messages := make([]*Message, limit)
	num := 0
	for rows.Next() {
		msg := &Message{}
		if err := rows.Scan(
			&msg.Message_id,
			&msg.Message_type, &msg.Status, &msg.Level, &msg.Time,
			&msg.Receiver, &msg.Sender, &msg.Hints, &msg.Raw_data); err != nil {
			return nil, err
		}
		if num >= len(messages) {
			messages = append(messages, msg)
		} else {
			messages[num] = msg
		}
		num++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages[:num], nil
}

func RetrieveUserMessageStats(db *sql.DB, username string, catrgory int) (map[string]int, error) {
	return retrieveUserMessageStats(db, MessageTableName_ForBrowser, username, catrgory)
}

func retrieveUserMessageStats(db *sql.DB, tableName, username string, category int) (map[string]int, error) {
	groupby := ""
	switch category {
	default:
		return nil, nil
	case StatCategory_MessageType:
		groupby = "TYPE"
	case StatCategory_MessageLevel:
		groupby = "LEVEL"
	}
	
	sqlstr := fmt.Sprintf(
			`select %s, COUNT(*) FROM %s 
			where RECEIVER='%s' and STATUS=%d
			group by %s`,
			groupby, tableName,
			username, Status_Unread,
			groupby)
			
	rows, err := db.Query(sqlstr)
	switch {
	case err == sql.ErrNoRows:
		return map[string]int{}, nil
	case err != nil:
		return nil, err
	}
	defer rows.Close()
	
	stats := map[string]int{}
	for rows.Next() {
		ttt, ccc := "", 0
		if err := rows.Scan(&ttt, &ccc); err != nil {
			return nil, err
		}
		stats[ttt] = ccc
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return stats, nil
}
