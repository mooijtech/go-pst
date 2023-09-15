// go-pst is a library for reading Personal Storage Table (.pst) files (written in Go/Golang).
//
// Copyright 2023 Marten Mooij
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pst

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/rotisserie/eris"
)

// TableContext represents the table context.
type TableContext struct {
	Properties [][]Property
	HeapOnNode *HeapOnNode
	File       *File
}

// GetPropertyReader returns reader of the property.
func (tableContext *TableContext) GetPropertyReader(property Property, localDescriptors ...LocalDescriptor) (PropertyReader, error) {
	// TODO - Caller never passes local descriptors?
	return NewPropertyReader(property, tableContext.HeapOnNode, tableContext.File, localDescriptors)
}

// GetTableContext returns the table context.
// If propertyIDsToGet is empty all properties will be returned.
// References "Table Context".
func (file *File) GetTableContext(heapOnNode *HeapOnNode, localDescriptors []LocalDescriptor, propertyIDsToGet ...uint16) (TableContext, error) {
	// TODO - We can merge ReadAt calls into a single one.
	tableType, err := heapOnNode.GetTableType()

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get table type")
	} else if tableType != 124 {
		// Must be Table Context.
		return TableContext{}, ErrTableTypeInvalid
	}

	hidUserRoot, err := heapOnNode.GetHIDUserRoot()

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get HID user root")
	}

	tableContextReader, err := file.GetHeapOnNodeReaderFromHNID(hidUserRoot, *heapOnNode.Reader, localDescriptors...)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get table context reader")
	}

	tableContextSignature := make([]byte, 1)

	if _, err = tableContextReader.ReadAt(tableContextSignature, 0); err != nil {
		return TableContext{}, eris.Wrap(err, "failed to read table context signature")
	} else if tableContextSignature[0] != 124 {
		return TableContext{}, ErrTableSignatureInvalid
	}

	tableColumnCount := make([]byte, 1)

	if _, err = tableContextReader.ReadAt(tableColumnCount, 1); err != nil {
		return TableContext{}, eris.Wrap(err, "failed to read table column count")
	} else if tableColumnCount[0] < 1 {
		return TableContext{}, ErrTableContextNoColumns
	}

	// TCI_1b is the start offset to the column which holds the 1-byte values.
	tci1b := make([]byte, 2)

	if _, err = tableContextReader.ReadAt(tci1b, 6); err != nil {
		return TableContext{}, eris.Wrap(err, "failed to read tci1b")
	}

	rowSize := make([]byte, 2)

	if _, err = tableContextReader.ReadAt(rowSize, 8); err != nil {
		return TableContext{}, eris.Wrap(err, "failed to read row size")
	}

	tableRowMatrixHNID := make([]byte, 4)

	if _, err = tableContextReader.ReadAt(tableRowMatrixHNID, 14); err != nil {
		return TableContext{}, eris.Wrap(err, "failed to read table row matrix HNID")
	} else if binary.LittleEndian.Uint32(tableRowMatrixHNID) == 0 {
		return TableContext{}, ErrTableContextNoRows
	}

	// This is a row matrix, having all its elements in a single row.
	tableRowMatrixReader, err := file.GetHeapOnNodeReaderFromHNID(Identifier(binary.LittleEndian.Uint32(tableRowMatrixHNID)), *heapOnNode.Reader, localDescriptors...)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get table row matrix reader")
	}

	// Get the columns descriptors.
	tableColumnDescriptors := make([]ColumnDescriptor, tableColumnCount[0])

	columnDescriptorsOffset := int64(22) // The column descriptors start at offset 22.
	columnIndexesToGet := make([]int, 0, len(propertyIDsToGet))

	for i := 0; i < int(tableColumnCount[0]); i++ {
		columnDescriptor, err := NewColumnDescriptor(tableContextReader, columnDescriptorsOffset)

		if err != nil {
			return TableContext{}, eris.Wrap(err, "failed to create column descriptor")
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
		return TableContext{}, eris.Wrap(err, "failed to get block size")
	}

	blockTrailerSize, err := file.GetBlockTrailerSize()

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get block trailer size")
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
			return TableContext{}, eris.Wrap(err, "failed to read cell existence block")
		}

		for _, columnIndex := range columnIndexesToGet {
			if cellExistenceBlock[tableColumnDescriptors[columnIndex].CellExistenceBitmapIndex/8]&(1<<(7-(tableColumnDescriptors[columnIndex].CellExistenceBitmapIndex%8))) == 0 {
				continue
			}

			property, err := file.GetTableContextProperty(tableRowMatrixReader, currentRowStartOffset, tableColumnDescriptors[columnIndex])

			if err != nil {
				return TableContext{}, eris.Wrap(err, "failed to get table context property")
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

	property.Identifier = column.PropertyID
	property.Type = column.PropertyType

	// Table Context uses a HNID for any data (PropertyType) exceeding 8 bytes.
	// Otherwise, the data is small enough to fit in the Property directly.
	if property.Type.GetDataSize() != -1 && column.DataSize <= 8 {
		// Single value.
		value := make([]byte, column.DataSize)

		if _, err := tableRowMatrixReader.ReadAt(value, rowOffset+int64(column.DataOffset)); err != nil {
			return Property{}, eris.Wrap(err, "failed to read table context data")
		}

		property.Value = value
	} else {
		// Variable size.
		hnid := make([]byte, 4)

		if _, err := tableRowMatrixReader.ReadAt(hnid, rowOffset+int64(column.DataOffset)); err != nil {
			return Property{}, eris.Wrap(err, "failed to read HNID")
		}

		property.HNID = Identifier(binary.LittleEndian.Uint32(hnid))
	}

	return property, nil
}

// ColumnDescriptor represents a column in the Table Context.
// References:
// - https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcoldesc
type ColumnDescriptor struct {
	PropertyType             PropertyType
	PropertyID               uint16
	DataOffset               uint16
	DataSize                 uint8
	CellExistenceBitmapIndex uint8
}

// NewColumnDescriptor is a constructor for creating column descriptors.
func NewColumnDescriptor(tableContextReader io.ReaderAt, columnStartOffset int64) (ColumnDescriptor, error) {
	columnDescriptor := make([]byte, 8)

	if _, err := tableContextReader.ReadAt(columnDescriptor, columnStartOffset); err != nil {
		return ColumnDescriptor{}, err
	}

	return ColumnDescriptor{
		PropertyType:             PropertyType(binary.LittleEndian.Uint16(columnDescriptor[:2])),
		PropertyID:               binary.LittleEndian.Uint16(columnDescriptor[2 : 2+2]),
		DataOffset:               binary.LittleEndian.Uint16(columnDescriptor[4 : 4+2]),
		DataSize:                 columnDescriptor[6:7][0],
		CellExistenceBitmapIndex: columnDescriptor[7:8][0],
	}, nil
}
