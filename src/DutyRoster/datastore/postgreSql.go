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
    "database/sql"
    _ "github.com/lib/pq"
    "github.com/jmoiron/sqlx"
    "DutyRoster/logging"
    "DutyRoster/config"
    "DutyRoster/errorset"
)

type postgreSqlDataStore struct {
    dblogger logging.LoggingInterface
    DBConn *sqlx.DB
}

var dbOnce sync.Once
var sqlObj = new(postgreSqlDataStore)

//Create a sql connection and store in the datastore object.
// Return '0' on success and errorcode otherwise.
// It is advised to make single connection in entire application.
func (sqlds *postgreSqlDataStore)CreateDBConnection() error{
    var err error
    dbconfig := config.GetConfigInstance()
    dbDriver := dbconfig.DB.Driver
    dbFile := dbconfig.DB.Dbname
    dbUser := dbconfig.DB.Uname
    dbPwd := dbconfig.DB.Pwd
    dbIpaddr := dbconfig.DB.Ipaddr
    dbPort := dbconfig.DB.Port

    //TODO :: Check if db connection present before creating.
    if (len(dbDriver) == 0 || len(dbFile) == 0) {
        sqlds.dblogger.Error("Failed to start application, NULL DB driver/name")
        return fmt.Errorf("%s",
            errorset.ERROR_TYPES[errorset.NULL_DB_CONFIG_PARAMS])
    }
    if dbDriver != "postgres" {
        //Only postgres driver can be handled here.
        sqlds.dblogger.Error("Failed to start application, Invalid driver :%s",
                             dbDriver)
        return fmt.Errorf("%s", errorset.ERROR_TYPES[errorset.INVALID_DB_DRIVER])
    }
    if len(dbUser) == 0 || len(dbPwd) == 0 || len(dbIpaddr) == 0 ||
        len(dbPort) == 0 {
            //Configuration file doesnt have user/pwd credentials
            sqlds.dblogger.Error(`Failed to start application, Invalid User
                                credentials`)
            return fmt.Errorf("%s",
                           errorset.ERROR_TYPES[errorset.INVALID_DB_CREDENTIALS])
    }
    dbparam := fmt.Sprintf(`host=%s port=%s user=%s password=%s
                            dbname=%s sslmode=disable`,
                            dbIpaddr, dbPort, dbUser, dbPwd,
                            dbFile)
    var dbHandle *sqlx.DB
    dbHandle, err = sqlx.Open(dbDriver, dbparam)
    if err != nil {
        sqlds.dblogger.Error("Failed to connect DB %s", err.Error())
        return err
    }
    sqlds.DBConn = dbHandle
    return nil
}

//Create all the postgresql tables for DutyRoster application.
func (sqlds *postgreSqlDataStore)CreateDataStoreTables() error {
    //Create Role table.

    roletable := new(sqlroles)
    roletable.createRoleTable(sqlds, sqlds.DBConn)
    orgtable := new(sqlorg)
    orgtable.createOrgTable(sqlds, sqlds.DBConn)
    usertable := new(sqlUsers)
    usertable.createUserTable(sqlds, sqlds.DBConn)
    return nil
}

func (sqlds *postgreSqlDataStore)CreateUserAccount(user *Users) error {
    usertable := new(sqlUsers)
    usertable.Users = *user
    Tx := sqlds.DBConn.MustBegin()
    err := usertable.createUserEntry(sqlds, Tx)
    if err != nil {
        Tx.Rollback()
        return err
    }
    Tx.Commit()
    return nil
}

func (sqlds *postgreSqlDataStore)GetUserAccount(user *Users) error {
    usertable := new(sqlUsers)
    usertable.Users = *user
    err := usertable.getUserwithIDPwd(sqlds, sqlds.DBConn)
    return err
}

func (sqlds *postgreSqlDataStore)DeleteUserAccount(user *Users) error {
    usertable := new(sqlUsers)
    usertable.Users = *user
    usertable.deleteUserEntry(sqlds, sqlds.DBConn)
    return nil
}

func (sqlds *postgreSqlDataStore)UpdateUserAccount(user *Users) error {
    usertable := new(sqlUsers)
    usertable.Users = *user
    Tx := sqlds.DBConn.MustBegin()
    err := usertable.createUserEntry(sqlds, Tx)
    if err != nil {
        Tx.Rollback()
        return err
    }
    Tx.Commit()
    return nil
}

// Exec operation on a postgreSQL DB can be either transactional or non-
// transactional. Helper function to find right exec function based on dbhandle
//type. Application not allowed to invoke db backend 'Exec' function. Instead
// it must use this function to invoke it in the specific context.
type sqlExecFn func (string, ...interface{}) (sql.Result, error)
func (sqlds *postgreSqlDataStore)getDBExecFunction(
                        handle interface{}) (sqlExecFn, error) {
    var dbhandle *sqlx.DB
    var dbtxhandle *sqlx.Tx
    var handleOk bool
    var execPtr sqlExecFn
    dbhandle, handleOk = handle.(*sqlx.DB)
    if handleOk {
        execPtr = dbhandle.Exec
    } else if dbtxhandle, handleOk = handle.(*sqlx.Tx); handleOk {
        execPtr = dbtxhandle.Exec
    } else {
        sqlds.dblogger.Error(
            "Failed to execute delete operation , Invalid DB handle")
        return nil, fmt.Errorf("%s",
                    errorset.ERROR_TYPES[errorset.INVALID_PARAM])

    }
    return execPtr, nil
}

// Get operation on a postgreSQL DB can be either transactional or non-
// transactional. Helper function to find right Get function based on dbhandle
//type. Application not allowed to invoke db backend 'Get' function. Instead
// it must use this function to invoke it in the specific context.
type sqlGetFn func (interface{}, string, ...interface{}) error
func (sqlds *postgreSqlDataStore)getDBGetFunction(
                        handle interface{}) (sqlGetFn, error) {
    var dbhandle *sqlx.DB
    var dbtxhandle *sqlx.Tx
    var handleOk bool
    var getPtr sqlGetFn
    dbhandle, handleOk = handle.(*sqlx.DB)
    if handleOk {
        getPtr = dbhandle.Get
    } else if dbtxhandle, handleOk = handle.(*sqlx.Tx); handleOk {
        getPtr = dbtxhandle.Get
    } else {
        sqlds.dblogger.Error(
            "Failed to execute delete operation , Invalid DB handle")
        return nil, fmt.Errorf("%s",
                    errorset.ERROR_TYPES[errorset.INVALID_PARAM])

    }
    return getPtr, nil
}

// Select operation on a postgreSQL DB can be either transactional or non-
// transactional. Helper function to find right Select function based on dbhandle
//type. Application not allowed to invoke db backend 'Select' function. Instead
// it must use this function to invoke it in the specific context.
type sqlSelectFn func (interface{}, string, ...interface{}) error
func (sqlds *postgreSqlDataStore)getDBSelectFunction(
                        handle interface{}) (sqlSelectFn, error) {
    var dbhandle *sqlx.DB
    var dbtxhandle *sqlx.Tx
    var handleOk bool
    var selectPtr sqlSelectFn
    dbhandle, handleOk = handle.(*sqlx.DB)
    if handleOk {
        selectPtr = dbhandle.Get
    } else if dbtxhandle, handleOk = handle.(*sqlx.Tx); handleOk {
        selectPtr = dbtxhandle.Get
    } else {
        sqlds.dblogger.Error(
            "Failed to execute delete operation , Invalid DB handle")
        return nil, fmt.Errorf("%s",
                    errorset.ERROR_TYPES[errorset.INVALID_PARAM])
    }
    return selectPtr, nil
}

// Only one SQL datastore object can be present in the system as connection
//pool can be handled in side the database connection itself
func getPSQLDataStoreObj() *postgreSqlDataStore {
    dbOnce.Do(func() {
        sqlObj.dblogger = logging.GetAppLoggerObj()
        sqlObj.dblogger.Trace("PostgreSql DB Object is created successfully")
        return
    })
    return sqlObj
}
