// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
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
		PropertyType:             int(binary.LittleEndian.Uint16(tableContext[columnStartOffset:columnStartOffset + 2])),
		PropertyID:               int(binary.LittleEndian.Uint16(tableContext[columnStartOffset + 2:columnStartOffset + 4])),
		DataOffset:               int(binary.LittleEndian.Uint16(tableContext[columnStartOffset + 4:columnStartOffset + 6])),
		DataSize:                 int(binary.LittleEndian.Uint16([]byte{tableContext[columnStartOffset + 6], 0})),
		CellExistenceBitmapIndex: int(binary.LittleEndian.Uint16([]byte{tableContext[columnStartOffset + 7], 0})),
	}
}

// TableContextItem represents an item within the table context.
type TableContextItem struct {
	PropertyType int
	PropertyID int
	ReferenceHNID int
	IsExternalValueReference bool
	Data []byte
}

// GetTableContext returns the table context.
// The number of rows to return may be -1 to return all rows.
// References "Table Context".
func (pstFile *File) GetTableContext(btreeNodeEntryHeapOnNode BTreeNodeEntry, localDescriptors []LocalDescriptor, formatType string, startAtRow int, numberOfRowsToReturn int, columnToGet int) ([][]TableContextItem, error) {
	if btreeNodeEntryHeapOnNode.GetTableType() != 124 {
		// Must be Table Context.
		return nil, errors.New("invalid table type, must be table context")
	}

	tableContextOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(btreeNodeEntryHeapOnNode.GetHIDUserRoot(), btreeNodeEntryHeapOnNode, localDescriptors, formatType)

	if err != nil {
		return nil, err
	}

	tableContext := btreeNodeEntryHeapOnNode.Data[tableContextOffsets.StartOffset:tableContextOffsets.EndOffset]

	tableContextSignature := int(binary.LittleEndian.Uint16([]byte{tableContext[0], 0}))

	if tableContextSignature != 124 {
		return nil, errors.New("invalid table context signature")
	}

	tableColumnCount := int(binary.LittleEndian.Uint16([]byte{tableContext[1], 0}))

	if tableColumnCount < 1 {
		return nil, errors.New("there are no columns in this table context")
	}

	// TCI_1b is the start offset to the column which holds the 1-byte values.
	tci1b := int(binary.LittleEndian.Uint16(tableContext[6:8]))

	rowSize := int(binary.LittleEndian.Uint16(tableContext[8:10]))

	tableRowMatrixHNID := int(binary.LittleEndian.Uint32(tableContext[14:18]))

	tableRowMatrixOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(tableRowMatrixHNID, btreeNodeEntryHeapOnNode, localDescriptors, formatType)

	if err != nil {
		return nil, err
	}

	if len(tableRowMatrixOffsets.Data) > 0 {
		btreeNodeEntryHeapOnNode.Data = tableRowMatrixOffsets.Data
	}

	tableRowMatrix := btreeNodeEntryHeapOnNode.Data[tableRowMatrixOffsets.StartOffset:tableRowMatrixOffsets.EndOffset]

	// Get the columns descriptors.
	tableColumnDescriptors := make([]ColumnDescriptor, tableColumnCount)

	offset := 22 // The column descriptors start at offset 22.
	var columnToGetIndex int

	for i := 0; i < tableColumnCount; i++ {
		tableColumnDescriptors[i] = NewColumnDescriptor(tableContext, offset)

		if tableColumnDescriptors[i].PropertyID == columnToGet {
			columnToGetIndex = i
		}

		offset = offset + 8 // Each column descriptor is 8 bytes in size.
	}

	blockSize, err := pstFile.GetBlockSize(formatType)

	if err != nil {
		return nil, err
	}

	blockTrailerSize, err := pstFile.GetBlockTrailerSize(formatType)

	if err != nil {
		return nil, err
	}

	blockCount := len(tableContext) / (blockSize - blockTrailerSize)
	rowsPerBlock := (blockSize - blockTrailerSize) / rowSize
	rowCount := (blockCount * rowsPerBlock) + ((len(tableRowMatrix) % (blockSize - blockTrailerSize)) / rowSize)
	cellExistenceBlockSize := int(math.Ceil(float64(tableColumnCount) / 8))

	if startAtRow == -1 || numberOfRowsToReturn == -1 {
		numberOfRowsToReturn = rowCount
		startAtRow = 0
	}

	var currentRowStartOffset int
	tableContextItems := make([][]TableContextItem, numberOfRowsToReturn)

	for i := 0; i < numberOfRowsToReturn; i++ {
		currentRowStartOffset = (((startAtRow + i) / rowsPerBlock) * (blockSize - blockTrailerSize)) + (((startAtRow + i) % rowsPerBlock) * rowSize)

		cellExistenceBlock := tableRowMatrix[currentRowStartOffset + tci1b:currentRowStartOffset + tci1b + cellExistenceBlockSize]

		x := 0

		if columnToGet > -1 {
			x = columnToGetIndex
		}

		for x < tableColumnCount {
			column := tableColumnDescriptors[x]

			if cellExistenceBlock[column.CellExistenceBitmapIndex / 8] & (1 << (7 - (column.CellExistenceBitmapIndex % 8))) == 0 {
				x += 1
				continue
			}

			var tableContextItem TableContextItem

			tableContextItem.PropertyID = column.PropertyID
			tableContextItem.PropertyType = column.PropertyType

			switch column.DataSize {
			case 1:
				// 1 byte data
				tableContextItem.ReferenceHNID = int(binary.LittleEndian.Uint16([]byte{tableRowMatrix[currentRowStartOffset + column.DataOffset], 0}))

				tableContextItem.IsExternalValueReference = true
				break
			case 2:
				// 2 byte data
				tableContextItem.ReferenceHNID = int(binary.LittleEndian.Uint16(tableRowMatrix[currentRowStartOffset + column.DataOffset:currentRowStartOffset + column.DataOffset + 2]))

				tableContextItem.IsExternalValueReference = true
				break
			case 8:
				// 8 byte data
				tableContextItem.Data = tableRowMatrix[currentRowStartOffset + column.DataOffset:currentRowStartOffset + column.DataOffset + 8]
				break
			default:
				// 4 byte data
				tableContextItem.ReferenceHNID = int(binary.LittleEndian.Uint32(tableRowMatrix[currentRowStartOffset + column.DataOffset:currentRowStartOffset + column.DataOffset + 4]))

				if column.PropertyType == PropertyTypeInteger32 || column.PropertyType == PropertyTypeFloating32 {
					// 32-bit data
					tableContextItem.IsExternalValueReference = true
					break
				}

				if (tableContextItem.ReferenceHNID & 0x1F) != 0 {
					tableContextItem.IsExternalValueReference = true
					break
				}

				dataOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(tableContextItem.ReferenceHNID, btreeNodeEntryHeapOnNode, localDescriptors, formatType)

				if err != nil {
					return nil, err
				}

				tableContextItem.Data = btreeNodeEntryHeapOnNode.Data[dataOffsets.StartOffset:dataOffsets.EndOffset]
			}

			tableContextItems[i] = append(tableContextItems[i], tableContextItem)
			x += 1
		}
	}

	return tableContextItems, nil
}