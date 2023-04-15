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
	"encoding/binary"
	"fmt"
	"github.com/tinylib/msgp/msgp"
	"math"
	"time"

	"github.com/rotisserie/eris"
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
func NewPropertyReader(property Property, heapOnNode *HeapOnNode, file *File, localDescriptors []LocalDescriptor) (PropertyReader, error) {
	switch {
	case property.HNID != 0 && len(property.Data) == 0:
		heapOnNodeReader, err := file.GetHeapOnNodeReaderFromHNID(property.HNID, *heapOnNode.Reader, localDescriptors...)

		if err != nil {
			return PropertyReader{}, eris.Wrap(err, "failed to get Heap-on-Node reader")
		}

		return PropertyReader{property, heapOnNodeReader, localDescriptors, file}, nil
	case len(property.Data) != 0:
		return PropertyReader{property, nil, localDescriptors, file}, nil
	default:
		return PropertyReader{}, eris.Wrap(ErrPropertyNoData, fmt.Sprintf("Property ID: %x", property.ID))
	}
}

// WriteMessagePackValue writes the Message Pack format of the property value.
// Used to populate struct fields.
func (propertyReader *PropertyReader) WriteMessagePackValue(writer *msgp.Writer) error {
	key := fmt.Sprintf("%d%d", propertyReader.Property.ID, propertyReader.Property.Type)

	switch propertyReader.Property.Type {
	case PropertyTypeString:
		value, err := propertyReader.GetString()

		if err != nil {
			return eris.Wrap(err, "failed to get string")
		}

		if err := writer.WriteString(key); err != nil {
			return eris.Wrap(err, "failed to write key")
		} else if err := writer.WriteString(value); err != nil {
			return eris.Wrap(err, "failed to write value")
		}

		return nil
	case PropertyTypeString8:
		value, err := propertyReader.GetString8(65001) // TODO - Get from called, this is UTF-8 for now.

		if err != nil {
			return eris.Wrap(err, "failed to get string8")
		}

		if err := writer.WriteString(key); err != nil {
			return eris.Wrap(err, "failed to write key")
		} else if err := writer.WriteString(value); err != nil {
			return eris.Wrap(err, "failed to write value")
		}

		return nil
	case PropertyTypeTime:
		value, err := propertyReader.GetDate()

		if err != nil {
			return eris.Wrap(err, "failed to get date")
		}

		if err := writer.WriteString(key); err != nil {
			return eris.Wrap(err, "failed to write key")
		} else if err := writer.WriteInt64(value); err != nil {
			return eris.Wrap(err, "failed to write value")
		}

		return nil
	case PropertyTypeInteger16:
		value, err := propertyReader.GetInteger16()

		if err != nil {
			return eris.Wrap(err, "failed to get integer16")
		}

		if err := writer.WriteString(key); err != nil {
			return eris.Wrap(err, "failed to write key")
		} else if err := writer.WriteInt16(value); err != nil {
			return eris.Wrap(err, "failed to write value")
		}

		return nil
	case PropertyTypeInteger32:
		value, err := propertyReader.GetInteger32()

		if err != nil {
			return eris.Wrap(err, "failed to write integer32")
		}

		if err := writer.WriteString(key); err != nil {
			return eris.Wrap(err, "failed to write key")
		} else if err := writer.WriteInt32(value); err != nil {
			return eris.Wrap(err, "failed to write value")
		}

		return nil
	case PropertyTypeInteger64:
		value, err := propertyReader.GetInteger64()

		if err != nil {
			return eris.Wrap(err, "failed to get integer64")
		}

		if err := writer.WriteString(key); err != nil {
			return eris.Wrap(err, "failed to write key")
		} else if err := writer.WriteInt64(value); err != nil {
			return eris.Wrap(err, "failed to write value")
		}

		return nil
	default:
		// TODO - Write Nil?
		return ErrPropertyNoData
	}
}

// GetString returns the string value of the property.
func (propertyReader *PropertyReader) GetString() (string, error) {
	if propertyReader.Property.Type != PropertyTypeString {
		return "", ErrPropertyTypeMismatch
	} else if propertyReader.HeapOnNodeReader == nil || propertyReader.Property.HNID == 0 {
		return "", ErrPropertyNoData
	}

	data := make([]byte, propertyReader.Size())

	if _, err := propertyReader.ReadAt(data, 0); err != nil {
		return "", eris.Wrap(err, "failed to read data")
	}

	return propertyReader.DecodeString(data)
}

// GetString8 returns the string using the external encoding.
func (propertyReader *PropertyReader) GetString8(codepageIdentifier int) (string, error) {
	if propertyReader.Property.Type != PropertyTypeString8 {
		return "", ErrPropertyTypeMismatch
	} else if propertyReader.HeapOnNodeReader == nil || propertyReader.Property.HNID == 0 {
		return "", ErrPropertyNoData
	}

	data := make([]byte, propertyReader.Size())

	if _, err := propertyReader.ReadAt(data, 0); err != nil {
		return "", eris.Wrap(err, "failed to read data")
	}

	return propertyReader.DecodeString8(data, codepageIdentifier)
}

// GetInteger16 returns the 16-bit integer value of the property.
func (propertyReader *PropertyReader) GetInteger16() (int16, error) {
	switch {
	case propertyReader.Property.Type != PropertyTypeInteger16:
		return 0, ErrPropertyTypeMismatch
	case len(propertyReader.Property.Data) == 0:
		return 0, ErrPropertyNoData
	default:
		return int16(binary.LittleEndian.Uint16(propertyReader.Property.Data)), nil
	}
}

// GetInteger32 returns the 32-bit integer value of the property.
func (propertyReader *PropertyReader) GetInteger32() (int32, error) {
	switch {
	case propertyReader.Property.Type != PropertyTypeInteger32:
		return 0, ErrPropertyTypeMismatch
	case len(propertyReader.Property.Data) == 0:
		return 0, ErrPropertyNoData
	default:
		return int32(binary.LittleEndian.Uint32(propertyReader.Property.Data)), nil
	}
}

// GetInteger64 returns the 64-bit integer value of the property.
func (propertyReader *PropertyReader) GetInteger64() (int64, error) {
	switch {
	case propertyReader.Property.Type != PropertyTypeInteger64:
		return 0, ErrPropertyTypeMismatch
	case len(propertyReader.Property.Data) != 0:
		return int64(binary.LittleEndian.Uint64(propertyReader.Property.Data)), nil
	case propertyReader.Property.HNID != 0:
		value := make([]byte, 8)

		if _, err := propertyReader.ReadAt(value, 0); err != nil {
			return 0, eris.Wrap(err, "failed to read integer64")
		}

		return int64(binary.LittleEndian.Uint64(value)), nil
	default:
		return 0, ErrPropertyNoData
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
		return "", eris.Wrapf(err, "failed to find IANA encoding: %d", codePageIdentifier)
	}

	return encoding.NewDecoder().String(string(data))
}

// GetDate returns the date value (Unix Nano epoch) of the property context item.
func (propertyReader *PropertyReader) GetDate() (int64, error) {
	if propertyReader.Property.Type != PropertyTypeTime {
		return 0, ErrPropertyTypeMismatch
	} else if propertyReader.Size() == 0 {
		return 0, ErrPropertyNoData
	}

	outputBuffer := make([]byte, 8)

	if _, err := propertyReader.ReadAt(outputBuffer, 0); err != nil {
		return 0, eris.Wrap(err, "failed to read date")
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
		return false, ErrPropertyTypeMismatch
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
