// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
)

// LocalDescriptor represents an item in the local descriptors.
type LocalDescriptor struct {
	Data []byte
}

// GetIdentifier returns the identifier of the local descriptor.
// References "Local Descriptors".
func (localDescriptor *LocalDescriptor) GetIdentifier(formatType string) (int, error) {
	var identifierBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		identifierBufferSize = 8
		break
	case FormatTypeUnicode4k:
		identifierBufferSize = 8
		break
	case FormatTypeANSI:
		identifierBufferSize = 4
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint64(localDescriptor.Data[:identifierBufferSize])), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(localDescriptor.Data[:identifierBufferSize])), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(localDescriptor.Data[:identifierBufferSize])), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetDataIdentifier returns the data identifier of the local descriptor.
// References "Local Descriptors".
func (localDescriptor *LocalDescriptor) GetDataIdentifier(formatType string) (int, error) {
	var identifierOffset int
	var identifierBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		identifierOffset = 8
		identifierBufferSize = 8
		break
	case FormatTypeUnicode4k:
		identifierOffset = 8
		identifierBufferSize = 8
		break
	case FormatTypeANSI:
		identifierOffset = 4
		identifierBufferSize = 4
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint64(localDescriptor.Data[identifierOffset:identifierOffset + identifierBufferSize])), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(localDescriptor.Data[identifierOffset:identifierOffset + identifierBufferSize])), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(localDescriptor.Data[identifierOffset:identifierOffset + identifierBufferSize])), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetLocalDescriptorsIdentifier returns the local descriptors identifier of the local descriptor.
// References "Local Descriptors".
func (localDescriptor *LocalDescriptor) GetLocalDescriptorsIdentifier(formatType string) (int, error) {
	var identifierOffset int
	var identifierBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		identifierOffset = 16
		identifierBufferSize = 8
		break
	case FormatTypeUnicode4k:
		identifierOffset = 16
		identifierBufferSize = 8
		break
	case FormatTypeANSI:
		identifierOffset = 8
		identifierBufferSize = 4
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint64(localDescriptor.Data[identifierOffset:identifierOffset + identifierBufferSize])), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(localDescriptor.Data[identifierOffset:identifierOffset + identifierBufferSize])), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(localDescriptor.Data[identifierOffset:identifierOffset + identifierBufferSize])), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetLocalDescriptors returns the local descriptors of the b-tree node entry.
// References "Local Descriptors".
func (pstFile *File) GetLocalDescriptors(btreeNodeEntry BTreeNodeEntry, formatType string) ([]LocalDescriptor, error) {
	localDescriptorsIdentifier, err := btreeNodeEntry.GetLocalDescriptorsIdentifier(formatType)

	if err != nil {
		return nil, err
	}

	blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

	if err != nil {
		return nil, err
	}

	localDescriptorsNode, err := pstFile.FindBTreeNode(blockBTreeOffset, localDescriptorsIdentifier, formatType)

	if err != nil {
		return nil, err
	}

	localDescriptorsOffset, err := localDescriptorsNode.GetFileOffset(false, formatType)

	if err != nil {
		return nil, err
	}

	signature, err := pstFile.Read(1, localDescriptorsOffset)

	if err != nil {
		return nil, err
	}

	if binary.LittleEndian.Uint16([]byte{signature[0], 0}) != 2 {
		return nil, errors.New("invalid local descriptors signature")
	}

	//localDescriptorsSize, err := localDescriptorsNode.GetSize(formatType)
	//
	//if err != nil {
	//	return nil, err
	//}

	localDescriptorsLevel, err := pstFile.Read(1, localDescriptorsOffset + 1)

	if err != nil {
		return nil, err
	}

	if binary.LittleEndian.Uint16([]byte{localDescriptorsLevel[0], 0}) > 0 {
		// Haven't seen branch nodes yet.
		return nil, errors.New("local descriptors level is not 0, please open an issue on GitHub for this to be implemented")
	}

	var localDescriptorEntrySize int

	switch formatType {
	case FormatTypeUnicode:
		localDescriptorEntrySize = 24
		break
	case FormatTypeUnicode4k:
		localDescriptorEntrySize = 24
		break
	case FormatTypeANSI:
		localDescriptorEntrySize = 12
		break
	default:
		return nil, errors.New("unsupported format type")
	}

	localDescriptorsEntryCount, err := pstFile.Read(2, localDescriptorsOffset + 2)

	if err != nil {
		return nil, err
	}

	var localDescriptorsEntriesOffset int

	switch formatType {
	case FormatTypeUnicode:
		localDescriptorsEntriesOffset = localDescriptorsOffset + 8
		break
	case FormatTypeUnicode4k:
		localDescriptorsEntriesOffset = localDescriptorsOffset + 8
		break
	case FormatTypeANSI:
		localDescriptorsEntriesOffset = localDescriptorsOffset + 4
		break
	default:
		return []LocalDescriptor{}, errors.New("unsupported format type")
	}

	localDescriptorsEntries, err := pstFile.Read(int(binary.LittleEndian.Uint16(localDescriptorsEntryCount)) * localDescriptorEntrySize, localDescriptorsEntriesOffset)

	localDescriptors := make([]LocalDescriptor, binary.LittleEndian.Uint16(localDescriptorsEntryCount))

	for i := 0; i < int(binary.LittleEndian.Uint16(localDescriptorsEntryCount)); i++ {
		localDescriptorEntry := localDescriptorsEntries[i * localDescriptorEntrySize:(i + 1) * localDescriptorEntrySize]

		localDescriptors[i] = LocalDescriptor {
			Data: localDescriptorEntry,
		}
	}

	return localDescriptors, nil
}

// FindLocalDescriptor returns the local descriptor with the specified identifier.
func (pstFile *File) FindLocalDescriptor(localDescriptors []LocalDescriptor, identifier int, formatType string) (LocalDescriptor, error) {
	for _, localDescriptor := range localDescriptors {
		localDescriptorIdentifier, err := localDescriptor.GetIdentifier(formatType)

		if err != nil {
			return LocalDescriptor{}, err
		}

		if localDescriptorIdentifier == identifier {
			return localDescriptor, nil
		}
	}

	return LocalDescriptor{}, errors.New("failed to find local descriptor")
}