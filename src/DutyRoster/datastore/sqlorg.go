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
    "github.com/jmoiron/sqlx"
    "DutyRoster/errorset"
    "DutyRoster/logging"
    "DutyRoster/syncParam"
)

//The db representation of org table. Used only for SQLX operations.
// It is not possible to interact with SQL with default 'go' specific datatypes.
//The following structure has a direct 1:1 mapping to 'org' structure.
//By having standard 'org' structure the system offers an abstraction for backend
//implementations. In SQL DB backend implementation, the org structure is mapped
//to sql model 'dborg'. All the 'org' specific methods are extended in sqlorg.
type dbOrg struct {
    Uuid string `db:"uuid"`
    Name string `db:"name"`
    Address sql.NullString `db:"address"`
    Parent sql.NullString `db:"parent"`//uuid string of parent
    Status uint64 `db:"status"`
    StartTime time.Time `db:"starttime"` //In date type.
    Validity sql.NullInt64 `db:"validity"` //Validity in days.
}

// SQL representation for Org.
type sqlorg struct {
    org
}

//String representation of Org table and its elements.
//Update the string reperesentation when make any change to the org/sqlorg struct.
const (
    ORG_NAME_STR_LEN = 500
    ORG_TABLE_NAME = "Org"
    ORG_FIELD_UUID = "uuid"
    ORG_FIELD_NAME = "name"
    ORG_FIELD_ADDRESS = "address"
    ORG_FIELD_PARENT = "parent"
    ORG_FIELD_STATUS = "status"
    ORG_FIELD_START_TIME = "starttime"
    ORG_FIELD_VALIDITY = "validity"
)

// SQL statements to be used to operate on org table.
var (
    //Check if table is exist in the DB
    orgtableExist = fmt.Sprintf("SELECT 1 FROM %s LIMIT 1;",
                            ORG_TABLE_NAME)
    //Create a table org
    orgschema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
                    (%s UUID NOT NULL PRIMARY KEY,
                     %s varchar(%d) NOT NULL,
                     %s varchar(%d),
                     %s UUID NULL REFERENCES %s(%s) ON DELETE SET NULL
                     ON UPDATE SET NULL,
                     %s bigint NOT NULL CHECK(%s > 0),
                     %s timestamp NOT NULL,
                     %s bigint NULL);`,
                     ORG_TABLE_NAME,
                     ORG_FIELD_UUID,
                     ORG_FIELD_NAME, ORG_NAME_STR_LEN,
                     ORG_FIELD_ADDRESS, ORG_NAME_STR_LEN,
                     ORG_FIELD_PARENT, //Linked to parent ORG_FIELD_UUID
                     ORG_TABLE_NAME, ORG_FIELD_UUID,
                     ORG_FIELD_STATUS, ORG_FIELD_STATUS,
                     ORG_FIELD_START_TIME,
                     ORG_FIELD_VALIDITY)
    //Create a org entry in table Org
    orgCreate = fmt.Sprintf(`INSERT INTO %s
                            (%s, %s, %s, %s, %s, %s, %s)
                            VALUES ($1, $2, $3, $4, $5, $6, $7)`,
                            ORG_TABLE_NAME,
                            ORG_FIELD_UUID, ORG_FIELD_NAME, ORG_FIELD_ADDRESS,
                            ORG_FIELD_PARENT,ORG_FIELD_STATUS,
                            ORG_FIELD_START_TIME, ORG_FIELD_VALIDITY)
    //Get number of org/unit with name. This should be either 1, 0 as name is
    //unique
    orgGetNameCnt = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s=($1)`,
                                ORG_TABLE_NAME, ORG_FIELD_NAME)
    //Get the org/Unit rows with specific uuid
    orgGetonUUID = fmt.Sprintf(`SELECT * FROM %s WHERE %s=($1)`,
                              ORG_TABLE_NAME, ORG_FIELD_UUID)
    //Get the org/unit using name, addr and parent UUID
    orgGetonNameAddrParent = fmt.Sprintf(`SELECT * FROM %s WHERE %s=($1) AND
                                (%s=($2) OR %s IS NULL) AND
                                (%s=($3) OR %s IS NULL)`,
                                ORG_TABLE_NAME,
                                ORG_FIELD_NAME,
                                ORG_FIELD_ADDRESS, ORG_FIELD_ADDRESS,
                                ORG_FIELD_PARENT, ORG_FIELD_PARENT)
    //Get org/unit rows with specific parent.
    orgGetonParent = fmt.Sprintf("SELECT * FROM %s WHERE %s=($1)",
                                ORG_TABLE_NAME, ORG_FIELD_PARENT)
    //Get total number of org/unit entries in table.
    orgGetTotNum = fmt.Sprintf("SELECT COUNT(*) FROM %s", ORG_TABLE_NAME)
    //Delete org entry in table org table
    orgDelete = fmt.Sprintf("DELETE FROM %s WHERE %s=($1)",
                                ORG_TABLE_NAME, ORG_FIELD_UUID)
    //Update org status and validitiy fields on a Name match.
    orgUpdate = fmt.Sprintf(`UPDATE %s SET %s=($1), %s=($2)
                             WHERE %s=($3)`,
                            ORG_TABLE_NAME, ORG_FIELD_STATUS,
                            ORG_FIELD_VALIDITY, ORG_FIELD_UUID))

func (org *sqlorg)createOrgTable(conn *sqlx.DB) error{
    var err error
    log := logging.GetAppLoggerObj()
    _, err = conn.Exec(orgtableExist)
    if err == nil {
        log.Info("Org table is already exist in the system. ")
        return nil
    }
    _, err = conn.Exec(orgschema)
    if err != nil {
        log.Error("Failed to create org table %s", err)
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.DB_TABLE_CREATE_FAILED])
    }
    return nil
}

//Create a new org/unit entry in org table using the sqlorg structure
// uuid, startTime will be self populated.
func (org *sqlorg)createOrgEntry(conn *sqlx.DB) error {
    var err error
    var res bool
    log := logging.GetAppLoggerObj()

    if org.IsOrgStatusValid() == false {
        log.Trace("Organization %s doesnt have a proper status")
        return fmt.Errorf("%s",
                          errorset.ERROR_TYPES[errorset.INVALID_PARAM])
    }
    //Populate UUID for all the ancestors for the record
    err = org.fillUUIDforOrgParents(conn, &org.org)
    if err != nil{
        log.Info("Cannot create a org entry as failed to find ancestors")
        return err
    }
    //Check if org entry already present in system.
    res, err = org.isOrgEntryPresentInTable(conn)
    if res == true {
        log.Trace("Organization %s already present in system, cannot create",
            org.name)
        return nil
    }
    if err != nil {
        //Cannot create the org entry as error while finding the entry
        return err
    }
    org.uuid,err = syncParam.NewUUID()
    if err != nil {
        log.Trace("Failed to create UUID, cannot create org table entry %s",
                   org.name)
        return fmt.Errorf("%s",
                          errorset.ERROR_TYPES[errorset.TRY_AGAIN])
    }
    org.startTime = time.Now()
    dbrow := org.orgToDBRowXlate()
    parent_uuid, _ := dbrow.Parent.Value()
    _, err = conn.Exec(orgCreate, dbrow.Uuid, dbrow.Name, dbrow.Address,
                   parent_uuid, dbrow.Status, dbrow.StartTime, dbrow.Validity)
    if err != nil {
        log.Trace("Failed to create a org record for %s : %s", dbrow.Name, err)
         return err
    }
    return nil
}

// Function to find and fill the UUID for specific Org entry. An org entry
// may have only provided with name , address and parent.Find and fill the UUID
// for specific org entry and all its parents.
func (org *sqlorg)fillUUIDforOrgParents(conn *sqlx.DB, orgentry *org) error{
    var err error
    log := logging.GetAppLoggerObj()
    if orgentry.parent != nil {
        //Need to start the processing from the root parent.
        org.fillUUIDforOrgParents(conn, orgentry.parent)
    }
    if !syncParam.IsUUIDEmpty(orgentry.uuid) {
        //UUID is present and no need to calculate.
        return nil
    }
    //Find UUID using name, address and parent UUID.
    var orgwrapper *sqlorg
    orgwrapper = new(sqlorg)
    orgwrapper.org = *orgentry
    dbrow := orgwrapper.orgToDBRowXlate()
    rows := []dbOrg{}
    err = conn.Select(&rows, orgGetonNameAddrParent, dbrow.Name, dbrow.Address,
                            dbrow.Parent)
    if err != nil {
        log.Trace("Failed to get a DB entry using %s, %s : %s",
                    dbrow.Name, dbrow.Address, err)
        return err
    }
    rowlen := len(rows)
    if rowlen > 1 {
        //Something is wrong, Its not possible to have more than one record that
        // has same name and address under same org parent.
        return fmt.Errorf("%s",
                errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_UNIQUE])
    } else if rowlen == 1 {
        orgentry.uuid = syncParam.StringtoUUID(rows[0].Uuid)
    }
    return nil
}


//Return TRUE if a Org already present in table and false otherwise.
//Check if the org hierarchy is present in the
func (org *sqlorg)isOrgEntryPresentInTable(conn *sqlx.DB) (bool, error) {
    var err error
    log := logging.GetAppLoggerObj()
    if org.parent != nil {
        parentorg := new(sqlorg)
        parentorg.org = *org.parent
        res, _ := parentorg.isOrgEntryPresentInTable(conn)
        if res == false {
            //Cannot find the parent of the org record, return error
            return false, fmt.Errorf("%s",
                      errorset.ERROR_TYPES[errorset.DB_PARENT_RECORD_NOT_FOUND])
        }
    }
    if syncParam.IsUUIDEmpty(org.uuid) {
        //UUID is empty, cannot find a org entry with empty UUID
        log.Trace("Empty UUID for the org record : %s-%s",
                        org.name, org.address)
        err = org.getOrgEntryByNameAddrParent(conn)
        if err != nil && err == sql.ErrNoRows {
            //No rows present in the DB, no need to return any error code
            return false, nil
        }
        return false, fmt.Errorf("%s",
                    errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_FOUND])
    }
    var dbrow dbOrg
    err = conn.Get(&dbrow, orgGetonUUID, syncParam.UUIDtoString(org.uuid))
    if err != nil {
        //Check if error is for no row found.
        if err == sql.ErrNoRows {
            //We dont find the row, so lets return accordingly.
            return false, nil
        }
        return false, err
    }
    return true, nil
}

//Function to get total number of entries that present in Org table.
func (org *sqlorg)getTotalOrgEntriesCnt(conn *sqlx.DB) (uint64, error) {
    var err error
    log := logging.GetAppLoggerObj()
    var totCnt uint64
    err = conn.Get(&totCnt, orgGetTotNum)
    if err != nil {
        log.Error("Failed to get total number of records in Org Table")
        return 0, err
    }
    return totCnt, nil
}

// Translate Org to DB row in table.
// the org table will be mapped to corresponding db row to insert into sql DB.
func (org *sqlorg)orgToDBRowXlate() *dbOrg {
    var dbrow *dbOrg
    dbrow = new(dbOrg)
    //UUID cannot be null, so no need to check if its null.
    dbrow.Uuid = syncParam.UUIDtoString(org.uuid)
    dbrow.Name = org.name
    dbrow.Address.Scan(org.address)
    dbrow.Parent.Scan(nil)
    if org.parent != nil {
        //Update the parent with parent UUID
        dbrow.Parent.Scan(syncParam.UUIDtoString(org.parent.uuid))
    }
    // Convert the time to unix format.
    dbrow.StartTime = org.startTime
    dbrow.Status = uint64(org.status)
    dbrow.Validity.Scan(org.validity)
    return dbrow
}
// Translate DB org row to a Org structure.
//It can be a recursive call to fill the parent fields accordigly
func (org *sqlorg)dbToOrgRowXlate(conn *sqlx.DB, dbrow *dbOrg) {
    var ret bool
    org.name = dbrow.Name

    org_address, _ := dbrow.Address.Value()
    //Check the type of value to avoid runtime panic on invalid datatype.
    if org.address, ret = org_address.(string); !ret {
        org.address = ""
    }
    org.uuid = syncParam.StringtoUUID(dbrow.Uuid)
    org.status = orgStatusBit(dbrow.Status)
    org_validity, _ := dbrow.Validity.Value()
    if org.validity, ret = org_validity.(uint64); !ret {
        org.validity = 0
    }
    org.startTime = dbrow.StartTime
    org_parent, _ := dbrow.Parent.Value()
    var org_parentStr string
    if org_parentStr, ret = org_parent.(string); !ret {
        // Nil value, so set it to empty string.
        org_parentStr = ""
    }
    //Recursively process the parent until we reach global parent.
    var parentOrg = new(sqlorg)
    parentOrg.uuid = syncParam.StringtoUUID(org_parentStr)
    org.parent = &parentOrg.org
    parentOrg.getOrgEntryByUUID(conn)
}

//Get the org entry with specific name, address and parent.
//It is only possible to have one entry with name, address and parent.
func (org *sqlorg)getOrgEntryByNameAddrParent(conn *sqlx.DB) error{
    var err error
    log := logging.GetAppLoggerObj()
    dbrow := org.orgToDBRowXlate()
    rows := []dbOrg{}
    err = conn.Select(&rows, orgGetonNameAddrParent, dbrow.Name, dbrow.Address,
                            dbrow.Parent)
    if err != nil {
        //Failed to get the rows.
        log.Trace("Failed to get record with name %s , address %s, parent %s",
            dbrow.Name, dbrow.Address, dbrow.Parent)
        return err
    }
    rowlen := len(rows)
    if  rowlen > 1 {
        log.Info("Multiple record present in DB with same name %s, address %s",
                    dbrow.Name, dbrow.Address)
        return fmt.Errorf("%s",
            errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_UNIQUE])
    } else if rowlen == 1 {
        //Expect only one row in DB.
        org.dbToOrgRowXlate(conn, &rows[0])
        return nil
    }
    return sql.ErrNoRows
}

//Function to get Org entry with specific UUID
// The values are written to the orgObj itself.
func (org *sqlorg)getOrgEntryByUUID(conn *sqlx.DB) error{
    var err error
    log := logging.GetAppLoggerObj()
    if syncParam.IsUUIDEmpty(org.uuid) {
        log.Trace("Empty org UUID, cannot find in org table %s.", org.name)
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_FOUND])
    }
    var row dbOrg
    err = conn.Get(&row, orgGetonUUID, syncParam.UUIDtoString(org.uuid))
    if err != nil {
        log.Trace("Failed to read org record for uuid %s", org.uuid)
        return err
    }
    org.dbToOrgRowXlate(conn, &row)
    return nil
}

//Function to check if org is differnt than the record in DB(dbrowOrg).
//Validity and status are allowed to change, hence need to validate only them.
//Return true if values are not equal false otherwise
func (org *sqlorg)isOrgNeedUpdate(dbrowOrg *sqlorg) bool{
    if org.validity != dbrowOrg.validity ||
        org.status != dbrowOrg.status {
            return true
    }
    return false
}

//Function to update a org entry and its children.
//Update is very expensive operation as finding child involves DB lookup.
//Only status and validity fields are allowed to update in org entry.
//To modify any other fields, delete and readd orgentry.
func (org *sqlorg)updateOrgEntry(conn *sqlx.DB) error {
    var err error
    log := logging.GetAppLoggerObj()
    //check if record present before any operation.
    var orgrow = new(sqlorg)
    *orgrow = *org

    if org.IsOrgStatusValid() == false {
        log.Info(`Cannot update the org record %s as invalid
                status bit provided`, org.name)
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.INVALID_PARAM])
    }

    if syncParam.IsUUIDEmpty(orgrow.uuid) {
        if err = orgrow.getOrgEntryByNameAddrParent(conn); err != nil {
            log.Info(`Failed to update record %s, error in finding record: %s`,
                    org.name, err)
            return err
        }
    } else {
        //Get the record with UUID
        if err = orgrow.getOrgEntryByUUID(conn); err != nil {
            log.Info(`Failed to update record %s, error in finding record using
                uuid : %s`, org.name, err)
            return err
        }
    }
    //orgrow will have the DB record, and org will have user input data
    if !org.isOrgNeedUpdate(orgrow) {
        log.Trace("No need to update org record %s as its same as DB record",
                    org.name)
        return nil
    }

    newstatus := org.status
    var newvalidity sql.NullInt64
    newvalidity.Scan(org.validity)

    //Workaround declaration, to call closures recursively.
    var updateFunc func(*sqlorg, orgStatusBit, sql.NullInt64)(error)
    //Update the children records as well before updating by itself.
    updateFunc = func(orgrow *sqlorg, newstatus orgStatusBit,
        newvalidity sql.NullInt64) error {
        rows := []dbOrg{}
        err = conn.Select(&rows, orgGetonParent,
                syncParam.UUIDtoString(orgrow.uuid))
        if err != nil {
            log.Info(`Failed to get the children rows of %s,
                  aborting the update, err : %s`, orgrow.name, err)
            return err
        }
        for _,row := range(rows) {
            childrow := new(sqlorg)
            childrow.dbToOrgRowXlate(conn, &row)
            err = updateFunc(childrow, newstatus, newvalidity)
            if err != nil {
                // XXX :: Chance of partial update ??
                log.Info("Failed to update children org records of %s, error %s",
                    row.Name, err)
                return err
            }
        }
        _, err = conn.Exec(orgUpdate, newstatus, newvalidity,
                        syncParam.UUIDtoString(orgrow.uuid))
        if err != nil {
            log.Info("Failed to update org table row %s err : %s",
                    orgrow.name, err)
            return err
        }
        return nil
    }
    return updateFunc(orgrow, newstatus, newvalidity)
}

//Function to delete a org hiearchy in DB.
// The orgname/unit name should be provided to delete a org entry from table.
func (org *sqlorg)deleteOrgEntry(conn *sqlx.DB) error {
    var err error
    log := logging.GetAppLoggerObj()
    if syncParam.IsUUIDEmpty(org.uuid) {
        //Find the entry using name address and parent.
        err = org.getOrgEntryByNameAddrParent(conn)
    } else {
        err = org.getOrgEntryByUUID(conn)
    }
    if err != nil {
        log.Info("Failed to delete a org entry, as cannot get org entry %s",
                    err)
        return err
    }
    if len(org.name) == 0 {
        log.Trace("Invalid/Null org record name, cannot delete from org table")
        return fmt.Errorf("%s",
                            errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_FOUND])
    }
    // Delete all the children org entries first before deleting the original
    rows := []dbOrg{}
    err = conn.Select(&rows, orgGetonParent, syncParam.UUIDtoString(org.uuid))
    if err != nil {
        //Something went wrong, lets not try to delete the records.
        log.Trace("Failed to get children records for org %s", org.name)
        return err
    }
    //Children records are present for specific org entry
    //Delete them first to avoid inconsistent records in the system.
    for _, row := range(rows) {
        var child = new(sqlorg)
        child.dbToOrgRowXlate(conn, &row)
        err = child.deleteOrgEntry(conn)
        if err != nil {
            // XXX:: Chance of partial delete.
            log.Info("Failed to delete child org record %s, err: %s", row.Name,
                        err)
            return err
        }
    }
    log.Trace("Deleting org entry %s", org.name)
    _, err = conn.Exec(orgDelete, syncParam.UUIDtoString(org.uuid))
    return nil
}
