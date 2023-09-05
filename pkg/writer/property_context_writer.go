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
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"io"
	"log/slog"
	"sync/atomic"
)

// PropertyContextWriter represents a writer for a pst.PropertyContext.
type PropertyContextWriter struct {
	// Writer represents the io.Writer used when writing.
	Writer io.Writer
	// BTreeOnHeapWriter represents the BTreeOnHeapWriter.
	BTreeOnHeapWriter *BTreeOnHeapWriter
	// PropertyWriter represents the PropertyWriter.
	PropertyWriter *PropertyWriter
	// PropertyWriteCallbackChannel represents the callback channel for writable properties.
	PropertyWriteCallbackChannel chan Property
	// LocalDescriptorsWriter represents the LocalDescriptorsWriter.
	LocalDescriptorsWriter *LocalDescriptorsWriter
	// WriteGroup represents Goroutines running writers.
	WriteGroup *errgroup.Group
	// TotalSize represents the total byte size written.
	// TODO - Don't use atomic, pass int64 to callback per processed?
	TotalSize atomic.Int64
}

// NewPropertyContextWriter creates a new PropertyContextWriter.
func NewPropertyContextWriter(formatType pst.FormatType, btreeType pst.BTreeType, btreeNodes []pst.Identifier, writeGroup *errgroup.Group) *PropertyContextWriter {
	heapOnNodeWriter := NewHeapOnNodeWriter(pst.SignatureTypePropertyContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)
	propertyWriteCallbackChannel := make(chan Property)
	propertyWriter := NewPropertyWriter(writeGroup, propertyWriteCallbackChannel)
	localDescriptorsWriter := NewLocalDescriptorsWriter(formatType, btreeType, btreeNodes)

	propertyContextWriter := &PropertyContextWriter{
		BTreeOnHeapWriter:            btreeOnHeapWriter,
		PropertyWriter:               propertyWriter,
		PropertyWriteCallbackChannel: propertyWriteCallbackChannel,
		LocalDescriptorsWriter:       localDescriptorsWriter,
		WriteGroup:                   writeGroup,
	}

	// Start Go channel for the property write callback.
	go propertyContextWriter.StartWriteCallbackHandler(propertyWriteCallbackChannel)

	return propertyContextWriter
}

// StartWriteCallbackHandler handles writing a received writable Property.
func (propertyContextWriter *PropertyContextWriter) StartWriteCallbackHandler(writeCallbackChannel chan Property) {
	for property := range writeCallbackChannel {
		propertyContextWriter.WriteGroup.Go(func() error {
			slog.Debug("Writing property...", "identifier", property.ID)

			// Write the property.
			if _, err := property.WriteTo(propertyContextWriter.Writer); err != nil {
				return eris.Wrap(err, "failed to write property")
			}

			return nil
		})
	}
}

func (propertyContextWriter *PropertyContextWriter) AddProperties(properties ...proto.Message) {
	propertyContextWriter.PropertyWriter.AddProperties(properties...)
}

func (propertyContextWriter *PropertyContextWriter) AddRawProperty(property Property) {

}

// WriteTo writes the pst.PropertyContext.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#property-context-pc
func (propertyContextWriter *PropertyContextWriter) WriteTo(writer io.Writer) (int64, error) {
	btreeOnHeapWrittenSize, err := propertyContextWriter.BTreeOnHeapWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Heap-on-Node")
	}

	propertiesWrittenSize, err := propertyContextWriter.WriteProperties(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write properties")
	}

	return btreeOnHeapWrittenSize + propertiesWrittenSize, nil
}

// WriteProperties writes the properties.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#pc-bth-record
func (propertyContextWriter *PropertyContextWriter) WriteProperties(writer io.Writer) (int64, error) {
	properties, err := propertyContextWriter.PropertyWriter.GetProperties()

	if err != nil {
		return 0, eris.Wrap(err, "failed to get properties")
	}

	var totalSize int64

	// Write properties.
	for _, property := range properties {
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
		}

		written, err := propertyBuffer.WriteTo(writer)

		if err != nil {
			return 0, eris.Wrap(err, "failed to write property")
		}

		totalSize += written
	}

	return totalSize, nil
}
