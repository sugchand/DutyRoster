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
    "DutyRoster/syncParam"
)

type orgStatusBit uint64

const (
    ORG_REQUESTED orgStatusBit = 1 << iota
    ORG_APPROVED orgStatusBit = 1 << iota
    //Last entry in the org status. Do not add anything below the delete status.
    ORG_DELETED orgStatusBit = 1 << iota
)

type org struct {
    // A unique ID assigned to an organization or division in organization.
    uuid syncParam.UUID
    //name of organization or division in organization.
    name string
    address string
    //parent to track the hierarchy in the organization. It is possible an org
    // can have various levels in a hierarchy. The organization will have depth
    // 0, and divisions in the org might get numbers assigned from 1,2,3 and so
    // on.
    parent *org
    //A new org will have a status requested/approved.
    // Creating a new org will having a state requested/approved or both.
    status orgStatusBit
    //timstamp when a org is created.
    startTime time.Time
    //validity of organization in days in the application.
    //Store 0 for unlimited validity.
    validity uint64
}

// Validate the rolebitset is valid.
// Return true for a valid rolebitset and false otherwise.
func (or *org)IsOrgStatusValid() bool{
    //Assuming there are no role bit present after rootadmin.
    var maxOrgBit orgStatusBit = (ORG_DELETED << 1) - 1 //All 0xFs.
    var minOrgBit orgStatusBit = ORG_REQUESTED
    if or.status < minOrgBit || or.status > maxOrgBit {
        return false
    }
    return true
}