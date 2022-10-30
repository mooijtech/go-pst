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
	"bytes"
	_ "embed"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"io"
)

// PropertyContext represents the property context.
type PropertyContext struct {
	Properties []Property
	HeapOnNode *HeapOnNode
	File       *File
}

// GetPropertyByID returns the property by ID.
func (propertyContext *PropertyContext) GetPropertyByID(propertyID uint16) (Property, error) {
	for _, property := range propertyContext.Properties {
		if property.ID == propertyID {
			return property, nil
		}
	}

	return Property{}, errors.WithStack(ErrPropertyNotFound)
}

// GetPropertyReader returns the reader for the property.
func (propertyContext *PropertyContext) GetPropertyReader(propertyID uint16, localDescriptors ...LocalDescriptor) (PropertyReader, error) {
	property, err := propertyContext.GetPropertyByID(propertyID)

	if err != nil {
		return PropertyReader{}, errors.WithStack(err)
	}

	return NewPropertyReader(property, propertyContext.HeapOnNode, propertyContext.File, localDescriptors...)
}

// GetPropertyContext returns the property context (BC Table).
// References https://github.com/mooijtech/go-pst/tree/master/docs#property-context-pc
func (file *File) GetPropertyContext(heapOnNode *HeapOnNode) (*PropertyContext, error) {
	tableType, err := heapOnNode.GetTableType()

	if err != nil {
		return nil, errors.WithStack(err)
	} else if tableType != 188 {
		// Must be Property Context.
		return nil, errors.WithStack(ErrTableTypeInvalid)
	}

	btreeOnHeapHeader, err := file.GetBTreeOnHeapHeader(heapOnNode)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	keyTableReader, err := file.GetHeapOnNodeReaderFromHNID(btreeOnHeapHeader.HIDRoot, *heapOnNode.Reader)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	keyCount := int(keyTableReader.Size()) / int(btreeOnHeapHeader.KeySize+btreeOnHeapHeader.ValueSize)

	var properties []Property
	offset := int64(0)

	for i := 0; i < keyCount; i++ {
		// PropertyContextItem represents an item within the property context.
		// References "Property Context B-Tree-on-Heap Record".
		// References [MS-PDF]: 2.3.3.3 PC BTH Record
		var property Property

		propertyID := make([]byte, 2)

		if _, err := keyTableReader.ReadAt(propertyID, offset); err != nil {
			return nil, errors.WithStack(err)
		}

		propertyType := make([]byte, 2)

		if _, err := keyTableReader.ReadAt(propertyType, offset+2); err != nil {
			return nil, errors.WithStack(err)
		}

		data := make([]byte, 4)

		if _, err := keyTableReader.ReadAt(data, offset+4); err != nil {
			return nil, errors.WithStack(err)
		}

		property.ID = binary.LittleEndian.Uint16(propertyID)
		property.Type = PropertyType(binary.LittleEndian.Uint16(propertyType))

		// Property Context uses a HNID for any data (PropertyType) exceeding 4 bytes.
		// Otherwise, the data is small enough to fit in the Property directly.
		// The PropertyReader will handle type conversion.
		if property.Type.GetDataSize() != -1 && property.Type.GetDataSize() <= 4 {
			property.Data = data
		} else {
			// Variable size data.
			property.HNID = Identifier(binary.LittleEndian.Uint32(data))
		}

		properties = append(properties, property)
		offset += 8
	}

	return &PropertyContext{
		Properties: properties,
		HeapOnNode: heapOnNode,
		File:       file,
	}, nil
}

// Populate all properties to the decodable.
func (propertyContext *PropertyContext) Populate(decodable msgp.Decodable, localDescriptors []LocalDescriptor) error {
	// Populate the properties.
	messagePackBuffer := &bytes.Buffer{}
	messagePackWriter := msgp.NewWriter(messagePackBuffer)

	if err := messagePackWriter.WriteMapHeader(uint32(len(propertyContext.Properties))); err != nil {
		return errors.WithStack(err)
	}

	for _, property := range propertyContext.Properties {
		propertyReader, err := propertyContext.GetPropertyReader(property.ID, localDescriptors...)

		if err != nil && !errors.Is(err, ErrPropertyNoData) {
			return errors.WithStack(err)
		}

		err = propertyReader.WriteMessagePackValue(messagePackWriter)

		if err != nil && !errors.Is(err, ErrPropertyNoData) {
			return errors.WithStack(err)
		}
	}

	if err := messagePackWriter.Flush(); err != nil {
		return errors.WithStack(err)
	}

	err := decodable.DecodeMsg(msgp.NewReader(messagePackBuffer))

	if err != nil && !errors.Is(err, io.EOF) {
		return errors.WithStack(err)
	}

	return nil
}
