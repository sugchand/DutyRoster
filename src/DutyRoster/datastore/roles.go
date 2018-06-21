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

// Each bit is reserved for a role type, its is only possible to have hierarchy
// of maximum 64 levels as the roletype is 64 bit integer.
// Adding role to this bitmap should be in hierarchical order as no role should
// be placed after ROOTADMIN.
type rolebit uint64

const (
    ENDUSER rolebit = 1 << iota
    MANAGER rolebit = 1 << iota
    // The application admin, who has access to all the datasets.
    ROOTADMIN
)

//User role in the application. User can have any role in the above list.
//It is possible to one user may have more than one role.
type roles struct {
    roleType rolebit
}

// Validate the rolebitset is valid.
// Return true for a valid rolebitset and false otherwise.
func (rl *roles)IsRoleBitsetValid() bool{
    //Assuming there are no role bit present after rootadmin.
    var maxRole rolebit = (ROOTADMIN << 1) - 1 //All 0xFs.
    var minRole rolebit = ENDUSER
    if rl.roleType < minRole || rl.roleType > maxRole {
        return false
    }
    return true
}