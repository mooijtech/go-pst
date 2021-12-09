// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
)

// LocalDescriptor represents an item in the local descriptors.
// A local descriptor is basically a reference to a node which contains the data.
type LocalDescriptor struct {
	data []byte
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

	return int(binary.LittleEndian.Uint32(localDescriptor.data[:identifierBufferSize])), nil
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

	return int(binary.LittleEndian.Uint32(localDescriptor.data[identifierOffset : identifierOffset+identifierBufferSize])), nil
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

	return int(binary.LittleEndian.Uint32(localDescriptor.data[identifierOffset : identifierOffset+identifierBufferSize])), nil
}

func (pstFile *File) GetLocalDescriptors(btreeNodeEntry BTreeNodeEntry, formatType string) ([]LocalDescriptor, error) {
	localDescriptorsIdentifier, err := btreeNodeEntry.GetLocalDescriptorsIdentifier(formatType)

	if err != nil {
		return nil, err
	}

	return pstFile.GetLocalDescriptorsFromIdentifier(localDescriptorsIdentifier, formatType)
}

// GetLocalDescriptorsFromIdentifier returns the local descriptors of the local descriptors identifier.
// References "Local Descriptors".
func (pstFile *File) GetLocalDescriptorsFromIdentifier(localDescriptorsIdentifier int, formatType string) ([]LocalDescriptor, error) {
	if localDescriptorsIdentifier == 0 {
		// There are no local descriptors
		return []LocalDescriptor{}, nil
	}

	localDescriptorsNode, err := pstFile.GetBlockBTreeNode(localDescriptorsIdentifier, formatType)

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

	localDescriptorsLevel, err := pstFile.Read(1, localDescriptorsOffset+1)

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

	localDescriptorsEntryCount, err := pstFile.Read(2, localDescriptorsOffset+2)

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

	localDescriptorsEntries, err := pstFile.Read(int(binary.LittleEndian.Uint16(localDescriptorsEntryCount))*localDescriptorEntrySize, localDescriptorsEntriesOffset)

	localDescriptors := make([]LocalDescriptor, binary.LittleEndian.Uint16(localDescriptorsEntryCount))

	for i := 0; i < int(binary.LittleEndian.Uint16(localDescriptorsEntryCount)); i++ {
		localDescriptorEntry := localDescriptorsEntries[i*localDescriptorEntrySize : (i+1)*localDescriptorEntrySize]

		localDescriptors[i] = LocalDescriptor{
			data: localDescriptorEntry,
		}
	}

	return localDescriptors, nil
}

// FindLocalDescriptor returns the local descriptor with the specified identifier.
func FindLocalDescriptor(localDescriptors []LocalDescriptor, identifier int, formatType string) (LocalDescriptor, error) {
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

// GetData returns all the local descriptor data from the data node.
func (localDescriptor *LocalDescriptor) GetData(pstFile *File, formatType string, encryptionType string) ([]byte, error) {
	dataIdentifier, err := localDescriptor.GetDataIdentifier(formatType)

	if err != nil {
		return nil, err
	}

	blockBTreeNode, err := pstFile.GetBlockBTreeNode(dataIdentifier, formatType)

	if err != nil {
		return nil, err
	}

	inputStream, err := pstFile.NewHeapOnNodeInputStream(blockBTreeNode, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	var data []byte

	if len(inputStream.Blocks) > 0 {
		currentOffset := 0

		for _, block := range inputStream.Blocks {
			blockSize, err := block.GetSize(formatType)

			if err != nil {
				return nil, err
			}

			blockData, err := inputStream.Read(blockSize, currentOffset)

			if err != nil {
				return nil, err
			}

			data = append(data, blockData...)
			currentOffset += blockSize
		}
	} else {
		data, err = inputStream.Read(inputStream.Size, 0)

		if err != nil {
			return nil, err
		}
	}

	return data, nil
}