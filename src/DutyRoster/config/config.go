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

package config

import (
    "os"
    "fmt"
    "encoding/json"
    "sync"
)

//Configuration json file format.
//Production/Debug configuration json files are created using this structure.
type Config struct {
    Logging struct {
        // loglevel can be trace, info, warning, error
        LogLevel string `json:"loglevel"`
        // Set filepath to empty to output logs only to stdout.
        FilePath string `json:"filepath"`
    }`json:"logging"`
    DB struct {
        //Name of DB driver, Only SQLLITE is supported now.
        Driver string `json:"driver"`
        //Path of DB to use in application., eg: /tmp/test.db
        Dbpath string `json:"dbpath"`
        //Ip address of Host where DB server is running.
        Ipaddr string `json:"ipaddr"`
        //Port at which DB server is listening for incoming connections.
        Port string `json:"port"`
        //Username if needed to connect to DB.
        Uname string `json:"uname"`
        //Password to connect to DB
        Pwd string `json:"pwd"`
        //Transport protocol to connect to db, can be tcp/udp
        Transport string `json:"transport`
    }`json:"db"`

}

var conf = new(Config)
var once sync.Once

// Singleton function to return the configuration object.
// Only one configuration object created for entire application.
// Any configuration change in the json file might need to restart the
// application.
func LoadConfigSingleton(configfile string) error {
    var err error
    err = nil
    once.Do(func() {
        var fp *os.File
        fp, err = os.Open(configfile)
        defer fp.Close()
        if err != nil {
            fmt.Printf("\nERROR: %s\n", err.Error())
            return
        }
        jsonParser := json.NewDecoder(fp)
        err = jsonParser.Decode(conf)
        if err != nil {
            fmt.Printf("\nERROR: Failed to parse JSON file, Invalid syntax\n")
            return
        }
    })
    return err
}

// Function to return config singleton instance.
func GetConfigInstance() *Config {
    return conf
}
