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
    "sync"
    "DutyRoster/syncParam"
    "fmt"
    "runtime"
)

type loggerProxy struct {
    logObj *Logging
    logChannel chan loggerProxyChannel
    //size of logger channel, each channel will assign the 'channelSize'
    channelSize uint64
}

type loggerProxyChannel struct {
    fnPtr func(string, ...interface{})
    msg string
}

var gblLogProxy = new(loggerProxy)
var logProxyOnce sync.Once
var LOG_CHANNEL_SIZE uint64

func (logProxy *loggerProxy)executeLoggerRoutine() {
    syncObj := syncParam.GetAppSyncObj()
    // Exiting the logger, so mark done in waitgroup
    defer syncObj. ExitRoutineInWaitGroup()
    for syncObj.IsLoggerExitSignaled() == false {
        //Exit the goroutine only when a exit signal triggered from main thread
        select {
            //Read the channel message
            case logMsg := <- logProxy.logChannel:
                //Invoke relevant logging.
                logMsg.fnPtr(logMsg.msg)
            default:
                //Do nothing.
        }
    }
}

func (logProxy *loggerProxy)startLoggerListner() {
    syncObj := syncParam.GetAppSyncObj()
    // Adding logger goroutine into the waitgroup
    syncObj.AddRoutineInWaitGroup()
    go logProxy.executeLoggerRoutine()
}

func (logProxy *loggerProxy)initloggerProxy() {
    LOG_CHANNEL_SIZE = 5000
    logProxyOnce.Do(func() {
            if logProxy.logObj == nil {
                logProxy.logObj = getLoggerInstance()
            }
            logProxy.channelSize = LOG_CHANNEL_SIZE
            //create channel for sending message.
            logProxy.logChannel = make(chan loggerProxyChannel, logProxy.channelSize)
            logProxy.startLoggerListner()
        })
}

// Function to read the caller function name from the function stack.
func (logProxy *loggerProxy)getCallerName() string{
    pc := make([]uintptr, 1)
    //Skipping the functions that are part of loggerProxy to get right caller.	
    runtime.Callers(4, pc)
    f := runtime.FuncForPC(pc[0])
    return f.Name()
}

func (logProxy *loggerProxy)appendlog(msgfmt string, args ...interface{}) string{
    var str string
    str = fmt.Sprintf(":" + logProxy.getCallerName() + ":" + msgfmt, args...)
    return str
}

// Create a trace channel message and send it to logger hander goroutine.
func (logProxy *loggerProxy)Trace(msgfmt string, args ...interface{}) {
    var ch loggerProxyChannel
    ch.fnPtr = logProxy.logObj.Trace
    ch.msg = logProxy.appendlog(msgfmt, args...)
    logProxy.logChannel <- ch
}

// Create a Info channel message and send it to logger hander goroutine.
func (logProxy *loggerProxy)Info(msgfmt string, args ...interface{}) {
    var ch loggerProxyChannel
    ch.fnPtr = logProxy.logObj.Info
    ch.msg = logProxy.appendlog(msgfmt, args...)
    logProxy.logChannel <- ch
}

// Create a Warning channel message and send it to logger hander goroutine.
func (logProxy *loggerProxy)Warning(msgfmt string, args ...interface{}) {
    var ch loggerProxyChannel
    ch.fnPtr = logProxy.logObj.Warning
    ch.msg = logProxy.appendlog(msgfmt, args...)
    logProxy.logChannel <- ch
}

// Create a Error channel message and send it to logger hander goroutine.
func (logProxy *loggerProxy)Error(msgfmt string, args ...interface{}) {
    var ch loggerProxyChannel
    ch.fnPtr = logProxy.logObj.Error
    ch.msg = logProxy.appendlog(msgfmt, args...)
    logProxy.logChannel <- ch
}


func getLoggerProxyInstance() *loggerProxy{
    gblLogProxy.initloggerProxy()
    return gblLogProxy
}

