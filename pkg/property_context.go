// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	_ "embed"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"
	"time"
)

// Constants defining the property types.
// References "Property types".
const (
	PropertyTypeInteger16            = 2
	PropertyTypeInteger32            = 3
	PropertyTypeFloating32           = 4
	PropertyTypeFloating64           = 5
	PropertyTypeCurrency             = 6
	PropertyTypeFloatingTime         = 7
	PropertyTypeErrorCode            = 10
	PropertyTypeBoolean              = 11
	PropertyTypeInteger64            = 20
	PropertyTypeString               = 31
	PropertyTypeString8              = 30
	PropertyTypeTime                 = 64
	PropertyTypeGUID                 = 72
	PropertyTypeServerID             = 251
	PropertyTypeRestriction          = 253
	PropertyTypeRuleAction           = 254
	PropertyTypeBinary               = 258
	PropertyTypeMultipleInteger16    = 4098
	PropertyTypeMultipleInteger32    = 4099
	PropertyTypeMultipleFloating32   = 4100
	PropertyTypeMultipleFloating64   = 4101
	PropertyTypeMultipleCurrency     = 4102
	PropertyTypeMultipleFloatingTime = 4103
	PropertyTypeMultipleInteger64    = 4116
	PropertyTypeMultipleString       = 4127
	PropertyTypeMultipleString8      = 4126
	PropertyTypeMultipleTime         = 4160
	PropertyTypeMultipleGUID         = 4168
	PropertyTypeMultipleBinary       = 4354
	PropertyTypeUnspecified          = 0
	PropertyTypeNull                 = 1
	PropertyTypeObject               = 13
)

// PropertyContextItem represents an item within the property context.
// References "Property Context B-Tree-on-Heap Record".
type PropertyContextItem struct {
	Index                    int
	PropertyID               int
	PropertyType             int
	ReferenceHNID            int
	IsExternalValueReference bool
	data                     []byte
}

// GetPropertyContext returns the property context (BC Table).
// References "Property Context".
func (pstFile *File) GetPropertyContext(heapOnNode HeapOnNode, formatType string, encryptionType string) ([]PropertyContextItem, error) {
	tableType, err := heapOnNode.GetTableType()

	if err != nil {
		return nil, err
	}

	if tableType != 188 {
		// Must be Property Context.
		return nil, errors.New("invalid table type for property context")
	}

	btreeOnHeapHeader, err := pstFile.GetBTreeOnHeapHeader(heapOnNode, []LocalDescriptor{}, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	keyTableInputStream, err := pstFile.NewHeapOnNodeInputStreamFromHNID(btreeOnHeapHeader.HIDRoot, heapOnNode, []LocalDescriptor{}, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	keyCount := keyTableInputStream.Size / (btreeOnHeapHeader.KeySize + btreeOnHeapHeader.ValueSize)

	var propertyContextItems []PropertyContextItem
	offset := 0

	for i := 0; i < keyCount; i++ {
		var propertyContextItem PropertyContextItem

		propertyID, err := keyTableInputStream.SeekAndReadUint16(2, offset)

		if err != nil {
			return nil, err
		}

		propertyType, err := keyTableInputStream.SeekAndReadUint16(2, offset+2)

		if err != nil {
			return nil, err
		}

		referenceHNID, err := keyTableInputStream.SeekAndReadUint32(4, offset+4)

		if err != nil {
			return nil, err
		}

		propertyContextItem.Index = i
		propertyContextItem.PropertyID = propertyID
		propertyContextItem.PropertyType = propertyType
		propertyContextItem.ReferenceHNID = referenceHNID

		switch propertyType {
		case PropertyTypeInteger16:
			propertyContextItem.ReferenceHNID &= 0xFFFF
			break
		case PropertyTypeInteger32:
			break
		case PropertyTypeNull:
			break
		case PropertyTypeBoolean:
			propertyContextItem.ReferenceHNID &= 0xFF
			propertyContextItem.IsExternalValueReference = true
			break
		default:
			propertyContextItem.IsExternalValueReference = true

			propertyNodeInputStream, err := pstFile.NewHeapOnNodeInputStreamFromHNID(referenceHNID, heapOnNode, []LocalDescriptor{}, formatType, encryptionType)

			if err != nil {
				// External node.
				// The data is in the local descriptors (when the HNID matches the local descriptor identifier) which
				// gives us a data identifier that points to a node in the block b-tree (Heap-on-Node).
				// We deal with this when getting the value of a property (like GetMessageString).
				break
			}

			propertyContextItem.IsExternalValueReference = false

			propertyContextItemData, err := propertyNodeInputStream.Read(propertyNodeInputStream.Size, 0)

			if err != nil {
				return nil, err
			}

			propertyContextItem.data = propertyContextItemData
		}

		propertyContextItems = append(propertyContextItems, propertyContextItem)
		offset = offset + 8
	}

	return propertyContextItems, nil
}

// FindPropertyContextItem returns the property context item from the property ID.
func FindPropertyContextItem(propertyContext []PropertyContextItem, propertyID int) (PropertyContextItem, error) {
	for _, propertyContextItem := range propertyContext {
		if propertyContextItem.PropertyID == propertyID {
			return propertyContextItem, nil
		}
	}

	return PropertyContextItem{}, errors.New("failed to find property context item")
}

// GetData returns all the data of the property context item.
// Used for property type binary (external reference) which may be located in the local descriptors.
func (propertyContextItem *PropertyContextItem) GetData(pstFile *File, localDescriptors []LocalDescriptor, formatType string, encryptionType string) ([]byte, error) {
	if len(propertyContextItem.data) != 0 {
		return propertyContextItem.data, nil
	}

	if len(localDescriptors) == 0 {
		return nil, errors.New("external reference but no local descriptors")
	}

	if propertyContextItem.PropertyType != PropertyTypeBinary {
		return nil, errors.New("attempting to get non-binary data")
	}

	localDescriptor, err := FindLocalDescriptor(localDescriptors, propertyContextItem.ReferenceHNID, formatType)

	if err != nil {
		return nil, err
	}

	localDescriptorData, err := localDescriptor.GetData(pstFile, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	return localDescriptorData, nil
}

// GetString returns the string value of the property context item.
func (propertyContextItem *PropertyContextItem) GetString(encoding Encoding, localDescriptors []LocalDescriptor, pstFile *File, formatType string, encryptionType string) (string, error) {
	if !propertyContextItem.IsExternalValueReference {
		if propertyContextItem.PropertyID == 4096 || propertyContextItem.PropertyID == 4115 { // Only the message body uses the specified encoding as far as I know.
			return DecodeBytesToString(encoding, propertyContextItem.data)
		} else {
			return DecodeBytesToUTF16String(propertyContextItem.data)
		}
	} else {
		// External value reference (data is stored in a separate node).
		localDescriptor, err := FindLocalDescriptor(localDescriptors, propertyContextItem.ReferenceHNID, formatType)

		if err != nil {
			return "", err
		}

		data, err := localDescriptor.GetData(pstFile, formatType, encryptionType)

		if err != nil {
			return "", err
		}

		return DecodeBytesToString(encoding, data)
	}
}

// GetInteger returns the integer value of the property context item.
func (propertyContextItem *PropertyContextItem) GetInteger() int {
	return propertyContextItem.ReferenceHNID
}

// GetDate returns the date value of the property context item.
func (propertyContextItem *PropertyContextItem) GetDate() time.Time {
	// References https://stackoverflow.com/a/57903746
	dateInteger := binary.LittleEndian.Uint64(propertyContextItem.data)

	t := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	d := time.Duration(dateInteger)

	for i := 0; i < 100; i++ {
		t = t.Add(d)
	}

	return t
}

//go:embed properties.csv
var properties string

// Property represents a property.
type Property struct {
	Name string
	ID int
}

// GetProperties returns all available properties.
func GetProperties() ([]Property, error) {
	csvReader := csv.NewReader(strings.NewReader(properties))

	csvProperties, err := csvReader.ReadAll()

	if err != nil {
		return nil, err
	}

	var properties []Property

	for _, propertyRow := range csvProperties {
		// https://pkg.go.dev/strconv#ParseInt
		propertyID, err := strconv.ParseInt(strings.Replace(propertyRow[1], "0x", "", 1), 16, 64)

		if err != nil {
			continue
		}

		properties = append(properties, Property {
			Name: propertyRow[0],
			ID: int(propertyID),
		})
	}

	return properties, nil
}

// FindProperty finds the property from the property ID.
func FindProperty(propertyID int) (Property, error) {
	properties, err := GetProperties()

	if err != nil {
		return Property{}, err
	}

	for _, property := range properties {
		if propertyID == property.ID {
			return property, nil
		}
	}

	return Property{}, errors.New("failed to find property")
}

// String returns the string representation of this property.
func (propertyContextItem *PropertyContextItem) String() (string, error) {
	property, err := FindProperty(propertyContextItem.PropertyID)

	if err != nil {
		return "", errors.New("failed to find property")
	}

	return property.Name, nil
}
