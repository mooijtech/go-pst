// go-pst is a library for reading Personal Storage Table (.pst) files (written in Go/Golang).
//
// Copyright (C) 2022  Marten Mooij
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
