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

package errorset

import (

)

//******************************************************************************
// ALL ERRORS MUST BE PREDEFINED TO USE IN APPLICATION. NEVER CREATE ERROR
// VALUES ON THE FLY IN THE APPLICATION.
// THE INDEX VALUES SHOULD MATCH WITH THE ERROR STRING. UPDATE BOTH arrays at
// same time.
//******************************************************************************

const (
    INVALID_PARAM = iota
    TRY_AGAIN
    NULL_DB_CONFIG_PARAMS
    INVALID_DB_DRIVER
    INVALID_DB_CREDENTIALS
    DB_TABLE_CREATE_FAILED
    DB_TRANSACTION_FAILED
    DB_RECORD_NOT_FOUND
    DB_PARENT_RECORD_NOT_FOUND
    DB_RECORD_NOT_UNIQUE
    DB_RECORD_RELATION_ERROR
)

var ERROR_TYPES = []string{
    //INVALID_PARAM
    "Invalid input parameters",
    //TRY_AGAIN
    "Operation failed, Try again.",
    //NULL_DB_CONFIG_PARAMS
    "DB params in the configuration is empty/invalid",
    //INVALID_DB_DRIVER
    "INVALID DB driver in configuration",
    //INVALID_DB_CREDENTIALS
    "Invalid DB credentails in the configuration",
    //DB_TABLE_CREATE_FAILED
    "Failed to create table in DB",
    //DB_TRANSACTION_FAILED
    "Failed to perform transaction on DB",
    //DB_RECORD_NOT_FOUND
    "Record not found in DB",
    //DB_PARENT_RECORD_NOT_FOUND
    "Parent record is not found in DB",
    //DB_RECORD_NOT_UNIQUE
    "More than one record found in DB",
    //DB_RECORD_RELATION_ERROR
    "Error in DB record relation/no valid relation found"}