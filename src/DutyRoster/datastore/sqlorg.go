package datastore

import (
    "fmt"
    "time"
    _ "github.com/mattn/go-sqlite3"
    "github.com/jmoiron/sqlx"
    "DutyRoster/errorset"
    "DutyRoster/logging"
    "DutyRoster/syncParam"
)

//The db representation of org table
type dbOrg struct {
    Uuid string `db:"uuid"`
    Name string `db:"name"`
    Parent string `db:"parent"`//uuid string of parent
    Status uint64 `db:"status"`
    StartTime int64 `db:"startTime"`
    Validity uint64 `db:"validity"`
}

type sqlorg struct {
    org
    // Database row structure.
    dbOrg *dbOrg
}

//String representation of Org table and its elements.
//Update the string reperesentation when make any change to the org/sqlorg struct.
const (
    ORG_TABLE_NAME = "Org"
    ORG_FIELD_UUIID = "uuid"
    ORG_FIELD_NAME = "name"
    ORG_FIELD_PARENT = "parent"
    ORG_FIELD_STATUS = "status"
    ORG_FIELD_START_TIME = "startTime"
    ORG_FIELD_VALIDITY = "validity"
)

// SQL statements to be used to operate on org table.
var (
    //Check if table is exist in the DB
    orgtableExist = fmt.Sprintf("SELECT 1 FROM %s LIMIT 1;",
                            ORG_TABLE_NAME)
    //Create a table org
    orgschema = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
                    (%s TEXT NOT NULL PRIMARY KEY,
                     %s TEXT NOT NULL UNIQUE,
                     %s  TEXT,
                     %s UNSIGNED BIG INT NOT NULL,
                     %s BIG INT NOT NULL,
                     %s UNSIGNED INT NOT NULL,
                     FOREIGN KEY(%s) REFERENCES %s(%s));`,
                     ORG_TABLE_NAME,
                     ORG_FIELD_UUIID,
                     ORG_FIELD_NAME,
                     ORG_FIELD_PARENT, //Linked to parent ORG_FIELD_UUIID
                     ORG_FIELD_STATUS,
                     ORG_FIELD_START_TIME,
                     ORG_FIELD_VALIDITY,
                     ORG_FIELD_PARENT, ORG_TABLE_NAME, ORG_FIELD_UUIID,)
    //Create a org entry in table Org
    orgCreate = fmt.Sprintf(`INSERT INTO %s 
                            (%s, %s, %s, %s, %s, %s) 
                            VALUES (?, ?, ?, ?, ?, ?)`,
                            ORG_TABLE_NAME,
                            ORG_FIELD_UUIID, ORG_FIELD_NAME, ORG_FIELD_PARENT,
                            ORG_FIELD_STATUS, ORG_FIELD_START_TIME,
                            ORG_FIELD_VALIDITY)
    //Get number of org/unit with name. This should be either 1, 0 as name is
    //unique
    orgGetNameCnt = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s=(?)`,
                                ORG_TABLE_NAME, ORG_FIELD_NAME)
    //Get the org/Unit rows with specific uuid
    orgGetUUID = fmt.Sprintf(`SELECT * FROM %s WHERE %s=(?)`,
                              ORG_TABLE_NAME, ORG_FIELD_UUIID)
    //Get the org/unit rows with specific org name. 
    orgGetName = fmt.Sprintf(`SELECT * FROM %s WHERE %s=(?)`,
                             ORG_TABLE_NAME, ORG_FIELD_NAME)
    //Get total number of org/unit entries in table.
    orgGetTotNum = fmt.Sprintf("SELECT COUNT(*) FROM %s", ORG_TABLE_NAME)
    //Delete org entry in table org table
    orgDeleteName = fmt.Sprintf("DELETE FROM %s WHERE %s=(?)", ORG_TABLE_NAME,
                                ORG_FIELD_NAME))

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
    log := logging.GetAppLoggerObj()

    if org.IsOrgStatusValid() == false {
        log.Trace("Organization %s doesnt have a proper status")
        return fmt.Errorf("%s",
                          errorset.ERROR_TYPES[errorset.INVALID_PARAM])
    }
    //Check if org entry already present in system.
    if org.isOrgEntryPresentInTable(conn) == true {
        log.Trace("Organization %s already present in system, cannot create",
            org.name)
        return nil
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
    _, err = conn.Exec(orgCreate, dbrow.Uuid, dbrow.Name, dbrow.Parent,
                       dbrow.Status, dbrow.StartTime, dbrow.Validity)
    if err != nil {
        log.Trace("Failed to create a org record for %s", dbrow.Name)
         return err
    }
    return nil
}

//Return TRUE if a Org already present in table and false otherwise.
func (org *sqlorg)isOrgEntryPresentInTable(conn *sqlx.DB) (bool) {
    var err error
    log := logging.GetAppLoggerObj()

    var res uint64
    err = conn.Get(&res, orgGetNameCnt, org.name)
    if err != nil || res == 0 {
        log.Trace("Failed to get/ the org entry %s not present in Org table",
                   org.name)
        return false
    }
    return true
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
    dbrow.Uuid = string(org.uuid)
    dbrow.Name = org.name
    dbrow.Parent = ""
    if org.parent != nil {
        //Update the parent with parent UUID
        dbrow.Parent = string(org.parent.uuid) 
    }
    // Convert the time to unix format.
    dbrow.StartTime = org.startTime.Unix()
    dbrow.Status = uint64(org.status)
    dbrow.Validity = org.validity
    return dbrow
} 
// Translate DB org row to a Org structure.
//It can be a recursive call to fill the parent fields accordigly
func (org *sqlorg)dbToOrgRowXlate(conn *sqlx.DB, dbrow *dbOrg) {
    org.name = dbrow.Name
    org.uuid = syncParam.UUID(dbrow.Uuid)
    org.status = orgStatusBit(dbrow.Status)
    org.validity = dbrow.Validity
    org.startTime = time.Unix(dbrow.StartTime, 0)
    if len(dbrow.Parent) == 0 {
        // No parent present, return it now.
        return
    }
    //Recursively process the parent until we reach global parent.
    var parentOrg = new(sqlorg)
    parentOrg.uuid = syncParam.UUID(dbrow.Parent)
    org.parent = &parentOrg.org
    org.dbOrg = nil //No need to populate DB row fields
    parentOrg.getOrgEntryByUUID(conn)
}

//Function to get Org entry with specific Name
// The values are written back to the org Obj itself.
func (org *sqlorg)getOrgEntryByName(conn *sqlx.DB) error{
    var err error
    log := logging.GetAppLoggerObj()
    if len(org.name) == 0 {
        log.Trace("Empty org name, cannot find in org table.")
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_FOUND])
    } 
    var row dbOrg
    err = conn.Get(&row, orgGetName, org.name)
    if err != nil {
        log.Trace("Failed to read org record for %s", org.name)
        return err
    }
    org.dbToOrgRowXlate(conn, &row)
    return nil
}

//Function to get Org entry with specific UUID
// The values are written to the orgObj itself.
func (org *sqlorg)getOrgEntryByUUID(conn *sqlx.DB) error{
    var err error
    log := logging.GetAppLoggerObj()
    if len(org.uuid) == 0 {
        log.Trace("Empty org UUID, cannot find in org table.")
        return fmt.Errorf("%s",
                        errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_FOUND])
    } 
    var row dbOrg
    err = conn.Get(&row, orgGetUUID, org.uuid)
    if err != nil {
        log.Trace("Failed to read org record for uuid %s", org.uuid)
        return err
    }
    org.dbToOrgRowXlate(conn, &row)
    return nil
}

//Function to delete a org hiearchy in DB.
// The orgname/unit name should be provided to delete a org entry from table.
func (org *sqlorg)deleteOrgEntryByName(conn *sqlx.DB) error {
    var err error
    log := logging.GetAppLoggerObj()
    if len(org.name) == 0 {
        log.Trace("Invalid/Null org record name, cannot delete from org table")
        return fmt.Errorf("%s",
                            errorset.ERROR_TYPES[errorset.DB_RECORD_NOT_FOUND])
    }
    _, err = conn.Exec(orgDeleteName, org.name)
    if err != nil {
        log.Trace("Failed to delete record %s from org table", org.name)
        return err
    }
    if org.parent == nil {
        //Global org entry have nil parent
        return nil
    }
    //Tail unroll to delete parents as well.
    var parentorg = new(sqlorg)
    //Need only the name to delete parent.
    parentorg.name = org.parent.name
    return parentorg.deleteOrgEntryByName(conn)
}