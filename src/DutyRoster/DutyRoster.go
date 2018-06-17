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

package main 

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "flag"
    "path/filepath"
    "DutyRoster/config"
    "DutyRoster/logging"
    "DutyRoster/syncParam"
)


//function to read configuration json file and convert it to configuration
//object.
func readConfig() {
    var err error
    var cfgAbsPath string
    var cfgfile *string
    var cfgfileInput = flag.String("c", "", "Appplication json configuration file")
    var cfgfileLongInput = flag.String("cfgfile", "", "Appplication json configuration file")
    flag.Parse()
    cfgfile = cfgfileInput
    if len(*cfgfileInput) == 0 {
        // May be parameter is provided with cfgfile longpath input
        cfgfile = cfgfileLongInput
    }
    cfgAbsPath, err = filepath.Abs(*cfgfile)
    if (err != nil) {
        fmt.Print("Failed to open Config file, Cannot start application")
        return
    }
    config.LoadConfigSingleton(cfgAbsPath)
}

//Function to setup the logging for application.
func setuplogging() {
    var appLogObj = logging.GetAppLoggerObj()
    appLogObj.Trace("test")
    appLogObj.Info("%s info test %d", "num", 1)
}

func printHelp() {
    helpstr := "\n\t DutyRoster Server Application" +
    "\n\t An application to schedule work shifts for employeess in an org." +
    "\n\t   USAGE: ./DutyRoster {ARGS}" +
    "\n\t      ARGS:" +
    "\n\t      -c <file>         :- Appplication json configuration file" +
    "\n\t      -cfgfile <file>  :- Appplication json configuration file\n\n"
    fmt.Print(helpstr)
}

func main() {
    if len(os.Args) <= 1 {
        //No arguments provided, print helpstring.
        printHelp()
        fmt.Println("ERROR: Failed to start application, " +
                    "manadatory args missing" )
        return
    }
    //Initilizing the app synchronization constructs.
    syncObj := syncParam.GetAppSyncObj()
    // Wait for all routines to coalesce
    defer syncObj.JoinAllRoutines()
    readConfig()
    setuplogging()

    // Exit the main thread on Ctrl C 
    fmt.Print("\n\n\n *** Press Ctrl+C to Exit *** \n\n\n")
    exitsignal := make(chan os.Signal, 1)
    signal.Notify(exitsignal, syscall.SIGINT, syscall.SIGTERM)
    //Add exit routine into waitgroup
    syncObj.AddRoutineInWaitGroup()
    go func() {
        // Blocking the routine for the exit signal.
        <- exitsignal
        //Send exit signal to all the goroutines.
        syncObj.DestroyAllRoutines()
        //Mark exit routine is done 
        syncObj.ExitRoutineInWaitGroup()
    }()
}

