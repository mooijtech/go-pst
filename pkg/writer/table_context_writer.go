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
	"fmt"
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/proto"
	"io"
	"reflect"
	"slices"
	"strconv"
)

// TableContextWriter represents a writer for a pst.TableContext.
type TableContextWriter struct {
	// BTreeOnHeapWriter represents the BTreeOnHeapWriter.
	BTreeOnHeapWriter *BTreeOnHeapWriter
	// Properties represents the properties to write.
	Properties proto.Message
}

// NewTableContextWriter creates a new TableContextWriter.
func NewTableContextWriter(properties proto.Message) *TableContextWriter {
	heapOnNodeWriter := NewHeapOnNodeWriter(pst.SignatureTypeTableContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)

	return &TableContextWriter{
		BTreeOnHeapWriter: btreeOnHeapWriter,
		Properties:        properties,
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

	// TODO - Make column descriptors from properties.
	headerWrittenSize, err := tableContextWriter.WriteHeader(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Table Context header")
	}

	return btreeOnHeapWrittenSize + headerWrittenSize, nil
}

// WriteColumnDescriptor writes the pst.ColumnDescriptor.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcoldesc
func (tableContextWriter *TableContextWriter) WriteColumnDescriptor(writer io.Writer, columnDescriptor pst.ColumnDescriptor) (int64, error) {
	columnDescriptorBuffer := bytes.NewBuffer(make([]byte, 8))

	columnDescriptorBuffer.Write(GetUint16(columnDescriptor.PropertyID))
	columnDescriptorBuffer.Write(GetUint16(columnDescriptor.DataOffset))
	columnDescriptorBuffer.WriteByte(columnDescriptor.DataSize)
	columnDescriptorBuffer.WriteByte(columnDescriptor.CellExistenceBitmapIndex)

	return columnDescriptorBuffer.WriteTo(writer)
}

// GetColumnDescriptors returns the Column Descriptors based on the properties.
func (tableContextWriter *TableContextWriter) GetColumnDescriptors() ([][]byte, error) {
	var columnDescriptors [][]byte

	propertyTypes := reflect.TypeOf(tableContextWriter.Properties).Elem()

	for i := 0; i < propertyTypes.NumField(); i++ {
		propertyField := propertyTypes.Field(i)

		if !propertyField.IsExported() {
			continue
		}

		tagPropertyID, ok := propertyField.Tag.Lookup("msg")

		if !ok {
			// No property ID in the tag.
			fmt.Printf("Skipping property without ID: %s\n", propertyField.Name)
			continue
		}

		tagPropertyType, ok := propertyField.Tag.Lookup("type")

		if !ok {
			fmt.Printf("Skipping property without type: %s\n", propertyField.Name)
			continue
		}

		propertyID, err := strconv.Atoi(tagPropertyID)

		if err != nil {
			return nil, eris.Wrap(err, "failed to convert propertyID to int")
		}

		propertyType, err := strconv.Atoi(tagPropertyType)

		if err != nil {
			return nil, eris.Wrap(err, "failed to convert propertyType to int")
		}

		// Write column descriptor.
		columnDescriptorBuffer := bytes.NewBuffer(make([]byte, 8))

		columnDescriptorBuffer.Write(GetUint16(uint16(propertyID)))
		columnDescriptorBuffer.Write(GetUint16(uint16(propertyType)))
		// TODO - Offset
		// TODO - Size
		// TODO - CellExistence
	}

	return columnDescriptors, nil
}

// WriteHeader writes the pst.TableContext header.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcinfo
func (tableContextWriter *TableContextWriter) WriteHeader(writer io.Writer) (int64, error) {
	columnDescriptors, err := tableContextWriter.GetColumnDescriptors()

	if err != nil {
		return 0, eris.Wrap(err, "failed to get column descriptors")
	}

	// 1+1+8+4+4+4+columnDescriptors
	header := bytes.NewBuffer(make([]byte, 22+(8*len(columnDescriptors))))

	// MUST be set to bTypeTC.
	header.Write([]byte{byte(pst.SignatureTypeTableContext)})

	// Column count.
	header.WriteByte(byte(len(columnDescriptors)))

	// This is an array of 4 16-bit values that specify the offsets of various groups of data in the actual row data.
	header.Write(make([]byte, 2)) // Ending offset of 8- and 4-byte data value group.
	header.Write(make([]byte, 2)) // Ending offset of 2-byte data value group.
	header.Write(make([]byte, 2)) // Ending offset of 1-byte data value group.
	header.Write(make([]byte, 2)) // Ending offset of the Cell Existence Block.

	// HID to the Row ID BTH (hidRowIndex).
	// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcrowid
	header.Write(make([]byte, 4))

	// HNID to the Row Matrix (that is, actual table data). (hnidRows)
	header.Write(make([]byte, 4))

	// Deprecated (hidIndex)
	header.Write(make([]byte, 4))

	// Sort column descriptors according to specification.
	slices.SortFunc(columnDescriptors, func(a []byte, b []byte) int {
		return cmp.Compare(binary.LittleEndian.Uint16(a[2:2+2]), binary.LittleEndian.Uint16(b[2:2+2]))
	})

	// Array of Column Descriptors.
	for _, columnDescriptor := range columnDescriptors {
		header.Write(columnDescriptor)
	}

	return header.WriteTo(writer)
}

// WriteRowMatrix writes the Row Matrix of the Table Context.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#row-matrix
func (tableContextWriter *TableContextWriter) WriteRowMatrix(writer io.Writer) (int64, error) {
	//rowMatrix := bytes.NewBuffer(make([]byte, 0)) // TODO - Size

	return 0, nil
}
