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
	_ "embed"
	"encoding/binary"
	"encoding/csv"
	"github.com/rotisserie/eris"
	"github.com/tinylib/msgp/msgp"
	"io"
	"strings"
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

	return Property{}, ErrPropertyNotFound
}

// GetPropertyReader returns the reader for the property, may return ErrPropertyNotFound.
func (propertyContext *PropertyContext) GetPropertyReader(propertyID uint16, localDescriptors []LocalDescriptor) (PropertyReader, error) {
	property, err := propertyContext.GetPropertyByID(propertyID)

	if err != nil {
		return PropertyReader{}, ErrPropertyNotFound
	}

	return NewPropertyReader(property, propertyContext.HeapOnNode, propertyContext.File, localDescriptors)
}

// GetPropertyContext returns the property context (BC Table).
// References https://github.com/mooijtech/go-pst/tree/master/docs#property-context-pc
func (file *File) GetPropertyContext(heapOnNode *HeapOnNode) (*PropertyContext, error) {
	tableType, err := heapOnNode.GetTableType()

	if err != nil {
		return nil, eris.Wrap(err, "failed to get table type")
	} else if tableType != 188 {
		// Must be Property Context.
		return nil, ErrTableTypeInvalid
	}

	btreeOnHeapHeader, err := file.GetBTreeOnHeapHeader(heapOnNode)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get b-tree-on-heap header")
	}

	keyTableReader, err := file.GetHeapOnNodeReaderFromHNID(btreeOnHeapHeader.HIDRoot, *heapOnNode.Reader)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get key table reader")
	}

	keyCount := int(keyTableReader.Size()) / int(btreeOnHeapHeader.KeySize+btreeOnHeapHeader.ValueSize)

	var properties []Property
	offset := int64(0)

	for i := 0; i < keyCount; i++ {
		// PropertyContextItem represents an item within the property context.
		// References "Property Context B-Tree-on-Heap Record".
		// References [MS-PDF]: 2.3.3.3 PC BTH Record
		var property Property

		// TODO - We can merge into a single ReadAt again.
		propertyID := make([]byte, 2)

		if _, err := keyTableReader.ReadAt(propertyID, offset); err != nil {
			return nil, eris.Wrap(err, "failed to read property ID")
		}

		propertyType := make([]byte, 2)

		if _, err := keyTableReader.ReadAt(propertyType, offset+2); err != nil {
			return nil, eris.Wrap(err, "failed to read property type")
		}

		data := make([]byte, 4)

		if _, err := keyTableReader.ReadAt(data, offset+4); err != nil {
			return nil, eris.Wrap(err, "failed to read data")
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
		return eris.Wrap(err, "failed to write MessagePack header")
	}

	for _, property := range propertyContext.Properties {
		propertyID := property.ID
		//propertyMap := PropertyMap[strconv.Itoa(int(propertyID))]
		//
		//if len(propertyMap) == 2 && propertyMap[1] != "" {
		//	// Get the propertyID from the Name-To-ID map.
		//	//propertyIDFromNameToIDMap, err := propertyContext.File.NameToIDMap.GetPropertyID(int(propertyID), PropertySet(propertyMap[1]))
		//	//
		//	//if err != nil {
		//	//	return eris.WithStack(err)
		//	//}
		//	//
		//	//fmt.Printf("Got: %d\n", propertyIDFromNameToIDMap)
		//
		//	fmt.Printf("Got: %s\n", propertyMap[1])
		//
		//	// TODO - !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		//	// TODO - properties.csv keep track of PidLid, PidTag, PidName etc
		//	// TODO - logic based on that to lookup in Name-To-ID or not.
		//
		//	for propertySetIndex, propertySet := range PropertySets {
		//		if propertySet == propertyMap[1] {
		//			fmt.Printf("Finding: %s - %d\n", propertyMap[0], propertySetIndex)
		//
		//			propertyIDFromNameToIDMap, err := propertyContext.File.NameToIDMap.GetPropertyID(int(property.ID), PropertySet(propertySetIndex))
		//
		//			if err != nil {
		//				fromNameToID = true
		//				break
		//				//return eris.WithStack(err)
		//			}
		//		}
		//	}
		//
		//	//for t, _ := range propertyContext.File.NameToIDMap. {
		//	//	fmt.Printf("Got1: %s\n", t)
		//	//}
		//
		//	// TODO - NameToID has the GUID (string) -  PropertySet(propertyMap[1] to PropertyID
		//}

		// TODO - Map using JSON name instead of propertyID.
		propertyReader, err := propertyContext.GetPropertyReader(propertyID, localDescriptors)

		if mappedID, ok := propertyContext.File.NameToIDMap.IDToName[int(propertyID)]; ok {
			propertyReader.Property.ID = uint16(mappedID)
		}
		if err != nil && !eris.Is(err, ErrPropertyNoData) {
			return eris.Wrap(err, "failed to get property reader")
		}

		// TODO - This is a hot-path, optimize this.
		err = propertyReader.WriteMessagePackValue(messagePackWriter)

		if err != nil && !eris.Is(err, ErrPropertyNoData) {
			return eris.Wrap(err, "failed to write MessagePack value")
		}
	}

	if err := messagePackWriter.Flush(); err != nil {
		return eris.Wrap(err, "failed to flush MessagePack writer")
	}

	err := decodable.DecodeMsg(msgp.NewReader(messagePackBuffer))

	if err != nil && !eris.Is(err, io.EOF) {
		return eris.Wrap(err, "failed to decode message")
	}

	return nil
}

//go:embed properties.csv
var PropertyMapCSV string

// PropertyMap maps the property ID to the struct tag and property set (if any).
var PropertyMap = make(map[string][]string)

func init() {
	propertyMapReader := csv.NewReader(strings.NewReader(PropertyMapCSV))

	csvProperties, err := propertyMapReader.ReadAll()

	if err != nil {
		panic(eris.Wrap(err, "failed to initialize property map"))
	}

	for _, row := range csvProperties {
		PropertyMap[row[0]] = row[1:]
	}
}
