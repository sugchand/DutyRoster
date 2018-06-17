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
    UNITADMIN rolebit = 1 << iota
    DIVADMIN
    BRANCHADMIN
    REGIONADMIN
    ORGADMIN
    // The application admin, who has access to all the datasets.
    ROOTADMIN
)

//User role in the application. User can have any role in the above list.
//It is possible to one user may have more than one role.
type roles struct {
    roleType uint64
}

func(role *roles)CreateRole() {
    
}

func (role *roles)DeleteRole() {
    
}

func (role *roles)GetRole() {
    
}

func (role *roles)CreateRoleTable() {
    
}