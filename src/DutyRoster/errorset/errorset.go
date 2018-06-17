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
    NULL_DB_CONFIG_PARAMS = iota
    INVALID_DB_DRIVER
)

var ERROR_TYPES = []string{
    //NULL_DB_CONFIG_PARAMS
    "DB params on the configuration is empty/invalid",
    //INVALID_DB_DRIVER
    "INVALID DB driver in configuration"}