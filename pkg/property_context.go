// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
)

// Constants defining the property types.
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
type PropertyContextItem struct {
	Index int
	EntryType int
	EntryValueType int
	EntryValueReference int
}

// GetPropertyContext returns the property context (BC Table).
func (pstFile *File) GetPropertyContext(btreeNodeEntryHeapOnNode BTreeNodeEntry, formatType string) error {
	if btreeNodeEntryHeapOnNode.GetHeapOnNodeTableType() != 188 {
		// Must be Property Context.
		return errors.New("invalid table type for property context")
	}

	btreeOnHeapHeader, err := pstFile.GetBTreeOnHeapHeader(btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return err
	}

	allocationTableOffsets, err := pstFile.GetHeapOnNodeAllocationTableOffsets(btreeOnHeapHeader.HIDRoot, btreeNodeEntryHeapOnNode, formatType)

	if err != nil {
		return err
	}

	keyTable := btreeNodeEntryHeapOnNode.Data[allocationTableOffsets.StartOffset:allocationTableOffsets.EndOffset]
	keyCount := len(keyTable) / (btreeOnHeapHeader.KeySize + btreeOnHeapHeader.ValueSize)

	offset := 0

	for i := 0; i < keyCount; i++ {
		log.Infof("Entry type: %d", binary.LittleEndian.Uint16(keyTable[offset:offset + 2]))
		log.Infof("Entry value type: %d", binary.LittleEndian.Uint16(keyTable[offset + 2:offset + 4]))

		offset = offset + 8
	}

	return nil
}