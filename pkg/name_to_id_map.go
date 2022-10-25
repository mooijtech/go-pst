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
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/unicode"
)

// PropertySets defines the property sets (GUIDs) in the Name-To-ID Map.
// "Property set: A GUID that identifies a group of properties with a similar purpose."
// References [MS-OXPROPS]: "1.3.2 Commonly Used Property Sets".
var PropertySets = []string{
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

// PropertySet represents a collection of properties.
type PropertySet uint8

// Constants defining the commonly used property sets.
const (
	PropertySetPublicStrings PropertySet = iota
	PropertySetCommon
	PropertySetAddress
	PropertySetInternetHeaders
	PropertySetAppointment
	PropertySetMeeting
	PropertySetLog
	PropertySetMessaging
	PropertySetNote
	PropertySetPostRSS
	PropertySetTask
	PropertySetUnifiedMessaging
	PropertySetMAPI
	PropertySetAirSync
	PropertySetSharing
	PropertySetXMLExtractedEntities
	PropertySetAttachment
)

// NameToIDMap represents the Name-To-ID Map.
type NameToIDMap struct {
	PropertySets []string
	NameToID     map[int]int
	IDToName     map[int]int
	StringToID   map[string]int
	IDToString   map[int]string
}

// GetNameToIDMap returns the Name-To-ID Map.
func (file *File) GetNameToIDMap() (*NameToIDMap, error) {
	nodeBTreeNode, err := file.GetNodeBTreeNode(IdentifierNameToIDMap)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	localDescriptors, err := file.GetLocalDescriptors(nodeBTreeNode)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	blockBTreeNode, err := file.GetBlockBTreeNode(nodeBTreeNode.DataIdentifier)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	heapOnNode, err := file.GetHeapOnNode(blockBTreeNode)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	propertyContext, err := file.GetPropertyContext(heapOnNode)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	propertySetToIndex := make(map[string]int, len(PropertySets))

	for i, propertySet := range PropertySets {
		propertySetToIndex[propertySet] = i
	}

	// References [MS-PST].pdf "2.4.7.2 GUID Stream"
	// References [MS-PST].pdf "2.1.2 Properties"
	// The GUID Stream is a flat array of 16-byte GUID values that contains the GUIDs associated with all
	// the property sets used in all the named properties in the PST.
	guidReader, err := propertyContext.GetPropertyReader(2, localDescriptors...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	guidCount := int(guidReader.Size() / 16)

	offset := int64(0)

	var guids []string
	guidIndexes := make([]int, guidCount)

	for i := 0; i < guidCount; i++ {
		guidBytes := make([]byte, 16)

		if _, err := guidReader.ReadAt(guidBytes, offset); err != nil {
			return nil, errors.WithStack(err)
		}

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
	entryStream, err := propertyContext.GetPropertyReader(3, localDescriptors...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	entryStreamData := make([]byte, entryStream.Size())

	if _, err := entryStream.ReadAt(entryStreamData, 0); err != nil {
		return nil, errors.WithStack(err)
	}

	// References "2.4.7.4  The String Stream"
	// References [MS-PST].pdf "2.1.2 Properties"
	// The String Stream is a packed list of strings that is used for all the named properties in the PST.
	stringStream, err := propertyContext.GetPropertyReader(4, localDescriptors...)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	stringStreamData := make([]byte, stringStream.Size())

	if _, err := stringStream.ReadAt(stringStreamData, 0); err != nil {
		return nil, errors.WithStack(err)
	}

	// Process the Entry Stream (NAMEID records).
	// References "2.4.7.1 NAMEID"
	nameToID := make(map[int]int)
	idToName := make(map[int]int)
	stringToID := make(map[string]int)
	idToString := make(map[int]string)

	for i := 0; i+8 <= len(entryStreamData); i += 8 {
		namedPropertyIdentifier := binary.LittleEndian.Uint16(entryStreamData[i : i+4])
		namedPropertyGUID := binary.LittleEndian.Uint16(entryStreamData[i+4 : i+6]) // The first bit should be ignored which is the named property identifier type.
		namedPropertyIdentifierType := namedPropertyGUID & 0x0001
		namedPropertyIndex := binary.LittleEndian.Uint16(entryStreamData[i+6 : i+8]) // Property index. This is the ordinal number of the named property, which is used to calculate the NPID of this named property.

		if namedPropertyIdentifierType == 0 {
			// If the named property identifier type is 0, the named property identifier contains the value of a numerical name.
			namedPropertyGUID >>= 1

			// The named property ID is calculated by adding 0x8000 to property index.
			namedPropertyIndex += 0x8000

			var propertySetIndex int

			switch namedPropertyGUID {
			case 1:
				propertySetIndex = int(PropertySetMAPI)
			case 2:
				propertySetIndex = int(PropertySetPublicStrings)
			default:
				propertySetIndex = guidIndexes[int(namedPropertyGUID)-3]
			}

			nameToID[int(namedPropertyIdentifier)|propertySetIndex<<32] = int(namedPropertyIndex)
			idToName[int(namedPropertyIndex)] = int(namedPropertyIdentifier)
		} else {
			// If the named property identifier type is 1, this value is the byte offset into the String stream in
			// which the string name of the property is stored.
			keyLength := binary.LittleEndian.Uint32(stringStreamData[int(namedPropertyIdentifier) : int(namedPropertyIdentifier)+4])

			if int(keyLength) > 0 && int(keyLength) < len(stringStreamData) {
				utf16Decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

				key, err := utf16Decoder.String(string(stringStreamData[int(namedPropertyIdentifier)+4 : int(namedPropertyIdentifier)+4+int(keyLength)]))

				if err != nil {
					return nil, errors.WithStack(err)
				}

				namedPropertyIndex += 0x8000

				stringToID[key] = int(namedPropertyIndex)
				idToString[int(namedPropertyIndex)] = key
			}
		}
	}

	return &NameToIDMap{
		PropertySets: guids,
		NameToID:     nameToID,
		IDToName:     idToName,
		StringToID:   stringToID,
		IDToString:   idToString,
	}, nil
}

// GetPropertyID returns the Name-To-ID property ID.
func (nameToIDMap *NameToIDMap) GetPropertyID(key int, propertySet PropertySet) (int, error) {
	nameToIDKey := int(propertySet)<<32 | key

	value, found := nameToIDMap.NameToID[nameToIDKey]

	if !found {
		return -1, errors.WithStack(ErrNameToIDMapKeyNotFound)
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
// References https://github.com/microsoft/go-winio/blob/master/pkg/guid/guid.go
func GUIDFromWindowsArray(b []byte) GUID {
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
