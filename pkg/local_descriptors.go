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

// LocalDescriptor represents an item in the local descriptors.
// A local descriptor is basically a reference to a node which contains the data.
type LocalDescriptor struct {
	Identifier                 Identifier
	DataIdentifier             Identifier
	LocalDescriptorsIdentifier Identifier
}

// NewLocalDescriptor creates a new local descriptor.
func NewLocalDescriptor(data []byte, formatType FormatType) LocalDescriptor {
	switch formatType {
	case FormatTypeANSI:
		ld := LocalDescriptor{
			Identifier:     Identifier(binary.LittleEndian.Uint32(data[:4])),
			DataIdentifier: Identifier(binary.LittleEndian.Uint32(data[4 : 4+4])),
		}

		if len(data) > 8 {
			ld.LocalDescriptorsIdentifier = Identifier(binary.LittleEndian.Uint32(data[8 : 8+4]))
		}

		return ld
	default:
		// TODO - Reference [MS-PDF] that this is actually 32-bit.
		ld := LocalDescriptor{
			Identifier:     Identifier(binary.LittleEndian.Uint32(data[:8])),
			DataIdentifier: Identifier(binary.LittleEndian.Uint32(data[8 : 8+8])),
		}

		if len(data) > 16 {
			ld.LocalDescriptorsIdentifier = Identifier(binary.LittleEndian.Uint32(data[16 : 16+8]))
		}

		return ld
	}
}

// GetLocalDescriptors returns the local descriptors of the b-tree node.
func (file *File) GetLocalDescriptors(btreeNodeEntry BTreeNode) ([]LocalDescriptor, error) {
	return file.GetLocalDescriptorsFromIdentifier(btreeNodeEntry.LocalDescriptorsIdentifier)
}

const (
	SlBlockEntry = 0
	SiBlockEntry = 1
)

// GetLocalDescriptorsFromIdentifier returns the local descriptors of the local descriptors identifier.
// References "Local Descriptors".
func (file *File) GetLocalDescriptorsFromIdentifier(localDescriptorsIdentifier Identifier) ([]LocalDescriptor, error) {
	if localDescriptorsIdentifier == 0 {
		// There are no local descriptors.
		return nil, nil
	}

	localDescriptorsNode, err := file.GetBlockBTreeNode(localDescriptorsIdentifier)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get local descriptors node")
	}

	// TODO - Merge signature, level, entry count etc into one ReadAt
	signature := make([]byte, 1)

	if _, err = file.Reader.ReadAt(signature, localDescriptorsNode.FileOffset); err != nil {
		return nil, eris.Wrap(err, "failed to read local descriptors signature")
	} else if signature[0] != 2 {
		return nil, ErrLocalDescriptorsSignatureInvalid
	}

	localDescriptorsLevel := make([]byte, 1)

	if _, err := file.Reader.ReadAt(localDescriptorsLevel, localDescriptorsNode.FileOffset+1); err != nil {
		return nil, eris.Wrap(err, "failed to read local descriptors level")
	}

	blockType := localDescriptorsLevel[0]

	localDescriptorsEntryCountBytes := make([]byte, 2)

	if _, err := file.Reader.ReadAt(localDescriptorsEntryCountBytes, localDescriptorsNode.FileOffset+2); err != nil {
		return nil, eris.Wrap(err, "failed to get local descriptors entry count")
	}

	localDescriptorsEntryCount := binary.LittleEndian.Uint16(localDescriptorsEntryCountBytes)

	var localDescriptorsEntriesOffset int64

	switch file.FormatType {
	case FormatTypeANSI:
		localDescriptorsEntriesOffset = localDescriptorsNode.FileOffset + 4
	default:
		localDescriptorsEntriesOffset = localDescriptorsNode.FileOffset + 8
	}

	localDescriptors := make([]LocalDescriptor, 0, localDescriptorsEntryCount)

	for i := 1; i <= int(localDescriptorsEntryCount); i++ {
		var entryDataSize int64
		if blockType == SlBlockEntry {
			if file.FormatType == FormatTypeANSI {
				entryDataSize = 12
			} else {
				entryDataSize = 24
			}
		} else {
			if file.FormatType == FormatTypeANSI {
				entryDataSize = 8
			} else {
				entryDataSize = 16
			}
		}

		localDescriptorEntry := make([]byte, entryDataSize)

		if _, err := file.Reader.ReadAt(localDescriptorEntry, localDescriptorsEntriesOffset); err != nil {
			return nil, eris.Wrap(err, "failed to get local descriptors entry count")
		}

		localDescriptor := NewLocalDescriptor(localDescriptorEntry, file.FormatType)
		if blockType == SlBlockEntry {
			localDescriptors = append(localDescriptors, localDescriptor)
		} else {
			descriptors, err := file.GetLocalDescriptorsFromIdentifier(localDescriptor.DataIdentifier)
			if err != nil {
				return nil, eris.Wrap(err, "failed to get descriptors from identifier")
			}

			localDescriptors = append(localDescriptors, descriptors...)
		}

		// Increase the offset so its at the start of the next entry
		localDescriptorsEntriesOffset += entryDataSize
	}

	return localDescriptors, nil
}

// FindLocalDescriptor returns the local descriptor with the specified identifier or an error if not found.
func FindLocalDescriptor(identifier Identifier, localDescriptors []LocalDescriptor) (LocalDescriptor, error) {
	for _, localDescriptor := range localDescriptors {
		if localDescriptor.Identifier == identifier {
			return localDescriptor, nil
		}
	}

	return LocalDescriptor{}, ErrLocalDescriptorNotFound
}
