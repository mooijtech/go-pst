// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
)

// GetHeapOnNodeAllocationTableOffset returns the offset from the allocation table of the given HID.
func (pstFile *File) GetHeapOnNodeAllocationTableOffset(hid int, btreeNodeEntryHeapOnNode BTreeNodeEntry, formatType string) (int, error) {
	hidBlockIndex := hid >> 16

	nodeEntryBlocks, err := pstFile.GetHeapOnNodeBlocks(btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return -1, err
	}

	if hidBlockIndex > len(nodeEntryBlocks) {
		return -1, errors.New("block doesn't exist")
	}

	hidIndex := (hid & 0xFFFF) >> 5

	heapOnNodePageMap := btreeNodeEntryHeapOnNode.GetHeapOnNodePageMap()
	// The allocation table starts at byte offset 4 from the page map.
	// Every 2 bytes in the allocation table is the offset of the item.
	heapOnNodePageMap += (2 * hidIndex) + 2

	return int(binary.LittleEndian.Uint16(btreeNodeEntryHeapOnNode.Data[heapOnNodePageMap:heapOnNodePageMap + 2])), nil
}

// BTreeOnHeapHeader represents the b-tree on heap header.
type BTreeOnHeapHeader struct {
	TableType int
	RecordKeySize int
	RecordValueSize int
	RecordLevels int
	HIDRoot int
}

// NewBTreeOnHeapHeader is a constructor for creating b-tree on heap headers.
func NewBTreeOnHeapHeader(tableTable int, recordKeySize int, recordValueSize int, recordLevels int, hidRoot int) BTreeOnHeapHeader {
	return BTreeOnHeapHeader {
		TableType: tableTable,
		RecordKeySize: recordKeySize,
		RecordValueSize: recordValueSize,
		RecordLevels: recordLevels,
		HIDRoot: hidRoot,
	}
}

// GetBTreeOnHeapHeader returns the btree on heap header.
func (pstFile *File) GetBTreeOnHeapHeader(btreeNodeEntryHeapOnNode BTreeNodeEntry, formatType string) (BTreeOnHeapHeader, error) {
	// All tables should have a BTree-on-Heap header at HID 0x20.
	hid := 0x20

	btreeOnHeapHeaderOffset, err := pstFile.GetHeapOnNodeAllocationTableOffset(hid, btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return BTreeOnHeapHeader{}, err
	}

	// Parse the b-tree on heap header.
	btreeOnHeapTableType := int(binary.LittleEndian.Uint16([]byte{btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset], 0}))
	btreeOnHeapRecordKeySize := int(binary.LittleEndian.Uint16([]byte{btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset + 1], 0}))
	btreeOnHeapRecordValueSize := int(binary.LittleEndian.Uint16([]byte{btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset + 2], 0}))
	btreeOnHeapRecordLevels := int(binary.LittleEndian.Uint16([]byte{btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset + 3], 0}))
	btreeOnHeapHIDRoot := int(binary.LittleEndian.Uint16(btreeNodeEntryHeapOnNode.Data[btreeOnHeapHeaderOffset + 4:btreeOnHeapHeaderOffset + 8]))

	return NewBTreeOnHeapHeader(btreeOnHeapTableType, btreeOnHeapRecordKeySize, btreeOnHeapRecordValueSize, btreeOnHeapRecordLevels, btreeOnHeapHIDRoot), nil
}