// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"errors"
)

// AllocationTableOffsets represent the start and end offset of a Heap-on-Node item.
type AllocationTableOffsets struct {
	Data        []byte
	StartOffset int
	EndOffset   int
}

// GetAllocationTableNodeInputStream returns the offsets from the allocation table of the given HID.
func (pstFile *File) GetAllocationTableNodeInputStream(hid int, heapOnNode HeapOnNode, localDescriptors []LocalDescriptor, formatType string, encryptionType string) (NodeInputStream, error) {
	if len(localDescriptors) > 0 {
		localDescriptor, err := pstFile.FindLocalDescriptor(localDescriptors, hid, formatType)

		if err == nil {
			// Found the local descriptor for this identifier.
			localDescriptorHeapOnNode, err := pstFile.NewHeapOnNodeFromLocalDescriptor(localDescriptor, formatType, encryptionType)

			if err != nil {
				return NodeInputStream{}, err
			}

			return localDescriptorHeapOnNode.NodeInputStream, nil
		}
	}

	if (hid & 0x1F) != 0 {
		return NodeInputStream{}, nil
	}

	hidBlockIndex := hid >> 16

	if hidBlockIndex > 0 {
		return NodeInputStream{}, errors.New("block doesn't exist")
	}

	hidIndex := (hid & 0xFFFF) >> 5

	heapOnNodePageMap, err := heapOnNode.GetPageMap()

	if err != nil {
		return NodeInputStream{}, err
	}

	// The allocation table starts at byte offset 4 from the page map.
	// Every 2 bytes in the allocation table is the offset of the item.
	heapOnNodePageMap += (2 * hidIndex) + 2

	startOffset, err := heapOnNode.NodeInputStream.SeekAndReadUint16(2, heapOnNodePageMap)

	if err != nil {
		return NodeInputStream{}, err
	}

	endOffset, err := heapOnNode.NodeInputStream.SeekAndReadUint16(2, heapOnNodePageMap+2)

	return NodeInputStream{
		File:           pstFile,
		EncryptionType: encryptionType,
		FileOffset:     heapOnNode.NodeInputStream.FileOffset + startOffset,
		Size:           endOffset - startOffset,
	}, nil
}

// BTreeOnHeapHeader represents the b-tree on heap header.
type BTreeOnHeapHeader struct {
	TableType int
	KeySize   int
	ValueSize int
	Levels    int
	HIDRoot   int
}

// GetBTreeOnHeapHeader returns the btree on heap header.
func (pstFile *File) GetBTreeOnHeapHeader(heapOnNode HeapOnNode, localDescriptors []LocalDescriptor, formatType string, encryptionType string) (BTreeOnHeapHeader, error) {
	// All tables should have a BTree-on-Heap header at HID 0x20 (HID User Root from the Heap-on-Node header).
	hidUserRoot, err := heapOnNode.GetHIDUserRoot()

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	nodeInputStream, err := pstFile.GetAllocationTableNodeInputStream(hidUserRoot, heapOnNode, localDescriptors, formatType, encryptionType)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapTableType, err := nodeInputStream.SeekAndReadUint16(1, 0)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapKeySize, err := nodeInputStream.SeekAndReadUint16(1, 1)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapValueSize, err := nodeInputStream.SeekAndReadUint16(1, 2)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapLevels, err := nodeInputStream.SeekAndReadUint16(1, 3)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapHIDRoot, err := nodeInputStream.SeekAndReadUint32(4, 4)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	return BTreeOnHeapHeader{
		TableType: btreeOnHeapTableType,
		KeySize:   btreeOnHeapKeySize,
		ValueSize: btreeOnHeapValueSize,
		Levels:    btreeOnHeapLevels,
		HIDRoot:   btreeOnHeapHIDRoot,
	}, nil
}
