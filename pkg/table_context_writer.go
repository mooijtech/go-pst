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
	"bytes"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"io"
	"math"
)

// TableContextWriter represents a writer for a pst.TableContext.
// The TableContext is used to store identifiers pointing to folders and messages.
type TableContextWriter struct {
	// Writer represents the io.Writer used while writing.
	Writer io.WriteSeeker
	// WriteGroup represents writers running in Goroutines.
	WriteGroup *errgroup.Group
	// FormatType represents the FormatType.
	FormatType FormatType
	// PropertyWriter represents the PropertyWriter.
	PropertyWriter *PropertyWriter
	// PropertyWriteCallbackChannel represents the callback for writable properties.
	// These writable properties are sent to the ColumnDescriptorsWriteChannel and RowMatrixWriterChannel.
	PropertyWriteCallbackChannel chan Property
	// HeaderWriteChannel represents the Go channel used for writing ColumnDescriptor to the header.
	// See StartHeaderWriteChannel.
	HeaderWriteChannel chan *bytes.Buffer
	// RowMatrixWriteChannel represents the channel used for writing to the Row Matrix.
	// See StartRowMatrixWriteChannel.
	RowMatrixWriteChannel chan Property
	// TableContextWriteCallback represents the callback used when the TableContextWriter is done writing.
	TableContextWriteCallback chan int64
}

// NewTableContextWriter creates a new TableContextWriter.
func NewTableContextWriter(writer io.WriteSeeker, writeGroup *errgroup.Group, parentIdentifier Identifier, formatType FormatType) (*TableContextWriter, error) {
	// Create PropertyWriter (see StartChannels).
	propertyWriteCallbackChannel := make(chan Property)
	propertyWriter := NewPropertyWriter(writer, writeGroup, propertyWriteCallbackChannel, formatType)

	// Create TableContextWriter
	tableContextWriteCallback := make(chan int64)
	tableContextWriter := &TableContextWriter{
		Writer:                       writer,
		WriteGroup:                   writeGroup,
		FormatType:                   formatType,
		PropertyWriter:               propertyWriter,
		PropertyWriteCallbackChannel: propertyWriteCallbackChannel,
		TableContextWriteCallback:    tableContextWriteCallback,
	}

	// Write the BTree-on-Heap.
	if err := tableContextWriter.WriteBTreeOnHeap(); err != nil {
		return nil, eris.Wrap(err, "failed to write BTree-on-Heap")
	}

	// Start channels for writing
	tableContextWriter.StartChannels()

	return tableContextWriter, nil
}

// AddIdentifier adds a reference to a folder or message to the TableContext.
func (tableContextWriter *TableContextWriter) AddIdentifier(identifier Identifier) {
	identifierProperty := Property{
		Identifier: 26610, // 26610 is always used for identifiers TODO reference actual PropertyName
		Type:       PropertyTypeInteger32,
		Value:      bytes.NewBuffer(identifier.Bytes(tableContextWriter.FormatType)),
	}

	tableContextWriter.PropertyWriteCallbackChannel <- identifierProperty
}

// WriteBTreeOnHeap writes the BTreeOnHeap of the TableContext.
func (tableContextWriter *TableContextWriter) WriteBTreeOnHeap() error {
	heapOnNodeWriter := NewHeapOnNodeWriter(SignatureTypeTableContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)
	btreeOnHeapWrittenSize, err := btreeOnHeapWriter.WriteTo(tableContextWriter.Writer)

	if err != nil {
		return eris.Wrap(err, "failed to write BTree-on-Heap")
	}

	tableContextWriter.TableContextWriteCallback <- btreeOnHeapWrittenSize

	return nil
}

// StartChannels starts the channels for writing.
//
// Forwards PropertyWriteCallbackChannel to the HeaderWriteChannel and RowMatrixWriteChannel.
// HeaderWriteChannel writes ColumnDescriptor to the Header.
// RowMatrixWriteChannel writes Property to the Row Matrix.
func (tableContextWriter *TableContextWriter) StartChannels() {
	go tableContextWriter.StartPropertyWriteCallbackChannel()
	go tableContextWriter.StartHeaderWriteChannel()
	go tableContextWriter.StartRowMatrixWriteChannel()
}

// Add the properties to the PropertyWriter write queue.
// Each Property is sent back to the PropertyWriteCallbackChannel (see StartPropertyWriteCallbackChannel).
func (tableContextWriter *TableContextWriter) Add(protoMessages ...proto.Message) {
	tableContextWriter.PropertyWriter.Add(protoMessages...)
}

// StartPropertyWriteCallbackChannel forwards the Property to the HeaderWriteChannel and RowMatrixWriteChannel.
func (tableContextWriter *TableContextWriter) StartPropertyWriteCallbackChannel() {
	propertyIndex := 0

	for writableProperty := range tableContextWriter.PropertyWriteCallbackChannel {
		// Create ColumnDescriptor of the Property.
		// Send ColumnDescriptor to the HeaderWriteChannel (see StartHeaderWriteChannel).
		tableContextWriter.HeaderWriteChannel <- tableContextWriter.GetColumnDescriptor(writableProperty, propertyIndex)

		// Send Property to the RowMatrixWriteChannel (see StartRowMatrixWriteChannel).
		tableContextWriter.RowMatrixWriteChannel <- writableProperty

		// Used to calculate the ColumnDescriptor start offset of this property.
		propertyIndex++
	}
}

type RowMatrixOffsets struct {
}

// StartHeaderWriteChannel writes the header.
// Waits for HeaderWriteChannel to write ColumnDescriptor.
func (tableContextWriter *TableContextWriter) StartHeaderWriteChannel() {
	tableContextWriter.WriteGroup.Go(func() error {
		// TODO - Move everything here.
		
		return nil
	})

	// Skip past the header until we have received all Column Descriptors.
	if _, err := tableContextWriter.Writer.Seek(22, io.SeekCurrent); err != nil {
		return eris.Wrap(err, "failed to seek")
	}

	// Write Column Descriptors (references to properties).
	for columnDescriptor := range tableContextWriter.HeaderWriteChannel {
		columnDescriptorWrittenSize, err := columnDescriptor.WriteTo(tableContextWriter.Writer)

		if err != nil {
			return eris.Wrap(err, "failed to write column descriptor")
		}

		tableContextWriter.TableContextWriteCallback <- int64(columnDescriptor.Len())
	}

	// Move back and write the header now that all properties have been written.

	propertyCount := tableContextWriter.PropertyWriter.PropertyCount

	// 1+1+8+4+4+4+columnDescriptors
	header := bytes.NewBuffer(make([]byte, 22+(8*propertyCount)))

	// MUST be set to SignatureTypeTableContext.
	header.Write([]byte{byte(SignatureTypeTableContext)})

	// Column count.
	header.WriteByte(byte(propertyCount))

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
	// TODO-
	//slices.SortFunc(columnDescriptors, func(a []byte, b []byte) int {
	//	return cmp.Compare(binary.LittleEndian.Uint16(a[2:2+2]), binary.LittleEndian.Uint16(b[2:2+2]))
	//})
}

// StartRowMatrixWriteChannel writes the Property to the Row Matrix.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#row-matrix
func (tableContextWriter *TableContextWriter) StartRowMatrixWriteChannel() {
	// Write the Row Matrix.

	for writableProperty := range tableContextWriter.RowMatrixWriteChannel {
		writableProperty.WriteTo(tableContextWriter.Writer, tableContextWriter.FormatType)
	}
}

// GetColumnDescriptor returns the ColumnDescriptor of the Property.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcoldesc
func (tableContextWriter *TableContextWriter) GetColumnDescriptor(property Property, propertyIndex int) *bytes.Buffer {
	columnDescriptorBuffer := bytes.NewBuffer(make([]byte, 8))

	// Tag
	columnDescriptorBuffer.Write(GetUint16(uint16(property.Identifier)))
	columnDescriptorBuffer.Write(GetUint16(uint16(property.Type)))

	// Offset
	columnDescriptorBuffer.Write(GetUint16(uint16(8 * propertyIndex)))

	// Data Size
	if property.Value.Len() <= 8 {
		// Property value size.
		columnDescriptorBuffer.WriteByte(byte(property.Value.Len()))
	} else {
		// Variable-sized data (size of an Identifier)
		columnDescriptorBuffer.WriteByte(byte(4))
	}

	// Cell Existence Bitmap Index
	columnDescriptorBuffer.WriteByte(byte(2 + propertyIndex)) // Skip 0 and 1

	return columnDescriptorBuffer
}

// WriteRowMatrix writes the Row Matrix of the Table Context.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#row-matrix
func (tableContextWriter *TableContextWriter) WriteRowMatrix(writer io.Writer) (int64, error) {
	rowMatrixBuffer := bytes.NewBuffer(make([]byte, tableContextWriter.GetRowMatrixSize(tableContextWriter.PropertyWriter.PropertyCount)))

	// Sort by byte-size so writes are aligned accordingly.
	//slices.SortFunc(properties, func(a, b Property) int {
	//	return cmp.Compare(a.Value.Len(), b.Value.Len())
	//})

	//propertyIndex := 0
	//
	//for writableProperty := range tableContextWriter.PropertyWriteCallbackChannel {
	//	// Create a ColumnDescriptor for the header.
	//	tableContextWriter.ColumnDescriptorCallbackChannel <- tableContextWriter.GetColumnDescriptor(writableProperty, propertyIndex)
	//	propertyIndex++
	//}

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
