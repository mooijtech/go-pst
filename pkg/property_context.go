// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
)

// Constants defining the property types.
// References "Property types".
const (
	PropertyTypeInteger16 = 2
	PropertyTypeInteger32 = 3
	PropertyTypeFloating32 = 4
	PropertyTypeFloating64 = 5
	PropertyTypeCurrency = 6
	PropertyTypeFloatingTime = 7
	PropertyTypeErrorCode = 10
	PropertyTypeBoolean = 11
	PropertyTypeInteger64 = 20
	PropertyTypeString = 31
	PropertyTypeString8 = 30
	PropertyTypeTime = 64
	PropertyTypeGUID = 72
	PropertyTypeServerID = 251
	PropertyTypeRestriction = 253
	PropertyTypeRuleAction = 254
	PropertyTypeBinary = 258
	PropertyTypeMultipleInteger16 = 4098
	PropertyTypeMultipleInteger32 = 4099
	PropertyTypeMultipleFloating32 = 4100
	PropertyTypeMultipleFloating64 = 4101
	PropertyTypeMultipleCurrency = 4102
	PropertyTypeMultipleFloatingTime = 4103
	PropertyTypeMultipleInteger64 = 4116
	PropertyTypeMultipleString = 4127
	PropertyTypeMultipleString8 = 4126
	PropertyTypeMultipleTime = 4160
	PropertyTypeMultipleGUID = 4168
	PropertyTypeMultipleBinary = 4354
	PropertyTypeUnspecified = 0
	PropertyTypeNull = 1
	PropertyTypeObject = 13
)

// PropertyContextItem represents an item within the property context.
// References "Property Context B-Tree-on-Heap Record".
type PropertyContextItem struct {
	Index int
	PropertyID int
	PropertyType int
	ReferenceHNID int
}

// GetPropertyContext returns the property context (BC Table).
// References "Property Context".
func (pstFile *File) GetPropertyContext(btreeNodeEntryHeapOnNode BTreeNodeEntry, formatType string) ([]PropertyContextItem, error) {
	if btreeNodeEntryHeapOnNode.GetTableType() != 188 {
		// Must be Property Context.
		return nil, errors.New("invalid table type for property context")
	}

	btreeOnHeapHeader, err := pstFile.GetBTreeOnHeapHeader(btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return nil, err
	}

	allocationTableOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(btreeOnHeapHeader.HIDRoot, btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return nil, err
	}

	keyTable := btreeNodeEntryHeapOnNode.Data[allocationTableOffsets.StartOffset:allocationTableOffsets.EndOffset]
	keyCount := len(keyTable) / (btreeOnHeapHeader.KeySize + btreeOnHeapHeader.ValueSize)

	var propertyContextItems []PropertyContextItem
	offset := 0

	for i := 0; i < keyCount; i++ {
		propertyID := int(binary.LittleEndian.Uint16(keyTable[offset:offset + 2]))
		propertyType := int(binary.LittleEndian.Uint16(keyTable[offset + 2:offset + 4]))
		referenceHNID := int(binary.LittleEndian.Uint16(keyTable[offset + 4:offset + 8]))

		propertyContextItems = append(propertyContextItems, PropertyContextItem {
			Index: i,
			PropertyID: propertyID,
			PropertyType: propertyType,
			ReferenceHNID: referenceHNID,
		})

		offset = offset + 8
	}

	return propertyContextItems, nil
}

// GetPropertyContextItem returns the property context item from the property ID.
// References GetPropertyContext.
func (pstFile *File) GetPropertyContextItem(propertyContext []PropertyContextItem, propertyID int) PropertyContextItem {
	var propertyContextItem PropertyContextItem

	for _, item := range propertyContext {
		if item.PropertyID == propertyID {
			propertyContextItem = item
			break
		}
	}

	return propertyContextItem
}