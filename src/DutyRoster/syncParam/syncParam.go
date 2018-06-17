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


package syncParam

//******************************************************************************
// Synchronization primitive for the application.The exit channels and waitgroups
// for the application should be defined in this file.
//******************************************************************************
import (
    "sync"
    "sync/atomic"
    "fmt"
)

type syncparams struct {
    // WaitGroup to keep track of threads that are currently running.
    appWaitGroups sync.WaitGroup
    // sync param for logproxy goroutine.
    do_log_exit chan bool
    // atomic counter to keep track of active goroutines.
    // Use Atomic ops to make sure synchronization.
    goroutineCnt int64
}

var appSync = new(syncparams)
var once sync.Once

// Initilize the global sync object with default set of values.
// Executed only once in application as it needed only for one sync object
func (syncObj *syncparams)InitSyncParams() {
    once.Do(func() {
        // Create the channel for logging thread handling.
        syncObj.do_log_exit = make(chan bool)
    })
}

func GetAppSyncObj() *syncparams {
    appSync.InitSyncParams()
    return appSync
}

// Exit the logger routine by sending the exit channel message to logger
// goroutine.
// Return true when message is sent successfully/ false otherwise.
func (syncObj *syncparams)ExitloggerRoutine() {
    syncObj.do_log_exit <- true
}

// Read the exit-log signal to see if kill signal is issues for logger routine
// Its a noblocking call, return true if exit is signaled, false otherwise.
// DO NOT INVOKE THIS FUNCTION FROM ANY FUNCTION OTHER THAN LOGGERPROXY LISTENER
func (syncObj *syncparams)IsLoggerExitSignaled() bool{
    select {
        case <- syncObj.do_log_exit:
            return true
        default:
            return false
    }
}

// Any goroutine invocation must precede with with this function.
// It allows the bookkeeping of currnetly running goroutines in the application.
func (syncObj *syncparams)AddRoutineInWaitGroup() {
    syncObj.appWaitGroups.Add(1)
    atomic.AddInt64(&syncObj.goroutineCnt, 1)
}

// Call when exiting the goroutine after its executing.
// It allows the book-keeping of active gorotuines in the application.
// NEVER INVOKE ExitRoutineInWaitGroup without AddRoutineInWaitGroup
func (syncObj *syncparams)ExitRoutineInWaitGroup() {
    syncObj.appWaitGroups.Done()
    // For simplicity, we assume done will not get called before add.
    atomic.AddInt64(&syncObj.goroutineCnt, -1)
}

// Function to wait for all the goroutines to complete execution.
// ONLY INVOKED FROM MAIN THREAD AS A LAST STATEMENT.
func (syncObj *syncparams)JoinAllRoutines() {
    syncObj.appWaitGroups.Wait()

}

// Function that signal exit message to all gorutines and wait for the
// routines to coalesce.
func (syncObj *syncparams)DestroyAllRoutines() {
    cnt := atomic.LoadInt64(&syncObj.goroutineCnt)
    if cnt <= 0 {
        return
    }
    //Exit logger routine
    syncObj.ExitloggerRoutine()
}

// Application panic.
func (syncObj *syncparams)PanicApp(msgfmt string, args ...interface{}) {
    fmt.Println("\n\n APPLICATION IS PANICKED \n\n")
    syncObj.DestroyAllRoutines()
    panicErr := fmt.Sprintf(msgfmt, args...)
    // System is panicked, hence mark all goroutine as done unconditionally, to
    // exit the application.
    // No lock for these operation as we assume there are no threads are now
    // operating at waitgroup. It is very unlikely a new thread is created/
    // deleted at this stage.
    cnt := atomic.LoadInt64(&syncObj.goroutineCnt)
    var i int64
    for i = 0; i < cnt ; i++ {
        syncObj.ExitRoutineInWaitGroup()
    }
    panic(panicErr)
}