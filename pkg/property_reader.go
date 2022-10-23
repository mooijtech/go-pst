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
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/unicode"
)

// PropertyReader allows reading a property from the Property/Table Context.
// References [MS-OXCDATA]: 2.11.1 Property Data Types.
// Implements io.SectionReader.
type PropertyReader struct {
	Property         Property
	HeapOnNodeReader *HeapOnNodeReader
	LocalDescriptors []LocalDescriptor
	File             *File
}

// NewPropertyReader creates a new property reader.
func NewPropertyReader(property Property, heapOnNode *HeapOnNode, file *File, localDescriptors ...LocalDescriptor) (PropertyReader, error) {
	switch {
	case property.HNID != 0 && len(property.Data) == 0:
		heapOnNodeReader, err := file.GetHeapOnNodeReaderFromHNID(property.HNID, *heapOnNode.Reader, localDescriptors...)

		if err != nil {
			return PropertyReader{}, errors.WithStack(err)
		}

		return PropertyReader{property, heapOnNodeReader, localDescriptors, file}, nil
	case len(property.Data) != 0:
		return PropertyReader{property, nil, localDescriptors, file}, nil
	default:
		return PropertyReader{}, errors.WithStack(errors.WithMessage(ErrPropertyNoData, fmt.Sprintf("Property ID: %x", property.ID)))
	}
}

// GetValue returns any value based on the property type.
func (propertyReader *PropertyReader) GetValue() (any, error) {
	switch propertyReader.Property.Type {
	case PropertyTypeString:
		return propertyReader.GetString()
	// TODO - PropertyTypeString8
	case PropertyTypeTime:
		return propertyReader.GetDate()
	case PropertyTypeInteger16:
		return propertyReader.GetInteger16()
	case PropertyTypeInteger32:
		return propertyReader.GetInteger32()
	case PropertyTypeInteger64:
		return propertyReader.GetInteger64()
	default:
		return nil, ErrPropertyNoData
	}
}

// GetString returns the string value of the property.
func (propertyReader *PropertyReader) GetString() (string, error) {
	if propertyReader.Property.Type != PropertyTypeString {
		return "", errors.WithStack(ErrPropertyTypeMismatch)
	} else if propertyReader.HeapOnNodeReader == nil || propertyReader.Property.HNID == 0 {
		return "", errors.WithStack(ErrPropertyNoData)
	}

	data := make([]byte, propertyReader.HeapOnNodeReader.Size())

	if _, err := propertyReader.ReadAt(data, 0); err != nil {
		return "", errors.WithStack(err)
	}

	return propertyReader.DecodeString(data)
}

// GetString8 returns the string using the external encoding.
func (propertyReader *PropertyReader) GetString8(codepageIdentifier int) (string, error) {
	if propertyReader.Property.Type != PropertyTypeString8 {
		return "", errors.WithStack(ErrPropertyTypeMismatch)
	} else if propertyReader.HeapOnNodeReader == nil || propertyReader.Property.HNID == 0 {
		return "", errors.WithStack(ErrPropertyNoData)
	}

	data := make([]byte, propertyReader.Size())

	if _, err := propertyReader.ReadAt(data, 0); err != nil {
		return "", errors.WithStack(err)
	}

	return propertyReader.DecodeString8(data, codepageIdentifier)
}

// GetInteger16 returns the 16-bit integer value of the property.
func (propertyReader *PropertyReader) GetInteger16() (int16, error) {
	switch {
	case propertyReader.Property.Type != PropertyTypeInteger16:
		return 0, errors.WithStack(errors.WithMessage(ErrPropertyTypeMismatch, fmt.Sprintf("Property type: %d", propertyReader.Property.Type)))
	case len(propertyReader.Property.Data) == 0:
		return 0, errors.WithStack(ErrPropertyNoData)
	default:
		return int16(binary.LittleEndian.Uint16(propertyReader.Property.Data)), nil
	}
}

// GetInteger32 returns the 32-bit integer value of the property.
func (propertyReader *PropertyReader) GetInteger32() (int32, error) {
	switch {
	case propertyReader.Property.Type != PropertyTypeInteger32:
		return 0, errors.WithStack(ErrPropertyTypeMismatch)
	case len(propertyReader.Property.Data) == 0:
		return 0, errors.WithStack(ErrPropertyNoData)
	default:
		return int32(binary.LittleEndian.Uint32(propertyReader.Property.Data)), nil
	}
}

// GetInteger64 returns the 64-bit integer value of the property.
func (propertyReader *PropertyReader) GetInteger64() (int64, error) {
	switch {
	case propertyReader.Property.Type != PropertyTypeInteger64:
		return 0, errors.WithStack(ErrPropertyTypeMismatch)
	case len(propertyReader.Property.Data) != 0:
		return int64(binary.LittleEndian.Uint64(propertyReader.Property.Data)), nil
	case propertyReader.Property.HNID != 0:
		//heapOnNodeReader, err := propertyReader.File.GetHeapOnNodeReaderFromHNID(propertyReader.Property.HNID, *propertyReader.HeapOnNodeReader, propertyReader.LocalDescriptors...)
		//
		//if err != nil {
		//	return 0, errors.WithStack(err)
		//}

		value := make([]byte, 8)

		if _, err := propertyReader.ReadAt(value, 0); err != nil {
			return 0, errors.WithStack(err)
		}

		return int64(binary.LittleEndian.Uint64(value)), nil
	default:
		return 0, errors.WithStack(ErrPropertyNoData)
	}
}

// DecodeString decodes the PropertyTypeString using Unicode (UTF-16LE).
// References [MS-OXCDATA]: 2.11.1 Property Data Types.
func (propertyReader *PropertyReader) DecodeString(data []byte) (string, error) {
	utf16Decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

	return utf16Decoder.String(string(data))
}

// DecodeString8 decodes the PropertyTypeString8 using the code page identifier.
func (propertyReader *PropertyReader) DecodeString8(data []byte, codePageIdentifier int) (string, error) {
	encoding, err := ianaindex.IANA.Encoding(CodePageIdentifierToEncoding[codePageIdentifier])

	if err != nil {
		return "", errors.WithStack(err)
	}

	return encoding.NewDecoder().String(string(data))
}

// // GetEncoding returns the encoding of the message.
// TODO -
// func (message *Message) GetEncoding() (Encoding, error) {
// 	encoding, err := message.GetInteger(16381) // PidTagMessageCodepage

// 	if err != nil {
// 		encoding, err = message.GetInteger(26307) // PidTagCodepage

// 		if err != nil {
// 			encoding, err = message.GetInteger(16350) // PidTagInternetCodepage

// 			if err != nil {
// 				// Encoding is set to -1
// 			}
// 		}
// 	}

// 	if encoding != -1 {
// 		// Found the encoding identifier.
// 		foundEncoding, err := FindEncoding(encoding)

// 		if err != nil {
// 			fmt.Printf("Unsupported encoding (%d), please open an issue on GitHub to support this encoding. Defaulting to UTF-8.\n", encoding)

// 			return Encoding{
// 				Identifier: 65001,
// 				Name:       "utf-8",
// 			}, nil
// 		}

// 		return foundEncoding, nil
// 	} else {
// 		// TODO - Lookup the global encoding in the Message Store.
// 		fmt.Printf("Failed to find message encoding, defaulting to UTF-8.\n")

// 		return Encoding{
// 			Identifier: 65001,
// 			Name:       "utf-8",
// 		}, nil
// 	}
// }

// GetDate returns the date value (Unix Nano epoch) of the property context item.
func (propertyReader *PropertyReader) GetDate() (int64, error) {
	if propertyReader.Property.Type != PropertyTypeTime {
		return 0, errors.WithStack(ErrPropertyTypeMismatch)
	} else if propertyReader.Size() == 0 {
		return 0, errors.WithStack(ErrPropertyNoData)
	}

	outputBuffer := make([]byte, 8)

	if _, err := propertyReader.ReadAt(outputBuffer, 0); err != nil {
		return 0, errors.WithStack(err)
	}

	// References https://stackoverflow.com/a/57903746
	// The number of 100-nanosecond intervals since January 1, 1601.
	input := int64(binary.LittleEndian.Uint64(outputBuffer))

	maxd := time.Duration(math.MaxInt64).Truncate(100 * time.Nanosecond)
	maxdUnits := int64(maxd / 100) // number of 100-ns units

	t := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)

	for input > maxdUnits {
		t = t.Add(maxd)
		input -= maxdUnits
	}

	if input != 0 {
		t = t.Add(time.Duration(input * 100))
	}

	return t.UnixNano(), nil
}

// GetBoolean returns the boolean value of this property.
func (propertyReader *PropertyReader) GetBoolean() (bool, error) {
	if propertyReader.Property.Type != PropertyTypeBoolean {
		return false, errors.WithStack(ErrPropertyTypeMismatch)
	}

	return propertyReader.Property.Data[0] == 1, nil
}

// ReadAt reads the underlying Heap-on-Node.
func (propertyReader *PropertyReader) ReadAt(outputBuffer []byte, offset int64) (int, error) {
	return propertyReader.HeapOnNodeReader.ReadAt(outputBuffer, offset)
}

// Size returns the size of the Heap-on-Node.
func (propertyReader *PropertyReader) Size() int64 {
	return propertyReader.HeapOnNodeReader.Size()
}
