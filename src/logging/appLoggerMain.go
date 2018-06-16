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

package logging

import (

)

// Main function that return the App logging object. It is possible to have
// multiple logger implementation of LoggingInterface present in the system.
// For eg: logging, loggingProxy and so on. Update this function to let
// application to use right logging implementation on different use case.
// By default application uses loggingproxy implementation for the logging.
func GetAppLoggerObj()LoggingInterface {
    // Can use different implementation if needed.
    return getLoggerProxyInstance()
}
