// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"errors"
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
	Index         int
	PropertyID    int
	PropertyType  int
	ReferenceHNID int
}

// GetPropertyContext returns the property context (BC Table).
// References "Property Context".
func (pstFile *File) GetPropertyContext(heapOnNode HeapOnNode, localDescriptors []LocalDescriptor, formatType string, encryptionType string) ([]PropertyContextItem, error) {
	tableType, err := heapOnNode.GetTableType()

	if err != nil {
		return nil, err
	}

	if tableType != 188 {
		// Must be Property Context.
		return nil, errors.New("invalid table type for property context")
	}

	btreeOnHeapHeader, err := pstFile.GetBTreeOnHeapHeader(heapOnNode, localDescriptors, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	keyTableNodeInputStream, err := pstFile.GetAllocationTableNodeInputStream(btreeOnHeapHeader.HIDRoot, heapOnNode, localDescriptors, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	keyCount := keyTableNodeInputStream.Size / (btreeOnHeapHeader.KeySize + btreeOnHeapHeader.ValueSize)

	var propertyContextItems []PropertyContextItem
	offset := 0

	for i := 0; i < keyCount; i++ {
		propertyID, err := keyTableNodeInputStream.SeekAndReadUint16(2, offset)

		if err != nil {
			return nil, err
		}

		propertyType, err := keyTableNodeInputStream.SeekAndReadUint16(2, offset+2)

		if err != nil {
			return nil, err
		}

		referenceHNID, err := keyTableNodeInputStream.SeekAndReadUint32(4, offset+4)

		if err != nil {
			return nil, err
		}

		propertyContextItems = append(propertyContextItems, PropertyContextItem{
			Index:         i,
			PropertyID:    propertyID,
			PropertyType:  propertyType,
			ReferenceHNID: referenceHNID,
		})

		offset = offset + 8
	}

	return propertyContextItems, nil
}

// FindPropertyContextItem returns the property context item from the property ID.
// References GetPropertyContext.
func (pstFile *File) FindPropertyContextItem(propertyContext []PropertyContextItem, propertyID int) (PropertyContextItem, error) {
	for _, item := range propertyContext {
		if item.PropertyID == propertyID {
			return item, nil
		}
	}

	return PropertyContextItem{}, errors.New("failed to find property context item")
}
