// go-pst is a library for reading Personal Storage Table (.pst) files (written in Go/Golang).
//
// Copyright (C) 2022  Marten Mooij
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package pst

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/pkg/errors"
)

// TableContext represents the table context.
type TableContext struct {
	Properties [][]Property
	HeapOnNode *HeapOnNode
	File       *File
}

// GetPropertyReader returns reader of the property.
func (tableContext *TableContext) GetPropertyReader(property Property, localDescriptors ...LocalDescriptor) (PropertyReader, error) {
	return NewPropertyReader(property, tableContext.HeapOnNode, tableContext.File, localDescriptors...)
}

// GetTableContext returns the table context.
// If propertyIDsToGet is empty all properties will be returned.
// References "Table Context".
func (file *File) GetTableContext(heapOnNode *HeapOnNode, localDescriptors []LocalDescriptor, propertyIDsToGet ...uint16) (TableContext, error) {
	tableType, err := heapOnNode.GetTableType()

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	} else if tableType != 124 {
		// Must be Table Context.
		return TableContext{}, errors.WithStack(ErrTableTypeInvalid)
	}

	hidUserRoot, err := heapOnNode.GetHIDUserRoot()

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	tableContextReader, err := file.GetHeapOnNodeReaderFromHNID(hidUserRoot, *heapOnNode.Reader, localDescriptors...)

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	tableContextSignature := make([]byte, 1)

	if _, err = tableContextReader.ReadAt(tableContextSignature, 0); err != nil {
		return TableContext{}, errors.WithStack(err)
	} else if tableContextSignature[0] != 124 {
		return TableContext{}, errors.WithStack(ErrTableSignatureInvalid)
	}

	tableColumnCount := make([]byte, 1)

	if _, err = tableContextReader.ReadAt(tableColumnCount, 1); err != nil {
		return TableContext{}, errors.WithStack(err)
	} else if tableColumnCount[0] < 1 {
		return TableContext{}, errors.WithStack(ErrTableContextNoColumns)
	}

	// TCI_1b is the start offset to the column which holds the 1-byte values.
	tci1b := make([]byte, 2)

	if _, err = tableContextReader.ReadAt(tci1b, 6); err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	rowSize := make([]byte, 2)

	if _, err = tableContextReader.ReadAt(rowSize, 8); err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	tableRowMatrixHNID := make([]byte, 4)

	if _, err = tableContextReader.ReadAt(tableRowMatrixHNID, 14); err != nil {
		return TableContext{}, errors.WithStack(err)
	} else if binary.LittleEndian.Uint32(tableRowMatrixHNID) == 0 {
		return TableContext{}, errors.WithStack(ErrTableContextNoRows)
	}

	// This is a row matrix, having all its elements in a single row.
	tableRowMatrixReader, err := file.GetHeapOnNodeReaderFromHNID(Identifier(binary.LittleEndian.Uint32(tableRowMatrixHNID)), *heapOnNode.Reader, localDescriptors...)

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	// Get the columns descriptors.
	tableColumnDescriptors := make([]ColumnDescriptor, tableColumnCount[0])

	columnDescriptorsOffset := int64(22) // The column descriptors start at offset 22.
	columnIndexesToGet := make([]int, 0, len(propertyIDsToGet))

	for i := 0; i < int(tableColumnCount[0]); i++ {
		columnDescriptor, err := NewColumnDescriptor(tableContextReader, columnDescriptorsOffset)

		if err != nil {
			return TableContext{}, errors.WithStack(err)
		}

		tableColumnDescriptors[i] = columnDescriptor

		for _, propertyIDToGet := range propertyIDsToGet {
			if columnDescriptor.PropertyID == propertyIDToGet {
				columnIndexesToGet = append(columnIndexesToGet, i)
				break
			}
		}

		columnDescriptorsOffset += 8 // Each column descriptor is 8 bytes in size.
	}

	if len(propertyIDsToGet) == 0 {
		for i := 0; i < int(tableColumnCount[0]); i++ {
			columnIndexesToGet = append(columnIndexesToGet, i)
		}
	}

	blockSize, err := file.GetBlockSize()

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	blockTrailerSize, err := file.GetBlockTrailerSize()

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	blockCount := int(tableRowMatrixReader.Size()) / (blockSize - blockTrailerSize)
	rowsPerBlock := (blockSize - blockTrailerSize) / int(binary.LittleEndian.Uint16(rowSize))

	rowCount := (blockCount * rowsPerBlock) + ((int(tableRowMatrixReader.Size()) % (blockSize - blockTrailerSize)) / int(binary.LittleEndian.Uint16(rowSize)))
	cellExistenceBlockSize := int(math.Ceil(float64(tableColumnCount[0]) / 8))

	startAtRow := 0
	numberOfRowsToReturn := rowCount

	var currentRowStartOffset int64
	tableContextItems := make([][]Property, numberOfRowsToReturn)

	for rowIndex := 0; rowIndex < numberOfRowsToReturn; rowIndex++ {
		currentRowStartOffset = int64((((startAtRow + rowIndex) / rowsPerBlock) * (blockSize - blockTrailerSize)) + (((startAtRow + rowIndex) % rowsPerBlock) * int(binary.LittleEndian.Uint16(rowSize))))
		cellExistenceBlock := make([]byte, cellExistenceBlockSize)

		if _, err := tableRowMatrixReader.ReadAt(cellExistenceBlock, currentRowStartOffset+int64(binary.LittleEndian.Uint16(tci1b))); err != nil {
			return TableContext{}, errors.WithStack(err)
		}

		for _, columnIndex := range columnIndexesToGet {
			if cellExistenceBlock[tableColumnDescriptors[columnIndex].CellExistenceBitmapIndex/8]&(1<<(7-(tableColumnDescriptors[columnIndex].CellExistenceBitmapIndex%8))) == 0 {
				continue
			}

			property, err := file.GetTableContextProperty(tableRowMatrixReader, currentRowStartOffset, tableColumnDescriptors[columnIndex])

			if err != nil {
				return TableContext{}, errors.WithStack(err)
			}

			tableContextItems[rowIndex] = append(tableContextItems[rowIndex], property)
		}
	}

	return TableContext{
		Properties: tableContextItems,
		HeapOnNode: heapOnNode,
		File:       file,
	}, nil
}

// GetTableContextProperty is used by GetTableContext to only returns certain columns.
// References [MS-PDF]: 2.3.4.4.1 Row Data Format
func (file *File) GetTableContextProperty(tableRowMatrixReader io.ReaderAt, rowOffset int64, column ColumnDescriptor) (Property, error) {
	var property Property

	property.ID = column.PropertyID
	property.Type = column.PropertyType

	// Table Context uses a HNID for any data (PropertyType) exceeding 8 bytes.
	// Otherwise, the data is small enough to fit in the Property directly.
	if property.Type.GetDataSize() != -1 && column.DataSize <= 8 {
		// Single value.
		data := make([]byte, column.DataSize)

		if _, err := tableRowMatrixReader.ReadAt(data, rowOffset+int64(column.DataOffset)); err != nil {
			return Property{}, errors.WithStack(err)
		}

		property.Data = data
	} else {
		// Variable size.
		hnid := make([]byte, 4)

		if _, err := tableRowMatrixReader.ReadAt(hnid, rowOffset+int64(column.DataOffset)); err != nil {
			return Property{}, errors.WithStack(err)
		}

		property.HNID = Identifier(binary.LittleEndian.Uint32(hnid))
	}

	return property, nil
}

// ColumnDescriptor represents a column in the Table Context.
// References "Table Context", "Table Context Column Descriptor".
type ColumnDescriptor struct {
	PropertyType             PropertyType
	PropertyID               uint16
	DataOffset               uint16
	DataSize                 uint8
	CellExistenceBitmapIndex uint8
}

// NewColumnDescriptor is a constructor for creating column descriptors.
func NewColumnDescriptor(tableContextReader io.ReaderAt, columnStartOffset int64) (ColumnDescriptor, error) {
	propertyType := make([]byte, 2)

	if _, err := tableContextReader.ReadAt(propertyType, columnStartOffset); err != nil {
		return ColumnDescriptor{}, errors.WithStack(err)
	}

	propertyID := make([]byte, 2)

	if _, err := tableContextReader.ReadAt(propertyID, columnStartOffset+2); err != nil {
		return ColumnDescriptor{}, errors.WithStack(err)
	}

	dataOffset := make([]byte, 2)

	if _, err := tableContextReader.ReadAt(dataOffset, columnStartOffset+4); err != nil {
		return ColumnDescriptor{}, errors.WithStack(err)
	}

	dataSize := make([]byte, 1)

	if _, err := tableContextReader.ReadAt(dataSize, columnStartOffset+6); err != nil {
		return ColumnDescriptor{}, errors.WithStack(err)
	}

	cellExistenceBitmapIndex := make([]byte, 1)

	if _, err := tableContextReader.ReadAt(cellExistenceBitmapIndex, columnStartOffset+7); err != nil {
		return ColumnDescriptor{}, errors.WithStack(err)
	}

	return ColumnDescriptor{
		PropertyType:             PropertyType(binary.LittleEndian.Uint16(propertyType)),
		PropertyID:               binary.LittleEndian.Uint16(propertyID),
		DataOffset:               binary.LittleEndian.Uint16(dataOffset),
		DataSize:                 dataSize[0],
		CellExistenceBitmapIndex: cellExistenceBitmapIndex[0],
	}, nil
}
