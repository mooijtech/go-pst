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

package pst

import (
	"encoding/binary"

	"github.com/rotisserie/eris"
)

// BTreeOnHeapHeader represents the B-Tree-on-Heap header.
type BTreeOnHeapHeader struct {
	TableType uint8
	KeySize   uint8
	ValueSize uint8
	Levels    uint8
	HIDRoot   Identifier
}

// GetBTreeOnHeapHeader returns the B-Tree-on-Heap header.
func (file *File) GetBTreeOnHeapHeader(heapOnNode *HeapOnNode) (*BTreeOnHeapHeader, error) {
	// All tables should have a BTree-on-Heap header at HID 0x20 (HID User Root from the Heap-on-Node header).
	hidUserRoot, err := heapOnNode.GetHIDUserRoot()

	if err != nil {
		return nil, eris.Wrap(err, "failed to get HID user root")
	}

	btreeOnHeapReader, err := file.GetHeapOnNodeReaderFromHNID(hidUserRoot, *heapOnNode.Reader)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get Heap-on-Node reader from HNID")
	}

	btreeOnHeap := make([]byte, 8)

	if _, err := btreeOnHeapReader.ReadAt(btreeOnHeap, 0); err != nil {
		return nil, eris.Wrap(err, "failed to read b-tree-on-heap")
	}

	return &BTreeOnHeapHeader{
		TableType: btreeOnHeap[0],
		KeySize:   btreeOnHeap[1],
		ValueSize: btreeOnHeap[2],
		Levels:    btreeOnHeap[3],
		HIDRoot:   Identifier(binary.LittleEndian.Uint32(btreeOnHeap[4:])),
	}, nil
}
