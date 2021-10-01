// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
)

// AllocationTableOffsets represent the start and end offset of a Heap-on-Node item.
type AllocationTableOffsets struct {
	Data []byte
	StartOffset int
	EndOffset int
}

// GetHeapOnNodeAllocationTableOffsets returns the offsets from the allocation table of the given HID.
func (pstFile *File) GetHeapOnNodeAllocationTableOffsets(hid int, btreeNodeEntryHeapOnNode BTreeNodeEntry, localDescriptors []LocalDescriptor, formatType string) (AllocationTableOffsets, error) {
	if len(localDescriptors) > 0 {
		localDescriptor, err := pstFile.FindLocalDescriptor(localDescriptors, hid, formatType)

		if err == nil {
			// Found the local descriptor for this identifier.
			localDescriptorDataIdentifier, err := localDescriptor.GetDataIdentifier(formatType)

			if err != nil {
				return AllocationTableOffsets{}, err
			}

			blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

			if err != nil {
				return AllocationTableOffsets{}, err
			}

			localDescriptorNode, err := pstFile.FindBTreeNode(blockBTreeOffset, localDescriptorDataIdentifier, formatType)

			if err != nil {
				return AllocationTableOffsets{}, err
			}

			localDescriptorOffset, err := localDescriptorNode.GetFileOffset(false, formatType)

			if err != nil {
				return AllocationTableOffsets{}, err
			}

			localDescriptorNodeSize, err := localDescriptorNode.GetSize(formatType)

			if err != nil {
				return AllocationTableOffsets{}, err
			}

			// TODO - Implement a "node input stream" and return a new input stream with this data.
			localDescriptorNodeData, err := pstFile.Read(localDescriptorNodeSize, localDescriptorOffset)

			return AllocationTableOffsets {
				Data: DecodeCompressibleEncryption(localDescriptorNodeData),
				StartOffset: 0,
				EndOffset: localDescriptorNodeSize,
			}, nil
		}
	}

	if (hid & 0x1F) != 0 {
		return AllocationTableOffsets{}, nil
	}

	hidBlockIndex := hid >> 16

	nodeEntryBlocks, err := btreeNodeEntryHeapOnNode.GetBlocks(formatType)

	if err != nil {
		return AllocationTableOffsets{}, err
	}

	if hidBlockIndex > len(nodeEntryBlocks) {
		return AllocationTableOffsets{}, errors.New("block doesn't exist")
	}

	hidIndex := (hid & 0xFFFF) >> 5

	heapOnNodePageMap := btreeNodeEntryHeapOnNode.GetPageMap()
	// The allocation table starts at byte offset 4 from the page map.
	// Every 2 bytes in the allocation table is the offset of the item.
	heapOnNodePageMap += (2 * hidIndex) + 2

	startOffset := int(binary.LittleEndian.Uint16(btreeNodeEntryHeapOnNode.Data[heapOnNodePageMap:heapOnNodePageMap + 2]))
	endOffset := int(binary.LittleEndian.Uint16(btreeNodeEntryHeapOnNode.Data[heapOnNodePageMap + 2:heapOnNodePageMap + 4]))

	return AllocationTableOffsets {
		StartOffset: startOffset,
		EndOffset: endOffset,
	}, nil
}

// BTreeOnHeapHeader represents the b-tree on heap header.
type BTreeOnHeapHeader struct {
	TableType int
	KeySize int
	ValueSize int
	Levels int
	HIDRoot int
}

// GetBTreeOnHeapHeader returns the btree on heap header.
func (pstFile *File) GetBTreeOnHeapHeader(btreeNodeEntryHeapOnNode BTreeNodeEntry, localDescriptors []LocalDescriptor, formatType string) (BTreeOnHeapHeader, error) {
	// All tables should have a BTree-on-Heap header at HID 0x20 (HID User Root from the Heap-on-Node header).
	allocationTableOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(btreeNodeEntryHeapOnNode.GetHIDUserRoot(), btreeNodeEntryHeapOnNode, localDescriptors, formatType)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	btreeOnHeapHeaderOffset := allocationTableOffsets.StartOffset

	// Parse the b-tree on heap header.
	btreeOnHeapTableType := int(binary.LittleEndian.Uint16([]byte{btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset], 0}))
	btreeOnHeapKeySize := int(binary.LittleEndian.Uint16([]byte{btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset + 1], 0}))
	btreeOnHeapValueSize := int(binary.LittleEndian.Uint16([]byte{btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset + 2], 0}))
	btreeOnHeapLevels := int(binary.LittleEndian.Uint16([]byte{btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset + 3], 0}))
	btreeOnHeapHIDRoot := int(binary.LittleEndian.Uint16(btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset + 4:btreeOnHeapHeaderOffset + 8]))

	return BTreeOnHeapHeader {
		TableType: btreeOnHeapTableType,
		KeySize: btreeOnHeapKeySize,
		ValueSize: btreeOnHeapValueSize,
		Levels: btreeOnHeapLevels,
		HIDRoot: btreeOnHeapHIDRoot,
	}, nil
}