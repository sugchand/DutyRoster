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

)

//Datastore Interface that provides the APIs exposed by datastore implementation.
type dataStoreInterface interface {
    CreateDBConnection() error
    // Create all the relevant tables/sessions that are needed for datastore
    //implementation.
    CreateDataStoreTables() error

    //***** User operations *****
    //Create the user account row in the DB.
    CreateUserAccount(*Users) error
    //Get a user account, the userID and pwd must be present in users
    // All other fields are populated by the function by reading from DB.
    GetUserAccount(*Users) error
    //Delete User account with 'userid' row in the DB,
    DeleteUserAccount(*Users) error
    //Update User account on 'Userid'.
    //Only emailid, hashpwd, mobileno, validity and status are allowed to
    //modify. All these fields must populate in the 'users' even if
    // update is not required.Otherwise the null values get written to DB.
    UpdateUserAccount(*Users) error
}