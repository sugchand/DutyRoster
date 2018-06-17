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
    "fmt"
    "os"
    "log"
    "DutyRoster/config"
    "sync"
    "io"
)

const (
    Trace = iota
    Info
    Warning
    Error
)

type Logging struct {
    // loglevel can be Trace/Info/Warning/Error
    currloglevel int
    //format of the logs to be printed.
    logformatFlags int
    tracerLogger *log.Logger
    infoLogger *log.Logger
    warningLogger *log.Logger
    errorLogger *log.Logger
    fp io.Writer
}

var logconf = new(Logging)
var once sync.Once

// Singleton function to initilized the Logging instance. 
// Application can have only single Logging instance to keep limited memory
// usage.
func (logger *Logging)logInitSingleton() {
    once.Do(func() {
        var err error
        var stdoutHandler io.Writer
        stdoutHandler = os.Stdout
        conf := config.GetConfigInstance()
        if conf == nil {
            fmt.Println("\nERROR: Cannot read configfile object\n")
            return
        }
        logger.currloglevel = logger.getloglevelInt(conf.Logging.LogLevel)
        logger.logformatFlags = log.Ldate | log.Ltime
        if len(conf.Logging.FilePath) == 0 {
            logger.fp = stdoutHandler
        } else {
            logger.fp, err = os.OpenFile(conf.Logging.FilePath,
                os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
            if err != nil {
                logger.fp = stdoutHandler 
            }
        }
        logger.initloggers()
    })
}

func (logger *Logging)initloggers() {
    logger.tracerLogger = log.New(logger.fp, "TRACE: ", logger.logformatFlags)
    logger.infoLogger = log.New(logger.fp, "INFO: ", logger.logformatFlags)
    logger.warningLogger = log.New(logger.fp, "WARNING: ", logger.logformatFlags)
    logger.errorLogger = log.New(logger.fp, "ERROR: ", logger.logformatFlags) 
}

// Translate the loglevel string provided in the config file to
// Integer.
func (logger *Logging)getloglevelInt(loglevelstr string) int {
    switch(loglevelstr) {
        case "trace":
            return Trace
        case "info":
            return Info
        case "warning":
            return Warning
        case "error":
            return Error
    }
    fmt.Println("Invalid loglevel, starting with defaul Logging 'INFO'")
    return Info
}

func getLoggerInstance() *Logging{
    logconf.logInitSingleton()
    return logconf
}

func (logger *Logging)Trace(msgfmt string, args ...interface{}) {
    if logger.currloglevel > Trace {
        return
    }
    logger.tracerLogger.Printf(msgfmt, args...)
}

func (logger *Logging)Info(msgfmt string, args ...interface{}) {
    if logger.currloglevel > Info {
        return
    }
    logger.infoLogger.Printf(msgfmt, args...)
}

func (logger *Logging)Warning(msgfmt string, args ...interface{}) {
    if logger.currloglevel > Warning {
        return
    }
    logger.warningLogger.Printf(msgfmt, args...)
}

func (logger *Logging)Error(msgfmt string, args ...interface{}) {
    if logger.currloglevel > Error {
        return
    }
    logger.errorLogger.Printf(msgfmt, args...)
}
