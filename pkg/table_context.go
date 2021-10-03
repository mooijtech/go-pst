// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"errors"
	"math"
)

// ColumnDescriptor represents a column in the Table Context.
// References "Table Context", "Table Context Column Descriptor".
type ColumnDescriptor struct {
	PropertyType             int
	PropertyID               int
	DataOffset               int
	DataSize                 int
	CellExistenceBitmapIndex int
}

// NewColumnDescriptor is a constructor for creating column descriptors.
func NewColumnDescriptor(tableContextNodeInputStream NodeInputStream, columnStartOffset int) (ColumnDescriptor, error) {
	propertyType, err := tableContextNodeInputStream.SeekAndReadUint16(2, columnStartOffset)

	if err != nil {
		return ColumnDescriptor{}, err
	}

	propertyID, err := tableContextNodeInputStream.SeekAndReadUint16(2, columnStartOffset+2)

	if err != nil {
		return ColumnDescriptor{}, err
	}

	dataOffset, err := tableContextNodeInputStream.SeekAndReadUint16(2, columnStartOffset+4)

	if err != nil {
		return ColumnDescriptor{}, err
	}

	dataSize, err := tableContextNodeInputStream.SeekAndReadUint16(1, columnStartOffset+6)

	if err != nil {
		return ColumnDescriptor{}, err
	}

	cellExistenceBitmapIndex, err := tableContextNodeInputStream.SeekAndReadUint16(1, columnStartOffset+7)

	if err != nil {
		return ColumnDescriptor{}, err
	}

	return ColumnDescriptor{
		PropertyType:             propertyType,
		PropertyID:               propertyID,
		DataOffset:               dataOffset,
		DataSize:                 dataSize,
		CellExistenceBitmapIndex: cellExistenceBitmapIndex,
	}, nil
}

// TableContextItem represents an item within the table context.
type TableContextItem struct {
	PropertyType             int
	PropertyID               int
	ReferenceHNID            int
	IsExternalValueReference bool
	Data                     []byte
}

// GetTableContext returns the table context.
// The number of rows to return may be -1 to return all rows.
// References "Table Context".
func (pstFile *File) GetTableContext(heapOnNode HeapOnNode, localDescriptors []LocalDescriptor, formatType string, encryptionType string, startAtRow int, numberOfRowsToReturn int, columnToGet int) ([][]TableContextItem, error) {
	tableType, err := heapOnNode.GetTableType()

	if err != nil {
		return nil, err
	}

	if tableType != 124 {
		// Must be Table Context.
		return nil, errors.New("invalid table type, must be table context")
	}

	hidUserRoot, err := heapOnNode.GetHIDUserRoot()

	if err != nil {
		return nil, err
	}

	tableContextNodeInputStream, err := pstFile.GetAllocationTableNodeInputStream(hidUserRoot, heapOnNode, localDescriptors, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	tableContextSignature, err := tableContextNodeInputStream.SeekAndReadUint16(1, 0)

	if err != nil {
		return nil, err
	}

	if tableContextSignature != 124 {
		return nil, errors.New("invalid table context signature")
	}

	tableColumnCount, err := tableContextNodeInputStream.SeekAndReadUint16(1, 1)

	if err != nil {
		return nil, err
	}

	if tableColumnCount < 1 {
		return nil, errors.New("there are no columns in this table context")
	}

	// TCI_1b is the start offset to the column which holds the 1-byte values.
	tci1b, err := tableContextNodeInputStream.SeekAndReadUint16(2, 6)

	if err != nil {
		return nil, err
	}

	rowSize, err := tableContextNodeInputStream.SeekAndReadUint16(2, 8)

	if err != nil {
		return nil, err
	}

	tableRowMatrixHNID, err := tableContextNodeInputStream.SeekAndReadUint32(4, 14)

	if err != nil {
		return nil, err
	}

	tableRowMatrixNodeInputStream, err := pstFile.GetAllocationTableNodeInputStream(tableRowMatrixHNID, heapOnNode, localDescriptors, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	// Get the columns descriptors.
	tableColumnDescriptors := make([]ColumnDescriptor, tableColumnCount)

	offset := 22 // The column descriptors start at offset 22.
	var columnToGetIndex int

	for i := 0; i < tableColumnCount; i++ {
		columnDescriptor, err := NewColumnDescriptor(tableContextNodeInputStream, offset)

		if err != nil {
			return nil, err
		}

		tableColumnDescriptors[i] = columnDescriptor

		if columnDescriptor.PropertyID == columnToGet {
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

	blockCount := tableContextNodeInputStream.Size / (blockSize - blockTrailerSize)
	rowsPerBlock := (blockSize - blockTrailerSize) / rowSize
	rowCount := (blockCount * rowsPerBlock) + ((tableRowMatrixNodeInputStream.Size % (blockSize - blockTrailerSize)) / rowSize)
	cellExistenceBlockSize := int(math.Ceil(float64(tableColumnCount) / 8))

	if startAtRow == -1 || numberOfRowsToReturn == -1 {
		numberOfRowsToReturn = rowCount
		startAtRow = 0
	}

	var currentRowStartOffset int
	tableContextItems := make([][]TableContextItem, numberOfRowsToReturn)

	for i := 0; i < numberOfRowsToReturn; i++ {
		currentRowStartOffset = (((startAtRow + i) / rowsPerBlock) * (blockSize - blockTrailerSize)) + (((startAtRow + i) % rowsPerBlock) * rowSize)

		cellExistenceBlock, err := tableRowMatrixNodeInputStream.Read(cellExistenceBlockSize, currentRowStartOffset+tci1b)

		if err != nil {
			return nil, err
		}

		x := 0

		if columnToGet > -1 {
			x = columnToGetIndex
		}

		for x < tableColumnCount {
			column := tableColumnDescriptors[x]

			if cellExistenceBlock[column.CellExistenceBitmapIndex/8]&(1<<(7-(column.CellExistenceBitmapIndex%8))) == 0 {
				x += 1
				continue
			}

			var tableContextItem TableContextItem

			tableContextItem.PropertyID = column.PropertyID
			tableContextItem.PropertyType = column.PropertyType

			switch column.DataSize {
			case 1:
				// 1 byte data
				referenceHNID, err := tableRowMatrixNodeInputStream.SeekAndReadUint16(1, currentRowStartOffset+column.DataOffset)

				if err != nil {
					return nil, err
				}

				tableContextItem.ReferenceHNID = referenceHNID
				tableContextItem.IsExternalValueReference = true
				break
			case 2:
				// 2 byte data
				referenceHNID, err := tableRowMatrixNodeInputStream.SeekAndReadUint16(2, currentRowStartOffset+column.DataOffset)

				if err != nil {
					return nil, err
				}

				tableContextItem.ReferenceHNID = referenceHNID
				tableContextItem.IsExternalValueReference = true
				break
			case 8:
				// 8 byte data
				data, err := tableRowMatrixNodeInputStream.Read(8, currentRowStartOffset+column.DataOffset)

				if err != nil {
					return nil, err
				}

				tableContextItem.Data = data
				break
			default:
				// 4 byte data
				referenceHNID, err := tableRowMatrixNodeInputStream.SeekAndReadUint32(4, currentRowStartOffset+column.DataOffset)

				if err != nil {
					return nil, err
				}

				tableContextItem.ReferenceHNID = referenceHNID

				if column.PropertyType == PropertyTypeInteger32 || column.PropertyType == PropertyTypeFloating32 {
					// 32-bit data
					tableContextItem.IsExternalValueReference = true
					break
				}

				if (referenceHNID & 0x1F) != 0 {
					tableContextItem.IsExternalValueReference = true
					break
				}

				tableContextItemNodeInputStream, err := pstFile.GetAllocationTableNodeInputStream(tableContextItem.ReferenceHNID, heapOnNode, localDescriptors, formatType, encryptionType)

				if err != nil {
					return nil, err
				}

				tableContextItemData, err := tableContextItemNodeInputStream.Read(tableContextItemNodeInputStream.Size, 0)

				tableContextItem.Data = tableContextItemData
			}

			tableContextItems[i] = append(tableContextItems[i], tableContextItem)
			x += 1
		}
	}

	return tableContextItems, nil
}
