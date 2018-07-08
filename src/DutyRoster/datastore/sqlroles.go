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
    _ "github.com/lib/pq"
    "DutyRoster/errorset"
    "DutyRoster/logging"
)

type sqlroles struct {
    roles
}

//String representation of columns/DB names. Should be in align with roles struct
const (
    ROLE_TABLE_NAME_STR = "roles"
    ROLE_TYPE_NAME_STR = "roleType"
)

// All the sql statements to run on role table.
var (
    //Check if table is exist in the DB
    roletableExist = fmt.Sprintf("SELECT 1 FROM %s LIMIT 1;", ROLE_TABLE_NAME_STR)
    //Create a table roles
    roleschema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
                            (%s bigint NOT NULL PRIMARY KEY CHECK(%s > 0));`,
                    ROLE_TABLE_NAME_STR, ROLE_TYPE_NAME_STR,
                    ROLE_TYPE_NAME_STR)
    //Create a role entry in table roles
    roleCreate = fmt.Sprintf("INSERT INTO %s (%s) VALUES ($1)", ROLE_TABLE_NAME_STR,
                            ROLE_TYPE_NAME_STR)
    //Get the roles row for specific roletype
    roleGet = fmt.Sprintf("SELECT * FROM %s WHERE %s=($1)", ROLE_TABLE_NAME_STR,
                        ROLE_TYPE_NAME_STR)
    //Get total number of role entries in table.
    roleGetNum = fmt.Sprintf("SELECT COUNT(*) FROM %s", ROLE_TABLE_NAME_STR)
    //Delete role entry in table roles.
    roleDelete = fmt.Sprintf("DELETE FROM %s WHERE %s=($1)", ROLE_TABLE_NAME_STR,
                        ROLE_TYPE_NAME_STR)
)

func (rl *sqlroles)createRoleTable(sqlds *postgreSqlDataStore,
                                     handle interface{}) error{
    log := logging.GetAppLoggerObj()
    execPtr, err := sqlds.getDBExecFunction(handle)
    if err != nil {
        log.Error("Failed to create role table, invalid DB handle err : %s",
                    err)
        return err
    }
    _, err = execPtr(roletableExist)
    if err == nil {
        log.Info("Role table is already exist in the system. ")
        return nil
    }
    _, err = execPtr(roleschema)
    if err != nil {
        log.Error("Failed to create role table: %s", err)
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.DB_TABLE_CREATE_FAILED])
    }
    return nil
}

//Function to create a role entry in table if not exist.
func (rl *sqlroles)createRoleEntry(sqlds *postgreSqlDataStore,
                                     handle interface{}) error {
    log := logging.GetAppLoggerObj()
    execPtr, err := sqlds.getDBExecFunction(handle)
    if err != nil {
        log.Error("Failed to create role entry %d, invalid DB handle err : %s",
            rl.roleType, err)
        return err
    }
    //Check if roleBit is valid before creating it in DB
    if rl.IsRoleBitsetValid() == false {
        log.Error("Invalid role bit, cannot create entry in role table")
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.INVALID_PARAM])
    }
    //role table has only one entry and its the primary key, duplication of
    // entry will cause error in DB insert. We are checking if entry is present
    // in DB to avoid doing trial and error.
    if rl.isRoleEntryPresentInTable(sqlds, handle) == true {
        //Role entry is present, no need to create
        log.Trace("No need to create role entry %d as its present in DB",
                    rl.roleType)
        return nil
    }
    _, err = execPtr(roleCreate, rl.roleType)
    if err != nil {
        log.Trace("Failed to create role table entry %d", rl.roleType)
        return err
    }
    return nil
}

//Return TRUE if a role entry present in table and false otherwise.
func (rl *sqlroles)isRoleEntryPresentInTable(sqlds *postgreSqlDataStore,
                                     handle interface{}) (bool) {
    log := logging.GetAppLoggerObj()
    getPtr, err := sqlds.getDBGetFunction(handle)
    if err != nil {
        log.Error("invalid DB handle to work on role entry %d err : %s",
                   rl.roleType, err)
        return false
    }
    var row uint64
    err = getPtr(&row, roleGet, rl.roleType)
    if err != nil {
        log.Trace("Failed to get the role entry %d from role table",
                    rl.roleType)
        return false
    }
    return true
}

//Function to get total number of entries that present in role table.
func (rl *sqlroles)getTotalRoleEntriesCnt(sqlds *postgreSqlDataStore,
                                     handle interface{}) (uint64, error) {
    log := logging.GetAppLoggerObj()
    var totCnt uint64
    getPtr, err := sqlds.getDBGetFunction(handle)
    if err != nil {
        log.Error("invalid DB handle to work on role entry %d err : %s",
                   rl.roleType, err)
        return 0, err
    }
    err = getPtr(&totCnt, roleGetNum)
    if err != nil {
        log.Error("Failed to get total number of records in roletable")
        return 0, err
    }
    return totCnt, nil
}

//function to delete a roleentry from the role table.
func (rl *sqlroles)delRoleEntry(sqlds *postgreSqlDataStore,
                                     handle interface{}) error {
    log := logging.GetAppLoggerObj()
    execPtr, err := sqlds.getDBExecFunction(handle)
    if err != nil {
        log.Error("Failed to create role table, invalid DB handle err : %s",
                    err)
        return err
    }
    _, err = execPtr(roleDelete, rl.roleType)
    if err != nil {
        log.Trace("Failed to delete the role entry from table")
        return err
    }
    return nil
}
