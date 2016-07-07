package notification

import (
	"database/sql"
	//"fmt"
	//"errors"

	//stat "github.com/asiainfoLDP/datahub_commons/statistics"
	//"github.com/asiainfoLDP/datahub_commons/log"
)



type DatabaseUpgrader_1 struct {
	DatabaseUpgrader_Base
	
	AlterSQL string
}

func newDatabaseUpgrader_1() *DatabaseUpgrader_1 {
	updater := &DatabaseUpgrader_1{}
	
	updater.currentTableCreationSqlFile = "initdb_v1.sql"
	
	updater.oldVersion = 0
	updater.newVersion = 1
	updater.AlterSQL = ``
	
	return updater
}

func (upgrader DatabaseUpgrader_1) Upgrade (db *sql.DB) error {
	return nil
}