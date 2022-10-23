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

	"github.com/pkg/errors"
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
		return nil, errors.WithStack(err)
	}

	btreeOnHeapReader, err := file.GetHeapOnNodeReaderFromHNID(hidUserRoot, *heapOnNode.Reader)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	btreeOnHeapTableType := make([]byte, 1)

	if _, err := btreeOnHeapReader.ReadAt(btreeOnHeapTableType, 0); err != nil {
		return nil, errors.WithStack(err)
	}

	btreeOnHeapKeySize := make([]byte, 1)

	if _, err := btreeOnHeapReader.ReadAt(btreeOnHeapKeySize, 1); err != nil {
		return nil, errors.WithStack(err)
	}

	btreeOnHeapValueSize := make([]byte, 1)

	if _, err := btreeOnHeapReader.ReadAt(btreeOnHeapValueSize, 2); err != nil {
		return nil, errors.WithStack(err)
	}

	btreeOnHeapLevels := make([]byte, 1)

	if _, err := btreeOnHeapReader.ReadAt(btreeOnHeapLevels, 3); err != nil {
		return nil, errors.WithStack(err)
	}

	btreeOnHeapHIDRoot := make([]byte, 4)

	if _, err := btreeOnHeapReader.ReadAt(btreeOnHeapHIDRoot, 4); err != nil {
		return nil, errors.WithStack(err)
	}

	return &BTreeOnHeapHeader{
		TableType: btreeOnHeapTableType[0],
		KeySize:   btreeOnHeapKeySize[0],
		ValueSize: btreeOnHeapValueSize[0],
		Levels:    btreeOnHeapLevels[0],
		HIDRoot:   Identifier(binary.LittleEndian.Uint32(btreeOnHeapHIDRoot)),
	}, nil
}
