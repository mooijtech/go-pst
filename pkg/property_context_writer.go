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
)

// PropertyContextWriter represents a writer for a pst.PropertyContext.
type PropertyContextWriter struct {
	// Writer represents the io.Writer used when writing.
	Writer io.Writer
	// FormatType represents the FormatType used while writing.
	FormatType FormatType
	// WriteGroup represents Goroutines running writers.
	WriteGroup *errgroup.Group
	// BTreeOnHeapWriter represents the BTreeOnHeapWriter.
	BTreeOnHeapWriter *BTreeOnHeapWriter
	// PropertyWriter represents the PropertyWriter.
	PropertyWriter *PropertyWriter
	// PropertyWriteCallbackChannel represents the callback channel for writable properties.
	PropertyWriteCallbackChannel chan int64
	// LocalDescriptorsWriter represents the LocalDescriptorsWriter.
	LocalDescriptorsWriter *LocalDescriptorsWriter
}

// NewPropertyContextWriter creates a new PropertyContextWriter.
func NewPropertyContextWriter(writer io.Writer, writeGroup *errgroup.Group, formatType FormatType) *PropertyContextWriter {
	heapOnNodeWriter := NewHeapOnNodeWriter(SignatureTypePropertyContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)
	// propertyWriteCallbackChannel returns the written byte size, used to wait for properties to be written.
	propertyWriteCallbackChannel := make(chan int64)
	// propertyWriter starts a Go channel for writing properties.
	propertyWriter := NewPropertyWriter(writer, writeGroup, propertyWriteCallbackChannel, formatType)
	localDescriptorsWriter := NewLocalDescriptorsWriter(writer, writeGroup, formatType)

	propertyContextWriter := &PropertyContextWriter{
		Writer:                       writer,
		FormatType:                   formatType,
		WriteGroup:                   writeGroup,
		BTreeOnHeapWriter:            btreeOnHeapWriter,
		PropertyWriter:               propertyWriter,
		PropertyWriteCallbackChannel: propertyWriteCallbackChannel,
		LocalDescriptorsWriter:       localDescriptorsWriter,
	}

	return propertyContextWriter
}

// Add adds the properties (properties.Message, properties.Attachment, etc.) to the write queue.
// Writable properties (Property) are returned to the PropertyWriteCallbackChannel.
func (propertyContextWriter *PropertyContextWriter) Add(properties ...proto.Message) {
	propertyContextWriter.PropertyWriter.Add(properties...)
}

// WriteTo writes the pst.PropertyContext.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#property-context-pc
func (propertyContextWriter *PropertyContextWriter) WriteTo(writer io.Writer) (int64, error) {
	btreeOnHeapWrittenSize, err := propertyContextWriter.BTreeOnHeapWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Heap-on-Node")
	}

	// Wait for the properties to be written.
	// TODO - Wait
	// TODO - Write to the structures directly, remove WriteTo altogether.
	var totalSize int64

	for writtenSize := range propertyContextWriter.PropertyWriteCallbackChannel {
		totalSize += writtenSize
	}

	return btreeOnHeapWrittenSize + totalSize, nil
}

// WriteProperty writes the property.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#pc-bth-record
func (propertyContextWriter *PropertyContextWriter) WriteProperty(writer io.Writer, property Property) (int64, error) {
	propertyBuffer := bytes.NewBuffer(make([]byte, 8))

	// Property ID
	propertyBuffer.Write(GetUint16(uint16(property.Identifier)))
	// Property Type
	propertyBuffer.Write(GetUint16(uint16(property.Type)))
	// Value
	if property.Value.Len() <= 4 {
		if _, err := property.Value.WriteTo(propertyBuffer); err != nil {
			return 0, eris.Wrap(err, "failed to write property value")
		}
	} else if property.Value.Len() <= 3580 {
		// HID
	} else {
		// NID Local Descriptor
		localDescriptorIdentifier := propertyContextWriter.LocalDescriptorsWriter.AddProperty(property)

		propertyBuffer.Write(localDescriptorIdentifier.Bytes(propertyContextWriter.FormatType))
	}

	return propertyBuffer.WriteTo(writer)
}
