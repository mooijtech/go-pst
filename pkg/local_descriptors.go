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
	Identifier int
	DataIdentifier int
	LocalDescriptorsIdentifier int
}

// GetLocalDescriptors returns the local descriptors of the b-tree node entry.
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

	localDescriptorsEntrySize, err := localDescriptorsNode.GetSize(formatType)

	if err != nil {
		return nil, err
	}

	localDescriptorsLevel, err := pstFile.Read(1, localDescriptorsOffset + 1)

	if err != nil {
		return nil, err
	}

	if binary.LittleEndian.Uint16([]byte{localDescriptorsLevel[0], 0}) > 0 {
		// Haven't seen branch nodes yet.
		return nil, errors.New("local descriptors level is not 0, please open an issue on GitHub for this to be implemented")
	}

	localDescriptorsEntryCount, err := pstFile.Read(2, localDescriptorsOffset + 2)

	if err != nil {
		return nil, err
	}

	localDescriptorsEntries, err := pstFile.Read(int(binary.LittleEndian.Uint16(localDescriptorsEntryCount)) * localDescriptorsEntrySize, localDescriptorsOffset + 8)

	localDescriptors := make([]LocalDescriptor, binary.LittleEndian.Uint16(localDescriptorsEntryCount))

	for i := 0; i < int(binary.LittleEndian.Uint16(localDescriptorsEntryCount)); i++ {
		localDescriptorEntry := localDescriptorsEntries[i * localDescriptorsEntrySize:(i + 1) * localDescriptorsEntrySize]

		localDescriptors = append(localDescriptors, LocalDescriptor {
			Identifier: int(binary.LittleEndian.Uint32(localDescriptorEntry[:8])),
			DataIdentifier: int(binary.LittleEndian.Uint32(localDescriptorEntry[8:16])),
			LocalDescriptorsIdentifier: int(binary.LittleEndian.Uint32(localDescriptorEntry[16:24])),
		})
	}

	return localDescriptors, nil
}

// FindLocalDescriptor returns the local descriptor with the specified identifier.
func (pstFile *File) FindLocalDescriptor(localDescriptors []LocalDescriptor, identifier int) (LocalDescriptor, error) {
	for _, localDescriptor := range localDescriptors {
		if localDescriptor.Identifier == identifier {
			return localDescriptor, nil
		}
	}

	return LocalDescriptor{}, errors.New("failed to find local descriptor")
}