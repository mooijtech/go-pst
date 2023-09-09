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
	"encoding/binary"
	"fmt"
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"github.com/rotisserie/eris"
	"github.com/tinylib/msgp/msgp"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// PropertyWriter represents a writer for properties.
// The PropertyContext should be used as a higher structure which manages this PropertyWriter.
type PropertyWriter struct {
	// StreamWriter is used to send write requests to a write channel and receive the results via a callback channel.
	StreamWriter *StreamWriter
	// FormatType represents the FormatType used while writing.
	FormatType FormatType
	// PropertyCount represents the amount of properties this PropertyWriter will write.
	PropertyCount int
}

// NewPropertyWriter creates a new PropertyWriter.
func NewPropertyWriter(writer io.WriteSeeker, writeGroup *errgroup.Group, formatType FormatType) *PropertyWriter {
	propertyWriter := &PropertyWriter{
		FormatType:    formatType,
		PropertyCount: 0,
	}

	streamWriter := NewStreamWriter(writer, writeGroup).WithTransform(propertyWriter.PropertyTransform)

	propertyWriter.StreamWriter = streamWriter

	// Start the stream writer.
	streamWriter.StartWriteChannel()

	// Start the property write channel.
	propertyWriter.StartPropertyWriteChannel()

	return propertyWriter
}

// WritableProperty for the StreamWriter.
type WritableProperty struct {
	Value proto.Message
}

func (writableProperty *WritableProperty) WriteTo(writer io.Writer) (int64, error) {

}

// AddProperties sends the properties to the write queue, picked up by Goroutines.
// Properties will be written to the PropertyWriteCallbackChannel (see StartPropertyWriteChannel).
func (propertyWriter *PropertyWriter) AddProperties(properties ...*WritableProperty) {
	for _, property := range properties {
		propertyWriter.StreamWriter.WriteChannel <- property
		propertyWriter.PropertyCount++
	}
}

// PropertyTransform sends the byte representation of the properties to a write channel and returns the writable properties.
func (propertyWriter *PropertyWriter) PropertyTransform(receivedProperties ...StreamRequest) ([]StreamResponse, error) {
	writableProperties, err := propertyWriter.GetProperties(receivedProperties...)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get writable properties")
	}

	return writableProperties, nil
}

//// StartPropertyWriteChannel starts the Go channel for writing properties.
//func (propertyWriter *PropertyWriter) StartPropertyWriteChannel() error {
//	propertyWriter.StreamWriter.WriteChannel <-
//
//	// Create writable byte representation of the properties.
//	writableProperties, err := propertyWriter.GetProperties(receivedProperties)
//
//	if err != nil {
//		return eris.Wrap(err, "failed to get writable properties")
//	}
//
//
//	propertyWriter.StreamWriter
//
//	// The caller is already in a Goroutine.
//	for receivedProperties := range propertyWriter.PropertyWriteChannel {
//		propertyWriter.WriteGroup.Go(func() error {
//			// Create writable byte representation of the properties.
//			writableProperties, err := propertyWriter.GetProperties(receivedProperties)
//
//			if err != nil {
//				return eris.Wrap(err, "failed to get writable properties")
//			}
//
//			// Callback, the PropertyContext handles writing this to the correct place.
//			for _, writableProperty := range writableProperties {
//				propertyWriter.PropertyWriteCallbackChannel <- writableProperty
//			}
//
//			return nil
//		})
//	}
//}

// GetProperties returns the writable properties.
// This code is a hot-path, do not use reflection here.
// Instead, we use code-generated setters thanks to https://github.com/tinylib/msgp (deserialize into structs).
func (propertyWriter *PropertyWriter) GetProperties(protoMessage ...StreamRequest) ([]StreamResponse, error) {
	var totalSize int

	totalSize += int(GetIdentifierSize(propertyWriter.FormatType))
	// TODO -

	messagePackBuffer := bytes.NewBuffer(make([]byte, totalSize))
	messagePackWriter := msgp.NewWriterSize(messagePackBuffer, totalSize)

	switch property := protoMessage.(type) {
	case *properties.Message:
		// TODO - Skip nil!
		property.MarshalMsg()
	case *properties.Attachment:

	}

	var properties []Property

	propertyTypes := reflect.TypeOf(protoMessage).Elem()
	propertyValues := reflect.ValueOf(protoMessage).Elem()

	for i := 0; i < propertyTypes.NumField(); i++ {
		if !propertyTypes.Field(i).IsExported() || propertyValues.Field(i).IsNil() {
			continue
		}

		// Get the struct tag which we use to get the property ID and property type.
		// These struct tags are generated by cmd/properties/generate.go.
		tag := strings.ReplaceAll(propertyTypes.Field(i).Tag.Get("msg"), ",omitempty", "")

		if tag == "" {
			fmt.Printf("Skipping property without tag: %s\n", propertyTypes.Field(i).Name)
			continue
		}

		propertyID, err := strconv.Atoi(strings.Split(tag, "-")[0])

		if err != nil {
			return nil, eris.Wrap(err, "failed to convert propertyID to int")
		}

		propertyType, err := strconv.Atoi(strings.Split(tag, "-")[1])

		if err != nil {
			return nil, eris.Wrap(err, "failed to convert propertyType to int")
		}

		var propertyBuffer bytes.Buffer

		switch propertyValue := propertyValues.Field(i).Elem().Interface().(type) {
		case string:
			// Binary is intended for fixed-size structures with obvious encodings.
			// Strings are not fixed size and do not have an obvious encoding.
			if _, err := io.WriteString(&propertyBuffer, propertyValue); err != nil {
				return nil, eris.Wrap(err, "failed to write string")
			}
		default:
			if err := binary.Write(&propertyBuffer, binary.LittleEndian, propertyValue); err != nil {
				return nil, eris.Wrap(err, "failed to write property")
			}
		}

		properties = append(properties, Property{
			Identifier: Identifier(propertyID),
			Type:       PropertyType(propertyType),
			Value:      propertyBuffer,
		})
	}

	return properties, nil
}
