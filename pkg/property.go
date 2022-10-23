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

// Property represents a property in the PropertyContext or TableContext.
// See PropertyReader.
type Property struct {
	ID   uint16
	Type PropertyType
	HNID Identifier
	// Data is only used for small values.
	// <= 8 bytes for the Table Context.
	// <= 4 bytes for the Property Context.
	// Other values will use the HNID.
	Data []byte
}

// PropertyType represents the data type of the property.
type PropertyType uint16

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
