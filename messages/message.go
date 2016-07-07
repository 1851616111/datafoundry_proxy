package messages

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	//"time"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/asiainfoLDP/datahub_commons/common"
	"github.com/asiainfoLDP/datahub_commons/message"
	
	"github.com/asiainfoLDP/datafoundry_proxy/messages/notification"
)

const ToUsersMessageTopic = "to_users.json"

//==================================================================
// 
//==================================================================

func getReceiverFromMap(m map[string]interface{}, multiple bool) (interface{}, int, *Error) {
	if multiple {
		receiver, e := mustStringParamInMap(m, "receiver", StringParamType_General)
		if e != nil {
			return nil, 0, e
		}
		
		if receiver == "*" {
			return receiver, 1, nil
		}
		
		receivers := EmailsString2EmailList(receiver)
		num := len(receivers)
		index := 0
		for i := 0; i < num; i ++ {
			email := receivers[i]
			rcvr, ok := common.ValidateEmail(email)
			if !ok {
				return nil, 0, GetError2(ErrorCodeInvalidParameters, "bad email: " + email)
			}
			receivers[index] = rcvr
			index ++
		}
		return receivers[:index], index, nil
	}
	
	r, e := mustStringParamInMap(m, "receiver", StringParamType_Email)
	return r, 1, e
}

func sendAdminBroadcastMessage(currentUserName string, level int, message_data string) error {
	mq := theMQ
	if mq == nil {
		return errors.New("theMQ == nil")
	}
	
	/*
	msg := &message.Message{
			Type:   "admin_message",
			Receiver: "*",
			Sender: currentUserName,
			Level: level,
			Data: message_data,
		}
	
	err := message.PushMessageToQueue(mq, GeneralMessageTopic, []byte("subscriptions"), msg)
	if err != nil {
		
	}
	*/
	
	msg := &message.Message{
			Type:   "0x00030000",
			Sender: currentUserName,
			Level: level,
			Data: message_data,
		}
	
	err := message.PushMessageToQueue(mq, ToUsersMessageTopic, []byte("subscriptions"), msg)
	if err != nil {
		
	}

	return err
}

//==================================================================
// 
//==================================================================

func CreateMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	currentUserName, e := mustCurrentUserName(r) // sender, may be a fake username, such as *, which is not email.
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	m, err := common.ParseRequestJsonAsMap(r)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeInvalidParameters, err.Error()), nil)
		return
	}

	data, ok := m["data"]
	if !ok {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeInvalidParameters, "data is not specified"), nil)
		return
	}

	message_type, e := mustStringParamInMap(m, "type", StringParamType_UrlWord) // must url word, for it is used in sql
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}
	
	var receiver interface{} = nil
	var level = notification.Level_General
	
	switch message_type {
	default:
		JsonResult(w, http.StatusOK, GetError2(ErrorCodeInvalidParameters, "unkown type: " + message_type), nil)
	case "admin_message":
		num := 0
		receiver, num, e = getReceiverFromMap(m, true)
		if e != nil {
			JsonResult(w, http.StatusUnauthorized, e, nil)
			return
		}
		if num < 1 || num > 100 {
			JsonResult(w, http.StatusOK, GetError2(ErrorCodeInvalidParameters, fmt.Sprintf("number of receivers must be in range [1,100], now %d", num)), nil)
			return
		}
		
		/*
		user, err := GetUserInfo(getUserService(), r.Header.Get("Authorization"), r.Header.Get("User"), currentUserName)
		if err != nil {
			JsonResult(w, http.StatusInternalServerError, GetError2(ErrorCodeInvalidParameters, err.Error()), nil)
			return
		}
		if user.UserType != UserType_Admin {
			JsonResult(w, http.StatusUnauthorized, GetError2(ErrorCodeInvalidParameters, "only admin can call this API"), nil)
			return
		}
		*/
	}
	
	if receiver == nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeInvalidParameters, "can't get determine reciver"), nil)
		return
	}

	// ...

	json_data, err := json.Marshal(&data)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeInvalidParameters, "input data can't be marshalled"), nil)
		return
	}
	message_data := string(json_data)
	
	switch rrr := receiver.(type) {
	default:
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeInvalidParameters, "bad reciver"), nil)
		return
	case string:
		if rrr == "*" {
			// todo: send message to user
			err := sendAdminBroadcastMessage(currentUserName, level, message_data)
			if err != nil {
				JsonResult(w, http.StatusInternalServerError, GetError2(ErrorCodeCreateMessage, err.Error()), nil)
				go func(){
					Logger.Warningf("send message {%s, %s, %s, %d, %s} error: %s",  message_type, rrr, currentUserName, level, message_data, err.Error())
				}()
				return
			}
		} else {
			_, err := notification.CreateMessageForBrowser(db, message_type, rrr, currentUserName, level, message_data)
			if err != nil {
				JsonResult(w, http.StatusInternalServerError, GetError2(ErrorCodeCreateMessage, err.Error()), nil)
				go func(){
					Logger.Warningf("send message {%s, %s, %s, %d, %s} error: %s",  message_type, rrr, currentUserName, level, message_data, err.Error())
				}()
				return
			}
		}
	case []string:
		go func() {
			for _, rcvr := range rrr {
				_, err := notification.CreateMessageForBrowser(db, message_type, rcvr, currentUserName, level, message_data)
				if err != nil {
					Logger.Warningf("batch send message {%s, %s, %s, %d, %s} error: %s",  message_type, rcvr, currentUserName, level, message_data, err.Error())
				}
			}
		}()
	}

	JsonResult(w, http.StatusOK, nil, nil)
}

/*
func ModifyMessage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	currentUserName, e := mustCurrentUserName(r)
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	messageid, e := mustIntParamInPath(params, "messageid")
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	action, e := mustStringParamInPath(params, "action", StringParamType_UrlWord)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	err := notification.ModifyUserMessage(db, currentUserName, messageid, action)
	if err != nil {
		JsonResult(w, http.StatusInternalServerError, GetError2(ErrorCodeModifyMessage, err.Error()), nil)
		return
	}

	JsonResult(w, http.StatusOK, nil, nil)
}
*/

func GetMyMessages(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	currentUserName, e := mustCurrentUserName(r)
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	r.ParseForm()
	
	forclient := false
	if r.Form.Get("forclient") != "" {
		forclient, e = mustBoolParamInQuery(r, "forclient")
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	level := notification.Level_Any
	if r.Form.Get("level") != "" {
		lvl, e := mustIntParamInQuery(r, "level")
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
		
		level = int(lvl)
	}

	status := notification.Status_Either
	if r.Form.Get("status") != "" {
		stts, e := mustIntParamInQuery(r, "status")
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
		status = int(stts)
		if status != notification.Status_Unread && status != notification.Status_Read {
			status = notification.Status_Either
		}
	}

	// message_type can be ""
	message_type := r.Form.Get("type")
	if message_type != "" {
		message_type, e = mustStringParamInQuery(r, "type", StringParamType_UrlWord)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	// sender can be ""
	sender := r.Form.Get("sender")
	if sender != "" {
		// the sender may be email, or some special word, ex, @zhang3#aaa.com, $system, ....
		sender, e = mustStringParamInQuery(r, "sender", StringParamType_UnicodeUrlWord) //StringParamType_EmailOrUrlWord)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}
	
	JsonResult(w, http.StatusNotImplemented, GetError(ErrorCodeUrlNotSupported), nil)
	return
	/*
	bt := r.Form.Get("beforetime")
	at := r.Form.Get("aftertime")
	if bt != "" && at != "" {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeInvalidParameters, "beforetime and aftertime can't be both specified"), nil)
		return
	}
	
	var beforetime *time.Time = nil
	if bt != "" {
		// beforetime = &(optionalTimeParamInQuery(r, "beforetime", time.RFC3339, time.Now().Add(32*time.Hour)))
		// shit! above line doesn't work in golang
		t := optionalTimeParamInQuery(r, "beforetime", time.RFC3339, time.Now().Add(32*time.Hour))
		beforetime = &t
	}
	var aftertime *time.Time = nil
	if at != "" {
		t, _ := time.Parse("2006-01-02", "2000-01-01")
		t = optionalTimeParamInQuery(r, "aftertime", time.RFC3339, t)
		aftertime = &t
	}
	*/
	
	offset, size := optionalOffsetAndSize(r, 30, 1, 100)

	if forclient {
		//client_messages, err := notification.GetUserMessagesForClient(db, currentUserName, message_type, status, sender, beforetime, aftertime)
		count, client_messages, err := notification.GetUserMessagesForClient(db, currentUserName, message_type, level, status, sender, offset, size)
		if err != nil {
			JsonResult(w, http.StatusInternalServerError, GetError2(ErrorCodeQueryMessage, err.Error()), nil)
			return
		}
	
		JsonResult(w, http.StatusOK, nil, newQueryListResult(count, client_messages))
	} else {
		// /browser_messages, err := notification.GetUserMessagesForBrowser(db, currentUserName, message_type, status, sender, beforetime, aftertime)
		count, browser_messages, err := notification.GetUserMessagesForBrowser(db, currentUserName, message_type, level, status, sender, offset, size)
		if err != nil {
			JsonResult(w, http.StatusInternalServerError, GetError2(ErrorCodeQueryMessage, err.Error()), nil)
			return
		}
		JsonResult(w, http.StatusOK, nil, newQueryListResult(count, browser_messages))
	}
}

func GetNotificationStats(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	currentUserName, e := mustCurrentUserName(r)
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	r.ParseForm()

	// category can be ""
	category := r.Form.Get("category")
	if category != "" {
		category, e = mustStringParamInQuery(r, "category", StringParamType_UrlWord)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}
	stat_category := notification.StatCategory_Unknown
	switch category {
	case "", "type":
		stat_category = notification.StatCategory_MessageType
	case "level":
		stat_category = notification.StatCategory_MessageLevel
	}
	if stat_category == notification.StatCategory_Unknown {
		JsonResult(w, http.StatusBadRequest, newInvalidParameterError("bad category param"), nil)
		return
	}

	message_stats, err := notification.RetrieveUserMessageStats_ForBrowser(db, currentUserName, stat_category)
	if err != nil {
		JsonResult(w, http.StatusInternalServerError, GetError2(ErrorCodeGetMessageStats, err.Error()), nil)
		return
	}

	if len(message_stats) == 0 {
		message_stats = nil
	}

	JsonResult(w, http.StatusOK, nil, message_stats)
}

func ClearNotificationStats(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	currentUserName, e := mustCurrentUserName(r)
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}
	
	_ = currentUserName
	JsonResult(w, http.StatusNotImplemented, GetError(ErrorCodeUrlNotSupported), nil)

	//err := notification.UpdateUserMessageStats(db, currentUserName, "", 0) // clear
	//if err != nil {
	//	JsonResult(w, http.StatusInternalServerError, GetError2(ErrorCodeResetMessageStats, err.Error()), nil)
	//	return
	//}

	//JsonResult(w, http.StatusOK, nil, nil)
}

//==================================================================
// 
//==================================================================

func HandleNotificationsFromQueue(topic string, key, value []byte) error {
	if len(key) == 0 && len(value) == 0 {
		return nil
	}

	db := getDB()
	if db == nil {
		return errors.New("db is not inited")
	}

	msg, err := message.ParseJsonMessage(value)
	if err != nil {
		return err
	}

	json_bytes, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}
	
	var forBrowser, forClient bool
	if len(key) == 0 {
		forBrowser = true
		forClient = false
	} else {
		keystr := strings.ToLower(string(key))
		forBrowser = strings.Index(keystr, "notforbrowser") < 0
		forClient = strings.Index(keystr, "forclient") >= 0
	}
	
	if forClient {
		_, err = notification.CreateMessageForClient(db, msg.Type, msg.Receiver, msg.Sender, msg.Level, string(json_bytes))
		if err != nil {
			Logger.Warningf("CreateMessageForClient error: %s\nMessage=%s", err.Error(), string(value))
		} else {
			Logger.Warningf("CreateMessageForClient succeeded: %s", string(value))
		}
	}

	if forBrowser {
		_, err = notification.CreateMessageForBrowser(db, msg.Type, msg.Receiver, msg.Sender, msg.Level, string(json_bytes))
		if err != nil {
			Logger.Warningf("CreateMessageForBrowser error: %s\nMessage=%s", err.Error(), string(value))
			return err
		} else {
			Logger.Warningf("CreateMessageForBrowser succeeded: %s", string(value))
		}
	}
	
	return nil
}

//==================================================================
//
//==================================================================

func HandleEmailsFromQueue(topic string, key, value []byte) error {
	if len(key) == 0 && len(value) == 0 {
		return nil
	}

	email, err := message.ParseJsonEmail(value)
	if err != nil {
		return err
	}

	return SendMail(EmailsString2EmailList(email.To), nil, nil, 
			email.Subject, email.Content, email.IsHTML)
}
