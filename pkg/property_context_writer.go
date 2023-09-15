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
	// streamWriter represents the writer for the PropertyContextWriter.
	streamWriter *StreamWriter
	// formatType represents the FormatType used while writing.
	formatType FormatType
	// btreeOnHeapWriter represents the BTreeOnHeapWriter.
	btreeOnHeapWriter *BTreeOnHeapWriter
	// propertyWriter represents the PropertyWriter.
	propertyWriter *PropertyWriter
	// propertyWriteCallbackChannel represents the callback channel for writable properties.
	propertyWriteCallbackChannel chan *bytes.Buffer
	// LocalDescriptorsWriter represents the LocalDescriptorsWriter.
	localDescriptorsWriter *LocalDescriptorsWriter
}

// NewPropertyContextWriter creates a new PropertyContextWriter.
func NewPropertyContextWriter(writer io.WriteSeeker, writeGroup *errgroup.Group, propertyContextWriteCallback chan int64, formatType FormatType) (*PropertyContextWriter, error) {
	// Stream writer is used to write the property context.
	streamWriter := NewStreamWriter[io.WriterTo, int64](writer, writeGroup)

	// Start the write channel.
	streamWriter.StartWriteChannel()
	// Send the write responses to the parent for calculating the total size.
	streamWriter.RegisterCallback(propertyContextWriteCallback)

	// Structures below the PropertyContext.
	heapOnNodeWriter := NewHeapOnNodeWriter(SignatureTypePropertyContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)
	localDescriptorsWriter := NewLocalDescriptorsWriter(writer, writeGroup, formatType, BTreeTypeBlock)

	// Property writer.
	// TODO - Fix callbacks.
	propertyWriteCallbackChannel := make(chan *bytes.Buffer)
	propertyWriter := NewPropertyWriter(writer, writeGroup, formatType, propertyWriteCallbackChannel)

	// Registers the property write callback.
	if err := propertyWriter.RegisterPropertyWriteCallback(propertyWriteCallbackChannel); err != nil {
		return nil, eris.Wrap(err, "failed to register property write callback")
	}

	// PropertyContext writer.
	return &PropertyContextWriter{
		streamWriter:                 streamWriter,
		formatType:                   formatType,
		btreeOnHeapWriter:            btreeOnHeapWriter,
		propertyWriter:               propertyWriter,
		propertyWriteCallbackChannel: propertyWriteCallbackChannel,
		localDescriptorsWriter:       localDescriptorsWriter,
	}, nil
}

// AddProperties adds the properties (properties.Message, properties.Attachment, etc.) to the write queue.
// Sends WritableProperties to StreamWriter and returns []Property (see PropertyWriteCallbackChannel).
// Once we have []Property we convert to a byte representation to be written.
func (propertyContextWriter *PropertyContextWriter) AddProperties(properties ...proto.Message) error {
	return propertyContextWriter.propertyWriter.AddProperties(properties...)
}

// WriteTo writes the pst.PropertyContext.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#property-context-pc
func (propertyContextWriter *PropertyContextWriter) WriteTo(writer io.Writer) (int64, error) {
	// Write the BTree-on-Heap.
	btreeOnHeapWrittenSize, err := propertyContextWriter.btreeOnHeapWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Heap-on-Node")
	}

	// Write the properties.
	var propertiesWrittenSize int64

	for streamResponse := range propertyContextWriter.propertyWriteCallbackChannel {
		written, err := streamResponse.WriteTo(writer)

		if err != nil {
			return 0, eris.Wrap(err, "failed to write property")
		}

		propertiesWrittenSize += written
	}

	return btreeOnHeapWrittenSize + propertiesWrittenSize, nil
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
		// TODO -
	} else {
		// NID (Local Descriptor)
		// Reference the identifier of the created Local Descriptor.
		localDescriptorIdentifier := propertyContextWriter.localDescriptorsWriter.AddProperty(property)

		propertyBuffer.Write(localDescriptorIdentifier.Bytes(propertyContextWriter.formatType))
	}

	return propertyBuffer.WriteTo(writer)
}
