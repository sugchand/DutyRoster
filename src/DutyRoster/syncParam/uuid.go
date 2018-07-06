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

import (
    "crypto/rand"
    "fmt"
    "bytes"
)

type UUID [16]byte

// newUUID generates a random UUID according to RFC 4122
func NewUUIDString() (string, error) {
    uuid, err := NewUUID()
    if err != nil {
        return "", nil
    }
    return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
        uuid[0],uuid[1],uuid[2],uuid[3],
        uuid[4],uuid[5],uuid[6],uuid[7],
        uuid[8],uuid[9],uuid[10],uuid[11],
        uuid[12],uuid[13],uuid[14],uuid[15]), nil
}

func NewUUID() (UUID, error) {
    uuid := new(UUID)
    n, err := rand.Read(uuid[:])
    if n != len(uuid) || err != nil {
        return UUID{}, err
    }
    // variant bits; see section 4.1.1
    uuid[8] = uuid[8]&^0xc0 | 0x80
    // version 4 (pseudo-random); see section 4.1.3
    uuid[6] = uuid[6]&^0xf0 | 0x40
    return *uuid,nil
}

func UUIDtoString(uuid UUID) (string) {
    return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
        uuid[0],uuid[1],uuid[2],uuid[3],
        uuid[4],uuid[5],uuid[6],uuid[7],
        uuid[8],uuid[9],uuid[10],uuid[11],
        uuid[12],uuid[13],uuid[14],uuid[15])
}

func StringtoUUID(uuidStr string) UUID {
    var uuid = new(UUID)
    fmt.Sscanf(uuidStr, "%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
        &uuid[0],&uuid[1],&uuid[2],&uuid[3],
        &uuid[4],&uuid[5],&uuid[6],&uuid[7],
        &uuid[8],&uuid[9],&uuid[10],&uuid[11],
        &uuid[12],&uuid[13], &uuid[14],&uuid[15])
    return *uuid
}

//Returns True when uuid is empty and false otherwise.
func IsUUIDEmpty(uuid UUID) bool{
    zerouuid := new(UUID)
    if bytes.Equal(zerouuid[:], uuid[:]) {
        return true
    }
    return false
}