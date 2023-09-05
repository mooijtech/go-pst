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
	"google.golang.org/protobuf/proto"
	"io"
)

// PropertyContextWriter represents a writer for a pst.PropertyContext.
type PropertyContextWriter struct {
	// BTreeOnHeapWriter represents the BTreeOnHeapWriter.
	BTreeOnHeapWriter *BTreeOnHeapWriter
	// PropertiesWriter represents the PropertiesWriter.
	PropertiesWriter *PropertyWriter
	// LocalDescriptorsWriter represents the LocalDescriptorsWriter.
	LocalDescriptorsWriter *LocalDescriptorsWriter
}

// NewPropertyContextWriter creates a new PropertyContextWriter.
func NewPropertyContextWriter(properties []*proto.Message) *PropertyContextWriter {
	heapOnNodeWriter := NewHeapOnNodeWriter(pst.SignatureTypePropertyContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)
	propertiesWriter := NewPropertyWriter(properties)
	localDescriptorsWriter := NewLocalDescriptorsWriter()

	return &PropertyContextWriter{
		BTreeOnHeapWriter:      btreeOnHeapWriter,
		PropertiesWriter:       propertiesWriter,
		LocalDescriptorsWriter: localDescriptorsWriter,
	}
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
	properties, err := propertyContextWriter.PropertiesWriter.GetProperties()

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
