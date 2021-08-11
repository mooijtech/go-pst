// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
)

type BTreeOnHeapHeader struct {
	StartOffset int
	EndOffset int


}

// GetBTreeOnHeapHeader returns the btree on heap header.
// TODO - Document this better.
// Based on https://github.com/rjohnsondev/java-libpst/blob/develop/src/main/java/com/pff/PSTTable.java#L206
func (pstFile *File) GetBTreeOnHeapHeader(btreeNodeEntryHeapOnNode BTreeNodeEntry, formatType string) error {
	// All tables should have a BTree-on-Heap header at NID 0x20
	nid := 0x20

	whichBlock := nid >> 16

	nodeEntryBlocks, err := pstFile.GetHeapOnNodeBlocks(btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return err
	}

	if whichBlock > len(nodeEntryBlocks) {
		return errors.New("block doesn't exist")
	}

	index := (nid & 0xFFFF) >> 5

	heapOnNodePageMap := btreeNodeEntryHeapOnNode.GetHeapOnNodePageMap()

	//allocationCount := binary.LittleEndian.Uint16(btreeNodeEntryHeapOnNode.Data[heapOnNodePageMap:heapOnNodePageMap + 2])
	//
	//log.Infof("Allocation count: %d", allocationCount)

	heapOnNodePageMap += (2 * index) + 2

	start := binary.LittleEndian.Uint16(btreeNodeEntryHeapOnNode.Data[heapOnNodePageMap:heapOnNodePageMap + 2])
	end := binary.LittleEndian.Uint16(btreeNodeEntryHeapOnNode.Data[heapOnNodePageMap + 2:heapOnNodePageMap + 2 + 2])

	log.Infof("Start: %d", start)
	log.Infof("End: %d", end)

	// TODO - Use the start offset to parse the table.

	return nil
}