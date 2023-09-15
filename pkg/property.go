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
	"io"
)

// Property represents a property in the TableContext or PropertyContext.
// See PropertyReader, PropertyWriter.
type Property struct {
	Identifier uint16
	Type       PropertyType
	HNID       Identifier
	// Value is only used for small values.
	// <= 8 bytes for the Table Context.
	// <= 4 bytes for the Property Context.
	// Other values will use the HNID.
	Value []byte
}

// WriteTo writes the byte representation of the Property.
func (property *Property) WriteTo(writer io.Writer) (int64, error) {
	// TODO - We can't pass formatType because io.WriterTo signature doesn't allow it.
	propertyBuffer := bytes.NewBuffer(make([]byte, 2+2+identifierSize+property.Value.Len()))

	propertyBuffer.Write(GetUint16(property.Identifier))
	propertyBuffer.Write(property.Type.Bytes())
	propertyBuffer.Write(property.HNID.Bytes(formatType))
	propertyBuffer.Write(property.Value.Bytes())

	return propertyBuffer.WriteTo(writer)
}

// PropertyType represents the data type of the property.
type PropertyType uint16

func (propertyType PropertyType) Bytes() []byte {
	return GetUint16(uint16(propertyType))
}

// Constants defining the property types.
//
// References "Property types".
// References [MS-OXCDATA]: 2.11.1 Property Data Types
const (
	PropertyTypeInteger16            PropertyType = 2
	PropertyTypeInteger32            PropertyType = 3
	PropertyTypeFloating32           PropertyType = 4
	PropertyTypeFloating64           PropertyType = 5
	PropertyTypeCurrency             PropertyType = 6
	PropertyTypeFloatingTime         PropertyType = 7
	PropertyTypeErrorCode            PropertyType = 10
	PropertyTypeBoolean              PropertyType = 11
	PropertyTypeInteger64            PropertyType = 20
	PropertyTypeString               PropertyType = 31
	PropertyTypeString8              PropertyType = 30
	PropertyTypeTime                 PropertyType = 64
	PropertyTypeGUID                 PropertyType = 72
	PropertyTypeServerID             PropertyType = 251
	PropertyTypeRestriction          PropertyType = 253
	PropertyTypeRuleAction           PropertyType = 254
	PropertyTypeBinary               PropertyType = 258
	PropertyTypeMultipleInteger16    PropertyType = 4098
	PropertyTypeMultipleInteger32    PropertyType = 4099
	PropertyTypeMultipleFloating32   PropertyType = 4100
	PropertyTypeMultipleFloating64   PropertyType = 4101
	PropertyTypeMultipleCurrency     PropertyType = 4102
	PropertyTypeMultipleFloatingTime PropertyType = 4103
	PropertyTypeMultipleInteger64    PropertyType = 4116
	PropertyTypeMultipleString       PropertyType = 4127
	PropertyTypeMultipleString8      PropertyType = 4126
	PropertyTypeMultipleTime         PropertyType = 4160
	PropertyTypeMultipleGUID         PropertyType = 4168
	PropertyTypeMultipleBinary       PropertyType = 4354
	PropertyTypeUnspecified          PropertyType = 0
	PropertyTypeNull                 PropertyType = 1
	PropertyTypeObject               PropertyType = 13
)

// GetDataSize returns the size of the data (in bytes) stored for this property type or -1 for variable sized data.
//
// References [MS-OXCDATA]: 2.11.1 Property Data Types
func (propertyType PropertyType) GetDataSize() int {
	switch propertyType {
	case PropertyTypeInteger16:
		return 2
	case PropertyTypeInteger32:
		return 4
	case PropertyTypeFloating32:
		return 4
	case PropertyTypeFloating64:
		return 8
	case PropertyTypeCurrency:
		return 8
	case PropertyTypeFloatingTime:
		return 8
	case PropertyTypeErrorCode:
		return 4
	case PropertyTypeBoolean:
		return 1
	case PropertyTypeInteger64:
		return 8
	case PropertyTypeTime:
		return 8
	case PropertyTypeGUID:
		return 16
	default:
		return -1
	}
}
