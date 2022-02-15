// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Defines the property sets (GUIDs) in the Name-To-ID Map.
// "Property set: A GUID that identifies a group of properties with a similar purpose."
// References [MS-OXPROPS].pdf "1.3.2 Commonly Used Property Sets"
var (
	PropertySets = []string{
		"00020329-0000-0000-C000-000000000046",
		"00062008-0000-0000-C000-000000000046",
		"00062004-0000-0000-C000-000000000046",
		"00020386-0000-0000-C000-000000000046",
		"00062002-0000-0000-C000-000000000046",
		"6ED8DA90-450B-101B-98DA-00AA003F1305",
		"0006200A-0000-0000-C000-000000000046",
		"41F28F13-83F4-4114-A584-EEDB5A6B0BFF",
		"0006200E-0000-0000-C000-000000000046",
		"00062041-0000-0000-C000-000000000046",
		"00062003-0000-0000-C000-000000000046",
		"4442858E-A9E3-4E80-B900-317A210CC15B",
		"00020328-0000-0000-C000-000000000046",
		"71035549-0739-4DCB-9163-00F0580DBBDF",
		"00062040-0000-0000-C000-000000000046",
		"23239608-685D-4732-9C55-4C95CB4E8E33",
		"96357F7F-59E1-47D0-99A7-46515C183B54",
	}
)

// Constants defining the commonly used property sets.
var (
	PropertySetPublicStrings = 0
	PropertySetCommon = 1
	PropertySetAddress = 2
	PropertySetInternetHeaders = 3
	PropertySetAppointment = 4
	PropertySetMeeting = 5
	PropertySetLog = 6
	PropertySetMessaging = 7
	PropertySetNote = 8
	PropertySetPostRSS = 9
	PropertySetTask = 10
	PropertySetUnifiedMessaging = 11
	PropertySetMAPI = 12
	PropertySetAirSync = 13
	PropertySetSharing = 14
	PropertySetXMLExtractedEntities = 15
	PropertySetAttachment = 16
)

// NameToIDMap represents the Name-To-ID Map.
type NameToIDMap struct {
	PropertySets []string
	NameToID map[int]int
	IDToName map[int]int
	StringToID map[string]int
	IDToString map[int]string
}

// InitializeNameToIDMap initializes the Name-To-ID Map.
func (pstFile *File) InitializeNameToIDMap(formatType string, encryptionType string) error {
	nameToIDMap, err := pstFile.GetNameToIDMap(formatType, encryptionType)

	if err != nil {
		return err
	}

	pstFile.NameToIDMap = nameToIDMap

	return nil
}

// GetNameToIDMap returns the Name-To-ID Map.
func (pstFile *File) GetNameToIDMap(formatType string, encryptionType string) (NameToIDMap, error) {
	nameToIDMapNode, err := pstFile.GetNodeBTreeNode(IdentifierTypeNameToIDMap, formatType)

	if err != nil {
		return NameToIDMap{}, err
	}

	localDescriptors, err := pstFile.GetLocalDescriptors(nameToIDMapNode, formatType)

	if err != nil {
		return NameToIDMap{}, err
	}

	nameToIDMapNodeDataIdentifier, err := nameToIDMapNode.GetDataIdentifier(formatType)

	if err != nil {
		return NameToIDMap{}, err
	}

	blockBTreeNode, err := pstFile.GetBlockBTreeNode(nameToIDMapNodeDataIdentifier, formatType)

	if err != nil {
		return NameToIDMap{}, err
	}

	heapOnNode, err := pstFile.NewHeapOnNodeFromNode(blockBTreeNode, formatType, encryptionType)

	if err != nil {
		return NameToIDMap{}, err
	}

	propertyContext, err := pstFile.GetPropertyContext(heapOnNode, formatType, encryptionType)

	if err != nil {
		return NameToIDMap{}, err
	}

	propertySetToIndex := make(map[string]int)

	for i, propertySet := range PropertySets {
		propertySetToIndex[propertySet] = i
	}

	// References [MS-PST].pdf "2.4.7.2 GUID Stream"
	// References [MS-PST].pdf "2.1.2 Properties"
	// The GUID Stream is a flat array of 16-byte GUID values that contains the GUIDs associated with all
	// the property sets used in all the named properties in the PST.
	guidStream, err := FindPropertyContextItem(propertyContext, 2)

	if err != nil {
		return NameToIDMap{}, err
	}

	guidStreamData, err := guidStream.GetData(pstFile, localDescriptors, formatType, encryptionType)

	if err != nil {
		return NameToIDMap{}, err
	}

	guidCount := len(guidStreamData) / 16

	var nameToIDMap NameToIDMap

	offset := 0

	var guids []string
	guidIndexes := make([]int, guidCount)

	for i := 0; i < guidCount; i++ {
		// Convert bytes to GUID.
		// References https://github.com/microsoft/go-winio/blob/master/pkg/guid/guid.go
		var guidBytes [16]byte

		copy(guidBytes[:], guidStreamData[offset:offset + 16])

		guid := GUIDFromWindowsArray(guidBytes)

		containsPropertySet := false

		for _, propertySet := range PropertySets {
			if strings.ToLower(propertySet) == guid.String() {
				guids = append(guids, propertySet)
				guidIndexes[i] = propertySetToIndex[propertySet]

				containsPropertySet = true
				break
			}
		}

		// Every GUID index (the GUID Stream) must map to a GUID, specify this index has no GUID.
		if !containsPropertySet {
			guidIndexes[i] = -1
		}

		offset += 16
	}

	// References [MS-PST].pdf "2.4.7.3 Entry Stream"
	// References [MS-PST].pdf "2.1.2 Properties"
	// The Entry Stream is a flat array of NAMEID records that represent all the named properties in the PST.
	entryStream, err := FindPropertyContextItem(propertyContext, 3)

	if err != nil {
		return NameToIDMap{}, err
	}

	entryStreamData, err := entryStream.GetData(pstFile, localDescriptors, formatType, encryptionType)

	if err != nil {
		return NameToIDMap{}, err
	}

	// References "2.4.7.4  The String Stream"
	// References [MS-PST].pdf "2.1.2 Properties"
	// The String Stream is a packed list of strings that is used for all the named properties in the PST.
	stringStream, err := FindPropertyContextItem(propertyContext, 4)

	if err != nil {
		return NameToIDMap{}, err
	}

	stringStreamData, err := stringStream.GetData(pstFile, localDescriptors, formatType, encryptionType)

	if err != nil {
		return NameToIDMap{}, err
	}

	// Process the Entry Stream (NAMEID records).
	// References "2.4.7.1 NAMEID"
	nameToID := make(map[int]int)
	idToName := make(map[int]int)
	stringToID := make(map[string]int)
	idToString := make(map[int]string)

	for i := 0; i + 8 <= len(entryStreamData); i += 8 {
		namedPropertyIdentifier := binary.LittleEndian.Uint16(entryStreamData[i:i + 4])
		namedPropertyGUID := binary.LittleEndian.Uint16(entryStreamData[i + 4:i+6]) // The first bit should be ignored which is the named property identifier type.
		namedPropertyIdentifierType := namedPropertyGUID & 0x0001
		namedPropertyIndex := binary.LittleEndian.Uint16(entryStreamData[i + 6:i + 8]) // Property index. This is the ordinal number of the named property, which is used to calculate the NPID of this named property.

		if namedPropertyIdentifierType == 0 {
			// If the named property identifier type is 0, the named property identifier contains the value of a numerical name.
			namedPropertyGUID >>= 1

			// The named property ID is calculated by adding 0x8000 to property index.
			namedPropertyIndex += 0x8000

			var propertySetIndex int

			switch namedPropertyGUID {
			case 1:
				propertySetIndex = PropertySetMAPI
				break
			case 2:
				propertySetIndex = PropertySetPublicStrings
				break
			default:
				propertySetIndex = guidIndexes[int(namedPropertyGUID) - 3]
				break
			}

			nameToID[int(namedPropertyIdentifier) | propertySetIndex << 32] = int(namedPropertyIndex)
			idToName[int(namedPropertyIndex)] = int(namedPropertyIdentifier)
		} else {
			// If the named property identifier type is 1, this value is the byte offset into the String stream in
			// which the string name of the property is stored.
			keyLength := binary.LittleEndian.Uint16(stringStreamData[int(namedPropertyIdentifier):int(namedPropertyIdentifier) + 4])

			if int(keyLength) > 0 && int(keyLength) < len(stringStreamData) {
				key, err := DecodeBytesToUTF16String(stringStreamData[int(namedPropertyIdentifier) + 4:int(namedPropertyIdentifier) + 4 + int(keyLength)])

				if err != nil {
					return NameToIDMap{}, err
				}

				namedPropertyIndex += 0x8000

				stringToID[key] = int(namedPropertyIndex)
				idToString[int(namedPropertyIndex)] = key
			}
		}
	}

	nameToIDMap.PropertySets = guids
	nameToIDMap.NameToID = nameToID
	nameToIDMap.IDToName = idToName
	nameToIDMap.StringToID = stringToID
	nameToIDMap.IDToString = idToString

	return nameToIDMap, nil
}

// GetPropertyID returns the Name-To-ID property ID.
func (nameToIDMap *NameToIDMap) GetPropertyID(key int, propertySetIndex int) (int, error) {
	nameToIDKey := propertySetIndex << 32 | key

	value, ok := nameToIDMap.NameToID[nameToIDKey]

	if !ok {
		return -1, errors.New("failed to find key in Name-To-ID Map")
	}

	return value, nil
}

// GUID represents a GUID/UUID. It has the same structure as
// golang.org/x/sys/windows.GUID so that it can be used with functions expecting
// that type. It is defined as its own type as that is only available to builds
// targeted at `windows`. The representation matches that used by native Windows
// code.
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// GUIDFromWindowsArray constructs a GUID from a Windows encoding array of bytes.
func GUIDFromWindowsArray(b [16]byte) GUID {
	var g GUID

	g.Data1 = binary.LittleEndian.Uint32(b[0:4])
	g.Data2 = binary.LittleEndian.Uint16(b[4:6])
	g.Data3 = binary.LittleEndian.Uint16(b[6:8])

	copy(g.Data4[:], b[8:16])

	return g
}

func (g GUID) String() string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		g.Data1,
		g.Data2,
		g.Data3,
		g.Data4[:2],
		g.Data4[2:])
}