// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
)

// ColumnDescriptor represents a column in the Table Context.
// References "Table Context", "Table Context Column Descriptor".
type ColumnDescriptor struct {
	DataOffset int
	DataSize int
}

// NewColumnDescriptor is a constructor for creating column descriptors.
func NewColumnDescriptor(tableContext []byte, columnStartOffset int) ColumnDescriptor {
	return ColumnDescriptor {

	}
}

// GetTableContext returns the table context.
// References "Table Context".
func (pstFile *File) GetTableContext(btreeNodeEntryHeapOnNode BTreeNodeEntry, formatType string) error {
	if btreeNodeEntryHeapOnNode.GetTableType() != 124 {
		// Must be Table Context.
		return errors.New("invalid table type, must be table context")
	}

	allocationTableOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(btreeNodeEntryHeapOnNode.GetHIDUserRoot(), btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return err
	}

	tableContext := btreeNodeEntryHeapOnNode.Data[allocationTableOffsets.StartOffset:allocationTableOffsets.EndOffset]

	tableContextSignature := int(binary.LittleEndian.Uint16([]byte{tableContext[0], 0}))

	if tableContextSignature != 124 {
		return errors.New("invalid table context signature")
	}

	tableColumnCount := int(binary.LittleEndian.Uint16([]byte{tableContext[1], 0}))

	tableRowMatrixHNID := int(binary.LittleEndian.Uint16(tableContext[14:18]))

	log.Infof("Table row matrix HNID: %d", tableRowMatrixHNID)

	// Get the columns descriptors.
	var tableColumnDescriptors []ColumnDescriptor

	offset := 22 // The column descriptors start at offset 22.

	if tableColumnCount > 0 {
		for i := 0; i < tableColumnCount; i++ {
			tableColumnDescriptors = append(tableColumnDescriptors, NewColumnDescriptor(tableContext, offset))

			offset = offset + 8
		}
	}

	return nil
}