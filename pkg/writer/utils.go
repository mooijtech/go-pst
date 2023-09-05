// go-pst is a library for reading Personal Storage Table (.pst) files (written in Go/Golang).
//
// Copyright 2023 Marten Mooij
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package writer

import "encoding/binary"

// GetUint64 returns the Little Endian byte representation of the uint64.
func GetUint64(integer uint64) []byte {
	return binary.LittleEndian.AppendUint64([]byte{}, integer)
}

// GetUint32 returns the Little Endian byte representation of the uint32.
func GetUint32(integer uint32) []byte {
	return binary.LittleEndian.AppendUint32([]byte{}, integer)
}

// GetUint16 returns the Little Endian byte representation of the uint16.
func GetUint16(integer uint16) []byte {
	return binary.LittleEndian.AppendUint16([]byte{}, integer)
}
