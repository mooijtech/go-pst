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
	"crypto/rand"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	"io"
)

// GetNodeBTreeOffset returns the file offset to the node b-tree.
// References TODO
func (file *File) GetNodeBTreeOffset() (int64, error) {
	var outputBuffer []byte
	var offset int64

	switch file.FormatType {
	case FormatTypeUnicode:
		outputBuffer = make([]byte, 8)
		offset = 224
	case FormatTypeUnicode4k:
		outputBuffer = make([]byte, 8)
		offset = 224
	case FormatTypeANSI:
		outputBuffer = make([]byte, 4)
		offset = 188
	default:
		return 0, ErrFormatTypeUnsupported
	}

	if _, err := file.Reader.ReadAt(outputBuffer, offset); err != nil {
		return 0, eris.Wrap(err, "failed to read node b-tree offset")
	}

	switch file.FormatType {
	case FormatTypeANSI:
		return int64(binary.LittleEndian.Uint32(outputBuffer)), nil
	default:
		return int64(binary.LittleEndian.Uint64(outputBuffer)), nil
	}
}

// GetBlockBTreeOffset returns the file offset to the block b-tree.
// References TODO
func (file *File) GetBlockBTreeOffset() (int64, error) {
	var outputBuffer []byte
	var offset int64

	switch file.FormatType {
	case FormatTypeUnicode4k:
		outputBuffer = make([]byte, 8)
		offset = 240
	case FormatTypeUnicode:
		outputBuffer = make([]byte, 8)
		offset = 240
	case FormatTypeANSI:
		outputBuffer = make([]byte, 4)
		offset = 196
	default:
		return 0, ErrFormatTypeUnsupported
	}

	if _, err := file.Reader.ReadAt(outputBuffer, offset); err != nil {
		return 0, eris.Wrap(err, "failed to read block b-tree offset")
	}

	switch file.FormatType {
	case FormatTypeANSI:
		return int64(binary.LittleEndian.Uint32(outputBuffer)), nil
	default:
		return int64(binary.LittleEndian.Uint64(outputBuffer)), nil
	}
}

// GetBTreeNodeEntryCount returns the amount of entries in the b-tree.
// References TODO
func (file *File) GetBTreeNodeEntryCount(btreeNode []byte) uint16 {
	switch file.FormatType {
	case FormatTypeUnicode4k:
		// References https://web.archive.org/web/20160528150307/https://blog.mythicsoft.com/2015/07/10/ost-2013-file-format-the-missing-documentation/
		return binary.LittleEndian.Uint16(btreeNode[4056:4058])
	case FormatTypeUnicode:
		return uint16(btreeNode[488])
	case FormatTypeANSI:
		return uint16(btreeNode[496])
	default:
		panic(ErrFormatTypeUnsupported)
	}
}

// GetBTreeNodeEntrySize returns the size of an entry in the b-tree.
// References TODO
func (file *File) GetBTreeNodeEntrySize(btreeNode []byte) uint8 {
	switch file.FormatType {
	case FormatTypeUnicode4k:
		// References TODO
		return btreeNode[4060]
	case FormatTypeUnicode:
		return btreeNode[490]
	case FormatTypeANSI:
		return btreeNode[498]
	default:
		panic(ErrFormatTypeUnsupported)
	}
}

// GetBTreeNodeLevel returns the level of the b-tree node.
// References "The node and block b-tree" (TODO).
func (file *File) GetBTreeNodeLevel(btreeNode []byte) uint8 {
	switch file.FormatType {
	case FormatTypeUnicode4k:
		// References TODO
		return btreeNode[4062]
	case FormatTypeUnicode:
		return btreeNode[491]
	case FormatTypeANSI:
		return btreeNode[499]
	default:
		panic(ErrFormatTypeUnsupported)
	}
}

// BTreeNode represents an entry in a b-tree node.
// Fields are set depending on the NodeLevel (branch or leaf).
// TODO - See NewBTreeNodeBranch and NewBTreeNodeLeaf.
type BTreeNode struct {
	// Identifier is only unique to the node level.
	Identifier                 Identifier `json:"identifier"`
	FileOffset                 int64      `json:"fileOffset"`
	DataIdentifier             Identifier `json:"dataIdentifier"`
	LocalDescriptorsIdentifier Identifier `json:"localDescriptorsIdentifier"`
	Size                       uint16     `json:"size"`
	NodeLevel                  uint8      `json:"nodeLevel"`

	// These variables are from the new OST format which uses ZLib.
	// References https://web.archive.org/web/20160528150307/https://blog.mythicsoft.com/2015/07/10/ost-2013-file-format-the-missing-documentation/
	CompressedSize   uint16 `json:"compressedSize"`
	DecompressedSize uint16 `json:"decompressedSize"`
}

// NewBTreeNodeBranch creates a new BTreeNode with a NodeLevel > 0
// References
//func NewBTreeNodeBranch(identifier Identifier) BTreeNode {
//	return BTreeNode{
//		Identifier: identifier,
//	}
//}
//
//// NewBTreeNodeLeaf creates a new BTreeNode leaf with a NodeLevel == 0
//// References
//func NewBTreeNodeLeaf(identifier Identifier) BTreeNode {
//	return BTreeNode{
//		Identifier: identifier,
//		NodeLevel:  0,
//	}
//}

// NewBTreeNodeReader is used by the HeapOnNode.
func NewBTreeNodeReader(btreeNode BTreeNode, reader Reader) *io.SectionReader {
	return io.NewSectionReader(reader, btreeNode.FileOffset, int64(btreeNode.Size))
}

// BTreeNodeLessFunc tests whether the current node is less than the given argument.
//
// This must provide a strict weak ordering.
// If !a.Less(b) && !b.Less(a), we treat this to mean a == b (we can only hold one of either a or b in the tree).
func BTreeNodeLessFunc(a BTreeNode, b BTreeNode) bool {
	if a.Identifier == b.Identifier {
		// We don't return the first identifier because there may be two or more nodes with
		// the same identifier, so prefer leaf nodes (which there is always only one of).
		return a.NodeLevel < b.NodeLevel
	} else {
		return a.Identifier < b.Identifier
	}
}

// GetBTreeNodeRawEntries returns the raw b-tree node entries in bytes.
// References https://github.com/mooijtech/go-pst/blob/master/docs/README.md#btpage
// Used by GetBTreeNodeEntries.
func (file *File) GetBTreeNodeRawEntries(btreeNodeOffset int64, callback func([]byte, error)) {
	var outputBuffer []byte

	switch file.FormatType {
	case FormatTypeUnicode4k:
		// References https://web.archive.org/web/20160528150307/https://blog.mythicsoft.com/2015/07/10/ost-2013-file-format-the-missing-documentation/
		outputBuffer = make([]byte, 4056)
	case FormatTypeUnicode:
		outputBuffer = make([]byte, 512)
	case FormatTypeANSI:
		outputBuffer = make([]byte, 512)
	default:
		callback(nil, ErrFormatTypeUnsupported)
	}

	if _, err := file.Reader.ReadAtAsync(outputBuffer, uint64(btreeNodeOffset), func(err error) {
		callback(outputBuffer, err)
	}); err != nil {
		callback(nil, err)
	}
}

// GetBTreeNodeEntryCompressedSize is only used for Unicode 4k.
func GetBTreeNodeEntryCompressedSize(btreeNodeEntryData []byte) uint16 {
	return binary.LittleEndian.Uint16(btreeNodeEntryData[16 : 16+2])
}

// GetBTreeNodeEntryDecompressedSize is only used for Unicode 4k.
func GetBTreeNodeEntryDecompressedSize(btreeNodeEntryData []byte) uint16 {
	return binary.LittleEndian.Uint16(btreeNodeEntryData[18 : 18+2])
}

// GetBTreeNodeEntries returns the entries in the b-tree node.
// References TODO
func (file *File) GetBTreeNodeEntries(btreeNodeOffset int64, btreeType BTreeType, callback func(btreeNodeEntries []BTreeNode, nodeLevel uint8, err error)) {
	parentBTreeNodeLevel, err := file.GetParentBTreeNodeLevel(btreeNodeOffset)

	if err != nil {
		callback(nil, 0, errors.WithStack(err))
	}

	file.GetBTreeNodeRawEntries(btreeNodeOffset, func(btreeNodeEntry []byte, err error) {
		if err != nil {
			callback(nil, 0, err)
			return
		}

		btreeNodeEntryCount := file.GetBTreeNodeEntryCount(btreeNodeEntry)
		//btreeNodeLevel := file.GetBTreeNodeLevel(btreeNodeEntry) // Of children.
		btreeNodeEntrySize := file.GetBTreeNodeEntrySize(btreeNodeEntry)
		btreeNodeEntries := make([]BTreeNode, btreeNodeEntryCount)

		for i := 0; i < int(btreeNodeEntryCount); i++ {
			btreeNodeEntryData := btreeNodeEntry[i*int(btreeNodeEntrySize) : (i*int(btreeNodeEntrySize))+int(btreeNodeEntrySize)]

			if parentBTreeNodeLevel > 0 && (btreeType == BTreeTypeNode || btreeType == BTreeTypeBlock) {
				// Branch node or block b-tree node.
				btreeNodeEntries[i] = BTreeNode{
					Identifier: GetBTreeNodeEntryIdentifier(btreeNodeEntryData, file.FormatType),
					FileOffset: GetBTreeNodeEntryFileOffset(btreeNodeEntryData, true, file.FormatType),
					NodeLevel:  parentBTreeNodeLevel,
				}
			} else if parentBTreeNodeLevel == 0 && btreeType == BTreeTypeNode {
				// Leaf node b-tree node.
				btreeNodeEntries[i] = BTreeNode{
					Identifier:                 GetBTreeNodeEntryIdentifier(btreeNodeEntryData, file.FormatType),
					DataIdentifier:             GetBTreeNodeEntryDataIdentifier(btreeNodeEntryData, file.FormatType),
					LocalDescriptorsIdentifier: GetBTreeNodeEntryLocalDescriptorsIdentifier(btreeNodeEntryData, file.FormatType),
					NodeLevel:                  parentBTreeNodeLevel,
				}
			} else if parentBTreeNodeLevel == 0 && btreeType == BTreeTypeBlock {
				// Leaf block b-tree node.
				btreeNodeEntries[i] = BTreeNode{
					Identifier: GetBTreeNodeEntryIdentifier(btreeNodeEntryData, file.FormatType),
					FileOffset: GetBTreeNodeEntryFileOffset(btreeNodeEntryData, false, file.FormatType),
					Size:       GetBTreeNodeEntrySize(btreeNodeEntryData, file.FormatType),
					NodeLevel:  parentBTreeNodeLevel,
				}
			}

			// Unicode 4k is used by OST and is the new format which supports ZLib.
			// ZLib support is handled by ZLibDecompressor which is used by the HeapOnNodeReader.
			if file.FormatType == FormatTypeUnicode4k {
				btreeNodeEntries[i].CompressedSize = GetBTreeNodeEntryCompressedSize(btreeNodeEntryData)
				btreeNodeEntries[i].DecompressedSize = GetBTreeNodeEntryDecompressedSize(btreeNodeEntryData)
			}
		}

		// TODO - Use channels.
		callback(btreeNodeEntries, parentBTreeNodeLevel, nil)
	})
}

// GetBTreeNodeEntryIdentifier returns the Identifier of this b-tree node entry.
// References "The b-tree entries" (TODO).
func GetBTreeNodeEntryIdentifier(btreeNodeEntryData []byte, formatType FormatType) Identifier {
	return GetIdentifierFromBytes(btreeNodeEntryData[:GetIdentifierSize(formatType)])
}

// GetIdentifierFromBytes returns the Identifier type from bytes.
func GetIdentifierFromBytes(identifierBytes []byte) Identifier {
	return Identifier(binary.LittleEndian.Uint32(identifierBytes))
}

// GetIdentifierSize returns the size of an Identifier.
func GetIdentifierSize(formatType FormatType) uint8 {
	switch formatType {
	case FormatTypeANSI:
		return 4
	default:
		return 8
	}
}

// Identifier represents a b-tree node identifier.
// TODO - Document the int types per use case and use separate types.
// Used by B-Tree nodes.
type Identifier int64

// NewIdentifier creates a new Identifier.
// Used by the writer so the BTreeNode can be identified.
func NewIdentifier(formatType FormatType) (Identifier, error) {
	var identifierSize int

	switch formatType {
	case FormatTypeUnicode4k:
		identifierSize = 8
	case FormatTypeUnicode:
		identifierSize = 8
	case FormatTypeANSI:
		identifierSize = 4
	default:
		return 0, ErrFormatTypeUnsupported
	}

	identifierBytes := make([]byte, identifierSize)

	if _, err := rand.Read(identifierBytes); err != nil {
		return 0, eris.Wrap(err, "failed to read random bytes using crypto/rand")
	}

	return Identifier(binary.LittleEndian.Uint64(identifierBytes)), nil
}

// Constants defining the special b-tree node identifiers.
const (
	IdentifierRootFolder                Identifier = 290
	IdentifierMessageStore              Identifier = 33
	IdentifierNameToIDMap               Identifier = 97
	IdentifierNormalFolderTemplate      Identifier = 161
	IdentifierSearchFolderTemplate      Identifier = 193
	IdentifierSearchManagementQueue     Identifier = 481
	IdentifierSearchActivityList        Identifier = 513
	IdentifierReserved1                 Identifier = 577
	IdentifierSearchDomainObject        Identifier = 609
	IdentifierSearchGathererQueue       Identifier = 641
	IdentifierSearchGathererDescriptor  Identifier = 673
	IdentifierReserved2                 Identifier = 737
	IdentifierReserved3                 Identifier = 769
	IdentifierSearchGathererFolderQueue Identifier = 801
)

// GetType returns the IdentifierType of this Identifier.
func (identifier Identifier) GetType() IdentifierType {
	// Bit-masking:
	// Use bitwise ANDing in order to extract a subset of the bits in the value.
	// 11111 (binary) = 0x1F (hex), which with bitwise ANDing extracts the first 5 bits.
	// See: https://www.rapidtables.com/convert/number/binary-to-hex.html
	return IdentifierType(identifier & 0x1F)
}

// WriteTo writes the byte representation of the identifier.
func (identifier Identifier) WriteTo(writer io.Writer, formatType FormatType) (int, error) {
	return writer.Write(identifier.Bytes(formatType))
}

// Bytes returns the byte representation of the pst.Identifier.
func (identifier Identifier) Bytes(formatType FormatType) []byte {
	switch formatType {
	case FormatTypeUnicode4k, FormatTypeUnicode:
		return GetUint64(uint64(identifier))
	case FormatTypeANSI:
		return GetUint32(uint32(identifier))
	default:
		panic(ErrFormatTypeUnsupported)
	}
}

// IdentifierType represents the type of Identifier.
type IdentifierType uint8

// Constants defining the identifier types.
// References "Identifier types" (TODO).
const (
	IdentifierTypeHID                     IdentifierType = 0
	IdentifierTypeInternal                IdentifierType = 1
	IdentifierTypeNormalFolder            IdentifierType = 2
	IdentifierTypeSearchFolder            IdentifierType = 3
	IdentifierTypeNormalMessage           IdentifierType = 4
	IdentifierTypeAttachment              IdentifierType = 5
	IdentifierTypeSearchUpdateQueue       IdentifierType = 6
	IdentifierTypeSearchCriteriaObject    IdentifierType = 7
	IdentifierTypeAssociatedMessage       IdentifierType = 8
	IdentifierTypeContentsTableIndex      IdentifierType = 10
	IdentifierTypeReceiveFolderTable      IdentifierType = 11
	IdentifierTypeOutgoingQueueTable      IdentifierType = 12
	IdentifierTypeHierarchyTable          IdentifierType = 13
	IdentifierTypeContentsTable           IdentifierType = 14
	IdentifierTypeAssociatedContentsTable IdentifierType = 15
	IdentifierTypeSearchContentsTable     IdentifierType = 16
	IdentifierTypeAttachmentTable         IdentifierType = 17
	IdentifierTypeRecipientTable          IdentifierType = 18
	IdentifierTypeSearchTableIndex        IdentifierType = 19
	IdentifierTypeLTP                     IdentifierType = 31
)

// GetBTreeNodeEntryFileOffset returns the file offset for this b-tree branch or leaf node.
// References "The b-tree entries" (TODO).
func GetBTreeNodeEntryFileOffset(btreeNodeEntryData []byte, isBranchNode bool, formatType FormatType) int64 {
	if isBranchNode {
		switch formatType {
		case FormatTypeANSI:
			return int64(binary.LittleEndian.Uint32(btreeNodeEntryData[8 : 8+4]))
		default:
			return int64(binary.LittleEndian.Uint64(btreeNodeEntryData[16 : 16+8]))
		}
	} else {
		switch formatType {
		case FormatTypeANSI:
			return int64(binary.LittleEndian.Uint32(btreeNodeEntryData[4 : 4+4]))
		default:
			return int64(binary.LittleEndian.Uint64(btreeNodeEntryData[8 : 8+8]))
		}
	}
}

// GetBTreeNodeEntryDataIdentifier returns the node identifier of the data (in the block b-tree).
// References "The b-tree entries" (TODO).
func GetBTreeNodeEntryDataIdentifier(btreeNodeEntryData []byte, formatType FormatType) Identifier {
	switch formatType {
	case FormatTypeANSI:
		return GetIdentifierFromBytes(btreeNodeEntryData[4 : 4+GetIdentifierSize(formatType)])
	default:
		return GetIdentifierFromBytes(btreeNodeEntryData[8 : 8+GetIdentifierSize(formatType)])
	}
}

// GetBTreeNodeEntryLocalDescriptorsIdentifier returns the Identifier to the local descriptors in the block b-tree.
func GetBTreeNodeEntryLocalDescriptorsIdentifier(btreeNodeEntryData []byte, formatType FormatType) Identifier {
	switch formatType {
	case FormatTypeANSI:
		return GetIdentifierFromBytes(btreeNodeEntryData[8 : 8+GetIdentifierSize(formatType)])
	default:
		return GetIdentifierFromBytes(btreeNodeEntryData[16 : 16+GetIdentifierSize(formatType)])
	}
}

// GetBTreeNodeEntrySize returns the size of the data in the block b-tree leaf node entry.
// References "The b-tree entries" TODO References
func GetBTreeNodeEntrySize(btreeNodeEntryData []byte, formatType FormatType) uint16 {
	switch formatType {
	case FormatTypeANSI:
		return binary.LittleEndian.Uint16(btreeNodeEntryData[8 : 8+2])
	default:
		return binary.LittleEndian.Uint16(btreeNodeEntryData[16 : 16+2])
	}
}

// BTreeType represents either the node b-tree or block b-tree.
type BTreeType byte

// Constants defining the b-tree types.
const (
	BTreeTypeNode  BTreeType = 129
	BTreeTypeBlock BTreeType = 128
)

// GetParentBTreeNodeLevel returns the level of the b-tree node.
// References "The node and block b-tree".
func (file *File) GetParentBTreeNodeLevel(btreeNodeOffset int64) (uint8, error) {
	outputBuffer := make([]byte, 1)
	var offset int64

	switch file.FormatType {
	case FormatTypeUnicode:
		offset = btreeNodeOffset + 491
	case FormatTypeUnicode4k:
		offset = btreeNodeOffset + 4061
	case FormatTypeANSI:
		offset = btreeNodeOffset + 499
	default:
		return 0, errors.WithStack(ErrFormatTypeUnsupported)
	}

	if _, err := file.Reader.ReadAt(outputBuffer, offset); err != nil {
		return 0, errors.WithStack(err)
	}

	return outputBuffer[0], nil
}

// WalkAndCreateBTree walks the b-tree and updates the given b-tree store.
func (file *File) WalkAndCreateBTree(btreeOffset int64, btreeType BTreeType, btreeStore BTreeStore) {
	file.GetBTreeNodeEntries(btreeOffset, btreeType, func(nodeEntries []BTreeNode, nodeLevel uint8, err error) {
		if nodeLevel > 0 {
			// Branch node entries.
			// TODO - Use channels and Goroutines per walk route branch.
			// TODO - Linux I/O URing
			// TODO - Align to Linux 4096 bytes.
			for i := 0; i < len(nodeEntries); i++ {
				nodeEntry := nodeEntries[i]

				if _, exists := btreeStore.Load(nodeEntry); exists {
					// TODO - *errgroup.Group so we can use errors properly.
					panic(errors.WithStack(ErrBTreeNodeConflict))
				}

				file.WalkAndCreateBTree(nodeEntry.FileOffset, btreeType, btreeStore)
			}
		} else {
			// Leaf node entries
			for i := 0; i < len(nodeEntries); i++ {
				if _, exists := btreeStore.Load(nodeEntries[i]); exists {
					// TODO - *errgroup.Group so we can use errors properly.
					panic(errors.WithStack(ErrBTreeNodeConflict))
				}
			}
		}
	})
}

// GetBTreeNode returns the node in the node or block b-tree with the given identifier.
func (file *File) GetBTreeNode(identifier Identifier, btreeStore BTreeStore) (BTreeNode, error) {
	btreeNode, found := btreeStore.Get(BTreeNode{Identifier: identifier})

	if !found {
		return BTreeNode{}, ErrBTreeNodeNotFound
	}

	return btreeNode, nil
}

// GetNodeBTreeNode returns the node with the given identifier in the node b-tree.
func (file *File) GetNodeBTreeNode(identifier Identifier) (BTreeNode, error) {
	return file.GetBTreeNode(identifier, file.NodeBTree)
}

// GetBlockBTreeNode returns the node with the given identifier in the block b-tree.
func (file *File) GetBlockBTreeNode(identifier Identifier) (BTreeNode, error) {
	// Clear the least significant bit (LSB), which is reserved, but sometimes set.
	return file.GetBTreeNode(identifier&0xfffffffe, file.BlockBTree)
}

// GetDataBTreeNode searches the identifier in the node b-tree, then searches the data identifier in the block b-tree.
func (file *File) GetDataBTreeNode(identifier Identifier) (BTreeNode, error) {
	nodeBTreeNode, err := file.GetNodeBTreeNode(identifier)

	if err != nil {
		return BTreeNode{}, eris.Wrap(err, "failed to get b-tree node")
	}

	return file.GetBlockBTreeNode(nodeBTreeNode.DataIdentifier)
}
