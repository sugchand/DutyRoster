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
    "fmt"
    "time"
    "database/sql"
    _ "github.com/lib/pq"
    "DutyRoster/errorset"
    "DutyRoster/logging"
)

const (
    USER_STR_LEN = 1000
    USER_TABLE_NAME = "users"
    USER_FIELD_USERID = "userid"
    USER_FIELD_EMAILID = "emailid"
    USER_FIELD_HASHPWD = "hashpwd"
    USER_FIELD_MOBILENO = "mobileno"
    USER_FIELD_DOB = "dob"
    USER_FIELD_STARTTIME = "starttime"
    USER_FIELD_VALIDITY = "validity"
    USER_FIELD_STATUS = "status"
)

// SQLX representation of user record. The go representation of user cannot use
// in sql statements. This table has 1:1 mapping to the users record.
type sqlDBUsers struct {
    Userid string `db:"userid"`
    Emailid string `db:"emailid"`
    Hashpwd string `db:"hashpwd"`
    Mobileno string `db:"mobileno"`
    Dob time.Time `db:"dob"`
    StartTime time.Time `db:"starttime"`
    Validity sql.NullInt64 `db:"validity"`
    Status uint64 `db:"status"`
}

type sqlUsers struct {
    Users
}

// SQL statements to be used to operate on user table.
var (
    //Create a table users
    userchema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
                    (%s varchar(%d) NOT NULL PRIMARY KEY,
                     %s varchar(%d) NOT NULL,
                     %s varchar(%d) NOT NULL,
                     %s varchar(%d) NOT NULL,
                     %s date NOT NULL,
                     %s timestamp NOT NULL,
                     %s bigint,
                     %s bigint NOT NULL CHECK(%s > 0));`,
                     USER_TABLE_NAME,
                     USER_FIELD_USERID, USER_STR_LEN,
                     USER_FIELD_EMAILID, USER_STR_LEN,
                     USER_FIELD_HASHPWD, USER_STR_LEN,
                     USER_FIELD_MOBILENO, USER_STR_LEN,
                     USER_FIELD_DOB,
                     USER_FIELD_STARTTIME,
                     USER_FIELD_VALIDITY,
                     USER_FIELD_STATUS, USER_FIELD_STATUS)

    //Create a user row entry in table User
    userCreate = fmt.Sprintf(`INSERT INTO %s
                            (%s, %s, %s, %s, %s, %s, %s, %s)
                            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
                            USER_TABLE_NAME,
                            USER_FIELD_USERID, USER_FIELD_EMAILID,
                            USER_FIELD_HASHPWD, USER_FIELD_MOBILENO,
                            USER_FIELD_DOB, USER_FIELD_STARTTIME,
                            USER_FIELD_VALIDITY, USER_FIELD_STATUS)
    //Get the user rows with specific userID
    userGetonUserID = fmt.Sprintf(`SELECT * FROM %s WHERE %s=($1)`,
                            USER_TABLE_NAME, USER_FIELD_USERID)
    //Get the user rows with specific userid and pwd
    userGetonUserIDPwd = fmt.Sprintf(`SELECT * FROM %s WHERE %s=($1)
                                    AND %s=($2)`,
                             USER_TABLE_NAME, USER_FIELD_USERID,
                             USER_FIELD_HASHPWD)
    //Get the user rows with specific emailid and pwd
    userGetonEmailIDPwd = fmt.Sprintf(`SELECT * FROM %s
                             WHERE %s=($1) AND %s=($2)`,
                             USER_TABLE_NAME, USER_FIELD_EMAILID,
                             USER_FIELD_HASHPWD)
    //Delete user record from users table with specific userid
    userDeleteOnID = fmt.Sprintf("DELETE FROM %s WHERE %s=($1)",
                                USER_TABLE_NAME,
                                USER_FIELD_USERID)
    userUpdateOnID = fmt.Sprintf(`UPDATE %s SET %s=($1), %s=($2),
                        %s=($3), %s=($4), %s=($5)`,
                        USER_TABLE_NAME,
                        USER_FIELD_EMAILID,
                        USER_FIELD_HASHPWD,
                        USER_FIELD_MOBILENO,
                        USER_FIELD_STATUS,
                        USER_FIELD_VALIDITY)
)

func (user *sqlUsers)createUserTable(sqlds *postgreSqlDataStore,
                                     handle interface{}) error{
    log := logging.GetAppLoggerObj()
    execPtr, err := sqlds.getDBExecFunction(handle)
    if err != nil {
        log.Error("Failed to create user table, invalid DB handle err : %s",
                   err)
        return err
    }
    _, err = execPtr(userchema)
    if err != nil {
        log.Error("Failed to create User table %s", err)
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.DB_TABLE_CREATE_FAILED])
    }
    return nil
}

//Function to translate sqlUser elements to DB format to operate on DB
func (user *sqlUsers)usertoDBRowXlate() *sqlDBUsers {
    dbuser := new(sqlDBUsers)
    dbuser.Userid = user.userid
    dbuser.Emailid = user.emailid
    dbuser.Hashpwd = user.hashpwd
    dbuser.Mobileno = user.mobileno
    dbuser.StartTime = user.startTime
    dbuser.Status = uint64(user.status)
    dbuser.Dob = user.dob
    dbuser.Validity.Scan(user.validity)
    return dbuser
}

//Function to transalte DB row data to sqlusers format.
func (user *sqlUsers)DBtoUserRowXlate(dbrow *sqlDBUsers) {
    user.userid = dbrow.Userid
    user.hashpwd = dbrow.Hashpwd
    user.emailid = dbrow.Emailid
    user.dob = dbrow.Dob
    user.mobileno = dbrow.Mobileno
    user.startTime = dbrow.StartTime
    user.status = userStatusBit(dbrow.Status)
    validity, res :=dbrow.Validity.Value()
    if res != nil {
        user.validity, _ = validity.(uint64)
    }
}

func (user *sqlUsers)getUserwithIDPwd(sqlds *postgreSqlDataStore,
                                     handle interface{}) error {
    log := logging.GetAppLoggerObj()
    getPtr, err := sqlds.getDBGetFunction(handle)
    if err != nil {
        log.Error("Failed to get user entry %s, invalid DB handle err : %s",
                   user.userid, err)
        return err
    }
    if len(user.userid) == 0  || len(user.hashpwd) == 0{
        log.Info("Empty userID/Pwd cannot find in user table")
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_FOUND])
    }
    var row sqlDBUsers
    err = getPtr(&row, userGetonUserIDPwd, user.userid, user.hashpwd)
    if err != nil {
        log.Trace("Failed to read user record for userid %s, err : %s",
                        user.userid, err)
        return err
    }
    user.DBtoUserRowXlate(&row)
    return nil
}

func (user *sqlUsers)getUserwithID(sqlds *postgreSqlDataStore,
                                     handle interface{}) error {
    log := logging.GetAppLoggerObj()
    getPtr, err := sqlds.getDBGetFunction(handle)
    if err != nil {
        log.Error("Failed to get user entry %s, invalid DB handle err : %s",
                   user.userid, err)
        return err
    }
    if len(user.userid) == 0 {
        log.Info("Empty userID cannot find in user table")
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_FOUND])
    }
    var row sqlDBUsers
    err = getPtr(&row, userGetonUserID, user.userid)
    if err != nil {
        log.Trace("Failed to read user record for userid %s, err : %s",
                        user.userid, err)
        return err
    }
    user.DBtoUserRowXlate(&row)
    return nil
}

func (user *sqlUsers)createUserEntry(sqlds *postgreSqlDataStore,
                                     handle interface{}) error {
    log := logging.GetAppLoggerObj()
    execPtr, err := sqlds.getDBExecFunction(handle)
    if err != nil {
        log.Error("Failed to create user entry %s, invalid DB handle err : %s",
            user.userid, err)
        return err
    }
    if len(user.userid) == 0 || len(user.emailid) == 0 ||
        len(user.hashpwd) == 0 || len(user.mobileno) == 0 ||
        user.status == 0 {
            log.Error("Cannot create user record, invalid params")
            return fmt.Errorf("%s",
                    errorset.ERROR_TYPES[errorset.INVALID_PARAM])
    }
    //Check if record already present before trying to create an entry.
    err = user.getUserwithID(sqlds, handle)
    if err != nil && err != sql.ErrNoRows {
        //Something wrong to get the records, Dont let create the records.
        log.Info("Failed to get record on userid %s, cannot create",
                                        user.userid)
        return err
    }
    if err == nil {
        log.Info("%s user record already present in system", user.userid)
        return nil
    }
    // now we are at 'err == sql.ErrNoRows'
    user.startTime = time.Now()
    dbrow := user.usertoDBRowXlate()
    _, err = execPtr(userCreate, dbrow.Userid, dbrow.Emailid, dbrow.Hashpwd,
                    dbrow.Mobileno, dbrow.Dob, dbrow.StartTime,
                    dbrow.Validity, dbrow.Status)
    if err != nil {
        log.Error("Failed to create user record for %s err : %s", user.userid,
                    err)
        return err
    }
    return nil
}


// Function to delete user entry with userID
func (user *sqlUsers)deleteUserEntry(sqlds *postgreSqlDataStore,
                                     handle interface{}) error {
    log := logging.GetAppLoggerObj()
    execPtr, err := sqlds.getDBExecFunction(handle)
    if err != nil {
        log.Error("Failed to delete user entry %s, invalid DB handle err : %s",
                    user.userid, err)
        return err
    }
    _, err = execPtr(userDeleteOnID, user.userid)
    if err != nil {
        log.Info("Failed to delete the record %s , err : %s", user.userid,
                        err)
        return err
    }
    return nil
}

//Function to update user fields emailid, hashpwd, mobileno, validity and
// status.
//Its responsibility of caller to make sure populate all the fields in the
// user structure.
func (user *sqlUsers)updateUserEntry(sqlds *postgreSqlDataStore,
                                     handle interface{}) error {
    log := logging.GetAppLoggerObj()
    execPtr, err := sqlds.getDBExecFunction(handle)
    if err != nil {
        log.Error("Failed to update user entry %s, invalid DB handle err : %s",
                    user.userid, err)
        return err
    }
    if user.status == 0 {
        log.Info("Cannot update user record %s, invalid status", user.userid)
        return fmt.Errorf("%s",
                    errorset.ERROR_TYPES[errorset.INVALID_PARAM])
    }
    // Not validating if fields need an update really.
    dbrow := user.usertoDBRowXlate()
    _, err = execPtr(userUpdateOnID, dbrow.Emailid, dbrow.Hashpwd,
                    dbrow.Mobileno, dbrow.Status, dbrow.Validity)
    if err != nil {
        log.Info("Failed to update the user entry %s error : %s",
                dbrow.Userid, err)
        return err
    }
    return nil
}
