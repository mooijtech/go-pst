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

package writer

import (
	"bytes"
	"cmp"
	"encoding/binary"
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/proto"
	"io"
	"math"
	"slices"
)

// TableContextWriter represents a writer for a pst.TableContext.
type TableContextWriter struct {
	// FormatType represents the FormatType.
	FormatType pst.FormatType
	// Properties represents the properties to write (properties.Attachment, properties.Folder etc).
	Properties proto.Message
	// BTreeOnHeapWriter represents the BTreeOnHeapWriter.
	BTreeOnHeapWriter *BTreeOnHeapWriter
	// PropertiesWriter represents the PropertiesWriter.
	PropertiesWriter *PropertiesWriter
}

// NewTableContextWriter creates a new TableContextWriter.
func NewTableContextWriter(formatType pst.FormatType, properties proto.Message) *TableContextWriter {
	heapOnNodeWriter := NewHeapOnNodeWriter(pst.SignatureTypeTableContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)
	propertiesWriter := NewPropertiesWriter(properties)

	return &TableContextWriter{
		FormatType:        formatType,
		Properties:        properties,
		BTreeOnHeapWriter: btreeOnHeapWriter,
		PropertiesWriter:  propertiesWriter,
	}
}

// WriteTo writes the pst.TableContext.
// References:
// - https://github.com/mooijtech/go-pst/blob/main/docs/README.md#creating-a-tc
func (tableContextWriter *TableContextWriter) WriteTo(writer io.Writer) (int64, error) {
	btreeOnHeapWrittenSize, err := tableContextWriter.BTreeOnHeapWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write BTree-on-Heap")
	}

	properties, err := tableContextWriter.PropertiesWriter.GetProperties()

	if err != nil {
		return 0, eris.Wrap(err, "failed to get properties")
	}

	headerWrittenSize, err := tableContextWriter.WriteHeader(writer, properties)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Table Context header")
	}

	rowMatrixWrittenSize, err := tableContextWriter.WriteRowMatrix(writer, properties)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write row matrix")
	}

	return btreeOnHeapWrittenSize + headerWrittenSize + rowMatrixWrittenSize, nil
}

// WriteHeader writes the pst.TableContext header.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcinfo
func (tableContextWriter *TableContextWriter) WriteHeader(writer io.Writer, properties []Property) (int64, error) {
	columnDescriptors, err := tableContextWriter.GetColumnDescriptors(properties)

	if err != nil {
		return 0, eris.Wrap(err, "failed to get column descriptors")
	}

	// 1+1+8+4+4+4+columnDescriptors
	header := bytes.NewBuffer(make([]byte, 22+(8*len(columnDescriptors))))

	// MUST be set to bTypeTC.
	header.Write([]byte{byte(pst.SignatureTypeTableContext)})

	// Column count.
	header.WriteByte(byte(len(columnDescriptors)))

	// TODO - Pass from row matrix
	// This is an array of 4 16-bit values that specify the offsets of various groups of data in the actual row data.
	header.Write(make([]byte, 2)) // Ending offset of 8- and 4-byte data value group.
	header.Write(make([]byte, 2)) // Ending offset of 2-byte data value group.
	header.Write(make([]byte, 2)) // Ending offset of 1-byte data value group.
	header.Write(make([]byte, 2)) // Ending offset of the Cell Existence Block.

	// HID to the Row ID BTH (hidRowIndex).
	// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcrowid
	// TODO
	header.Write(make([]byte, 4))

	// HNID to the Row Matrix (that is, actual table data). (hnidRows)
	// TODO-
	header.Write(make([]byte, 4))

	// Deprecated (hidIndex)
	header.Write(make([]byte, 4))

	// Sort column descriptors by property ID (according to specification).
	slices.SortFunc(columnDescriptors, func(a []byte, b []byte) int {
		return cmp.Compare(binary.LittleEndian.Uint16(a[2:2+2]), binary.LittleEndian.Uint16(b[2:2+2]))
	})

	// Array of Column Descriptors.
	for _, columnDescriptor := range columnDescriptors {
		header.Write(columnDescriptor)
	}

	return header.WriteTo(writer)
}

// GetColumnDescriptors returns the Column Descriptors based on the properties.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcoldesc
func (tableContextWriter *TableContextWriter) GetColumnDescriptors(properties []Property) ([][]byte, error) {
	var columnDescriptors [][]byte

	for i, property := range properties {
		columnDescriptorBuffer := bytes.NewBuffer(make([]byte, 8))

		// Tag
		columnDescriptorBuffer.Write(GetUint16(uint16(property.ID)))
		columnDescriptorBuffer.Write(GetUint16(uint16(property.Type)))
		// Offset
		columnDescriptorBuffer.Write(GetUint16(uint16(8 * i)))

		// Data Size
		if property.Value.Len() <= 8 {
			columnDescriptorBuffer.WriteByte(byte(property.Value.Len()))
		} else {
			// Variable-sized data (size of a HNID)
			columnDescriptorBuffer.WriteByte(byte(4))
		}

		// Cell Existence Bitmap Index
		columnDescriptorBuffer.WriteByte(byte(2 + i)) // Skip 0 and 1
	}

	return columnDescriptors, nil
}

// WriteRowMatrix writes the Row Matrix of the Table Context.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#row-matrix
func (tableContextWriter *TableContextWriter) WriteRowMatrix(writer io.Writer, properties []Property) (int64, error) {
	rowMatrixBuffer := bytes.NewBuffer(make([]byte, tableContextWriter.GetRowMatrixSize(properties)))

	// Sort by byte-size so writes are aligned accordingly.
	slices.SortFunc(properties, func(a, b Property) int {
		return cmp.Compare(a.Value.Len(), b.Value.Len())
	})

	for _, property := range properties {
		// The 32-bit value that corresponds to the dwRowID value in this row's corresponding TCROWID record.
		// Note that this value corresponds to the PidTagLtpRowId property.
		// dwRowID TODO
		rowMatrixBuffer.Write(make([]byte, 4))

		// Already sorted by byte-size.
		if _, err := property.Value.WriteTo(rowMatrixBuffer); err != nil {
			return 0, eris.Wrap(err, "failed to write property value")
		}

		// Cell Existence Block.
		rowMatrixBuffer.Write(make([]byte, 0))
	}

	return rowMatrixBuffer.WriteTo(writer)
}

// GetRowMatrixSize returns the total size of the Row Matrix based on the properties to write.
func (tableContextWriter *TableContextWriter) GetRowMatrixSize(properties []Property) int {
	totalRowIDSize := len(properties) * 4
	totalCellExistenceBlockSize := len(properties) * int(math.Ceil(float64(len(properties))/8))

	totalSize := totalRowIDSize + totalCellExistenceBlockSize

	for _, property := range properties {
		totalSize += property.Value.Len()
	}

	return totalSize
}
