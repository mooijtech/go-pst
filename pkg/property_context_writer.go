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
	"log/slog"
)

// PropertyContextWriter represents a writer for a pst.PropertyContext.
type PropertyContextWriter struct {
	// Writer represents the io.Writer used when writing.
	Writer io.Writer
	// WriteGroup represents Goroutines running writers.
	WriteGroup *errgroup.Group
	// BTreeOnHeapWriter represents the BTreeOnHeapWriter.
	BTreeOnHeapWriter *BTreeOnHeapWriter
	// PropertyWriter represents the PropertyWriter.
	PropertyWriter *PropertyWriter
	// PropertyWriteCallbackChannel represents the callback channel for writable properties.
	PropertyWriteCallbackChannel chan WriteCallbackResponse
	// LocalDescriptorsWriter represents the LocalDescriptorsWriter.
	LocalDescriptorsWriter *LocalDescriptorsWriter
}

// NewPropertyContextWriter creates a new PropertyContextWriter.
func NewPropertyContextWriter(writer io.Writer, writeGroup *errgroup.Group, propertyWriteCallbackChannel chan WriteCallbackResponse, formatType FormatType, btreeType BTreeType) *PropertyContextWriter {
	heapOnNodeWriter := NewHeapOnNodeWriter(SignatureTypePropertyContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)
	propertyWriteChannel := make(chan Property)
	propertyWriter := NewPropertyWriter(writeGroup, propertyWriteChannel)
	localDescriptorsWriter := NewLocalDescriptorsWriter(writer, writeGroup, formatType, btreeType)

	propertyContextWriter := &PropertyContextWriter{
		Writer:                       writer,
		WriteGroup:                   writeGroup,
		BTreeOnHeapWriter:            btreeOnHeapWriter,
		PropertyWriter:               propertyWriter,
		PropertyWriteCallbackChannel: propertyWriteCallbackChannel,
		LocalDescriptorsWriter:       localDescriptorsWriter,
	}

	// Start the write channel.
	go propertyContextWriter.StartWriteChannel(propertyWriteChannel)

	return propertyContextWriter
}

// AddProperties adds the properties to the write queue.
// Writable properties are returned to the StartWriteCallbackChannel.
func (propertyContextWriter *PropertyContextWriter) AddProperties(properties ...proto.Message) {
	propertyContextWriter.PropertyWriter.AddProperties(properties...)
}

// StartWriteChannel handles writing a received writable Property.
func (propertyContextWriter *PropertyContextWriter) StartWriteChannel(writeChannel chan Property) {
	for property := range writeChannel {
		propertyContextWriter.WriteGroup.Go(func() error {
			// Write the property.
			slog.Debug("Writing property...", "identifier", property.ID)

			written, err := property.WriteTo(propertyContextWriter.Writer)

			if err != nil {
				return eris.Wrap(err, "failed to write property")
			}

			// Callback amount of bytes written.
			propertyContextWriter.PropertyWriteCallbackChannel <- NewWriteCallbackResponse(written)

			return nil
		})
	}
}

// WriteTo writes the pst.PropertyContext.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#property-context-pc
func (propertyContextWriter *PropertyContextWriter) WriteTo(writer io.Writer) (int64, error) {
	btreeOnHeapWrittenSize, err := propertyContextWriter.BTreeOnHeapWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Heap-on-Node")
	}

	// Wait for the properties to be written.
	var totalSize int64

	for writeCallbackResponse := range propertyContextWriter.PropertyWriteCallbackChannel {
		totalSize += writeCallbackResponse.Written
	}

	return btreeOnHeapWrittenSize + totalSize, nil
}

// WriteProperty writes the property.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#pc-bth-record
func (propertyContextWriter *PropertyContextWriter) WriteProperty(writer io.Writer, property Property) (int64, error) {
	propertyBuffer := bytes.NewBuffer(make([]byte, 8))

	// Property ID
	propertyBuffer.Write(GetUint16(uint16(property.ID)))
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
		propertyContextWriter.LocalDescriptorsWriter.AddProperty()
	}

	return propertyBuffer.WriteTo(writer)
}
