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
    "time"
)

type userStatusBit uint64
const (
    USER_REQUESTED userStatusBit = 1 << iota
    USER_APPROVED userStatusBit = 1 << iota
    //Last entry in the org status. Do not add anything below the delete status.
    USER_DELETED userStatusBit = 1 << iota
)

type Users struct {
    userid string
    emailid string
    hashpwd string
    mobileno string
    //store in bitfields as DD|MM|YYYY
    dob time.Time
    //Time when user record is being created
    startTime time.Time
    //validity of userrecord, Needed for bookkeeping.
    validity uint64
    //Status of user record.
    status userStatusBit
}

//Structure to track link between user, roles and Org.
// Each user entry will have specific role in every org/unit
type userOrgRole struct {
    //Anonymous User account
    *Users
    *roles
    *org
}

