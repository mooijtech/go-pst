// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
)

// ColumnDescriptor represents a column in the Table Context.
// References "Table Context", "Table Context Column Descriptor".
type ColumnDescriptor struct {
	PropertyType int
	PropertyID int
	DataOffset int
	DataSize int
	CellExistenceBitmapIndex int
}

// NewColumnDescriptor is a constructor for creating column descriptors.
func NewColumnDescriptor(tableContext []byte, columnStartOffset int) ColumnDescriptor {
	return ColumnDescriptor {
		PropertyType: int(binary.LittleEndian.Uint16(tableContext[columnStartOffset:columnStartOffset + 2])),
		PropertyID: int(binary.LittleEndian.Uint16(tableContext[columnStartOffset + 2:columnStartOffset + 4])),
		DataOffset: int(binary.LittleEndian.Uint16(tableContext[columnStartOffset + 4:columnStartOffset + 6])),
		DataSize: int(binary.LittleEndian.Uint16([]byte{tableContext[columnStartOffset + 6], 0})),
		CellExistenceBitmapIndex: int(binary.LittleEndian.Uint16([]byte{tableContext[columnStartOffset + 7], 0})),
	}
}

// TableContextItem represents an item within the table context.
type TableContextItem struct {
	Index int
	PropertyType int
	PropertyID int
}

// GetTableContext returns the table context.
// References "Table Context".
func (pstFile *File) GetTableContext(btreeNodeEntryHeapOnNode BTreeNodeEntry, formatType string, startAtRow int, numberOfRowsToReturn int) error {
	if btreeNodeEntryHeapOnNode.GetTableType() != 124 {
		// Must be Table Context.
		return errors.New("invalid table type, must be table context")
	}

	tableContextOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(btreeNodeEntryHeapOnNode.GetHIDUserRoot(), btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return err
	}

	tableContext := btreeNodeEntryHeapOnNode.Data[tableContextOffsets.StartOffset:tableContextOffsets.EndOffset]

	tableContextSignature := int(binary.LittleEndian.Uint16([]byte{tableContext[0], 0}))

	if tableContextSignature != 124 {
		return errors.New("invalid table context signature")
	}

	tableColumnCount := int(binary.LittleEndian.Uint16([]byte{tableContext[1], 0}))

	if tableColumnCount < 1 {
		return errors.New("there are no columns in this table context")
	}

	// TCI_1b is the start offset to the column which holds the 1-byte values.
	tci1b := int(binary.LittleEndian.Uint16(tableContext[6:8]))

	rowSize := int(binary.LittleEndian.Uint16(tableContext[8:10]))

	tableRowMatrixHNID := int(binary.LittleEndian.Uint16(tableContext[14:18]))

	tableRowMatrixOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(tableRowMatrixHNID, btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return err
	}

	tableRowMatrix := btreeNodeEntryHeapOnNode.Data[tableRowMatrixOffsets.StartOffset:tableRowMatrixOffsets.EndOffset]

	log.Infof("Table row matrix HNID: %d", tableRowMatrixHNID)
	log.Infof("Table row matrix: %d", len(tableRowMatrix))

	// Get the columns descriptors.
	tableColumnDescriptors := make([]ColumnDescriptor, tableColumnCount)

	offset := 22 // The column descriptors start at offset 22.

	for i := 0; i < tableColumnCount; i++ {
		tableColumnDescriptors[i] = NewColumnDescriptor(tableContext, offset)

		offset = offset + 8 // Each column descriptor is 8 bytes in size.
	}

	blockSize, err := pstFile.GetBlockSize(formatType)

	if err != nil {
		return err
	}

	blockTrailerSize, err := pstFile.GetBlockTrailerSize(formatType)

	if err != nil {
		return err
	}

	blockCount := len(tableContext) / (blockSize - blockTrailerSize)
	rowsPerBlock := (blockSize - blockTrailerSize) / rowSize
	rowCount := (blockCount * rowsPerBlock) + ((len(tableRowMatrix) % (blockSize - blockTrailerSize)) / rowSize)
	cellExistenceBlockSize := int(math.Ceil(float64(tableColumnCount) / 8))

	log.Infof("Row count: %d", rowCount)

	var currentRowStartOffset int

	for i := 0; i < numberOfRowsToReturn; i++ {
		currentRowStartOffset = (((startAtRow + i) / rowsPerBlock) * (blockSize - blockTrailerSize)) + (((startAtRow + i) % rowsPerBlock) * rowSize)

		cellExistenceBlock := tableRowMatrix[currentRowStartOffset + tci1b:currentRowStartOffset + tci1b + cellExistenceBlockSize]

		for i := 0; i < tableColumnCount; i++ {
			column := tableColumnDescriptors[i]

			// Check if this column exists.
			if cellExistenceBlock[column.CellExistenceBitmapIndex / 8] & (0x01 << (7 - (column.CellExistenceBitmapIndex % 8))) == 0 {
				continue
			}

			switch column.DataSize {
			default:
				// Four bytes data
				referenceHNID := int(binary.LittleEndian.Uint32(tableRowMatrix[currentRowStartOffset + column.DataOffset:currentRowStartOffset + column.DataOffset + 4]))

				if column.PropertyType == PropertyTypeInteger32 || column.PropertyType == PropertyTypeFloating32 {
					// 32-bit data
				}

				dataOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(referenceHNID, btreeNodeEntryHeapOnNode, formatType)

				if err != nil {
					return err
				}

				if fmt.Sprintf("%x", column.PropertyID) == "3001" {
					data := btreeNodeEntryHeapOnNode.Data[dataOffsets.StartOffset:dataOffsets.EndOffset]

					log.Infof("Root folder display name: %s", string(data))
				}
			}

		}
	}

	return nil
}