// Copyright 2018 Sugesh Chandran
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
    "sync"
    "fmt"
    "path/filepath"
    _ "github.com/mattn/go-sqlite3"
    "github.com/jmoiron/sqlx"
    "DutyRoster/logging"
    "DutyRoster/config"
    "DutyRoster/errorset"
)

type sqlLiteDataStore struct {
    dblogger logging.LoggingInterface
    DBConn *sqlx.DB
}

var dbOnce sync.Once
var sqlObj = new(sqlLiteDataStore)

//Create a sql connection and store in the datastore object.
// Return '0' on success and errorcode otherwise.
// It is advised to make single connection in entire application.
func (sqlds *sqlLiteDataStore)CreateDBConnection() error{
    var err error
    dbconfig := config.GetConfigInstance()
    dbDriver := dbconfig.DB.Driver
    dbFile := dbconfig.DB.Dbpath

    //TODO :: Check if db connection present before creating.
    if (len(dbDriver) == 0 || len(dbFile) == 0) {
        sqlds.dblogger.Error("Failed to start application, NULL DB driver/path")
        return fmt.Errorf("%s",
            errorset.ERROR_TYPES[errorset.NULL_DB_CONFIG_PARAMS])
    }
    if dbDriver != "sqlite3" {
        //Only sqllite3 driver can be handled here.
        sqlds.dblogger.Error("Failed to start application, Invalid driver :%s",
                             dbDriver)
        return fmt.Errorf("%s", errorset.ERROR_TYPES[errorset.INVALID_DB_DRIVER])
    }
    dbFile, err = filepath.Abs(dbFile)
    if err != nil {
        sqlds.dblogger.Error("Failed to open DB file, %s", err.Error())
        return err
    }
    var dbHandle *sqlx.DB
    dbHandle, err = sqlx.Open(dbDriver, dbFile)
    if err != nil {
        sqlds.dblogger.Error("Failed to connect DB %s", err.Error())
        return err
    }
    sqlds.DBConn = dbHandle
    return nil
}

//Create all the sqllite tables for DutyRoster application.
func (sqlds *sqlLiteDataStore)CreateDataStoreTables() error {
    //Create Role table.
    roletable := new(sqlroles)
    roletable.createRoleTable(sqlds.DBConn)
    //Create Org table
    orgtable := new(sqlorg)
    orgtable.createOrgTable(sqlds.DBConn)
    return nil
}

// Only one SQL datastore object can be present in the system as connection
//pool can be handled in side the database connection itself
func getSqlLiteDataStoreObj() *sqlLiteDataStore {
    dbOnce.Do(func() {
        sqlObj.dblogger = logging.GetAppLoggerObj()
        sqlObj.dblogger.Trace("SQLLite DB Object is created successfully")
        return
    })
    return sqlObj
}