// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
	"github.com/mooijtech/btree/v2"
)

// GetNodeBTreeOffset returns the file offset to the node b-tree.
// References "The 64-bit header data", "The 32-bit header data".
func (pstFile *File) GetNodeBTreeOffset(formatType string) (int, error) {
	var nodeBTreeFileOffset int
	var nodeBTreeBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		nodeBTreeFileOffset = 224
		nodeBTreeBufferSize = 8
		break
	case FormatTypeUnicode4k:
		nodeBTreeFileOffset = 224
		nodeBTreeBufferSize = 8
		break
	case FormatTypeANSI:
		nodeBTreeFileOffset = 188
		nodeBTreeBufferSize = 4
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	nodeBTreeOffset, err := pstFile.Read(nodeBTreeBufferSize, nodeBTreeFileOffset)

	if err != nil {
		return -1, err
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint64(nodeBTreeOffset)), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(nodeBTreeOffset)), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(nodeBTreeOffset)), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetBlockBTreeOffset returns the file offset to the block b-tree.
// References "The 64-bit header data", "The 32-bit header data".
func (pstFile *File) GetBlockBTreeOffset(formatType string) (int, error) {
	var blockBTreeFileOffset int
	var blockBTreeBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		blockBTreeFileOffset = 240
		blockBTreeBufferSize = 8
		break
	case FormatTypeUnicode4k:
		blockBTreeFileOffset = 240
		blockBTreeBufferSize = 8
		break
	case FormatTypeANSI:
		blockBTreeFileOffset = 196
		blockBTreeBufferSize = 4
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	blockBTreeOffset, err := pstFile.Read(blockBTreeBufferSize, blockBTreeFileOffset)

	if err != nil {
		return -1, err
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint64(blockBTreeOffset)), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(blockBTreeOffset)), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(blockBTreeOffset)), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetBTreeNodeEntryCount returns the amount of entries in the b-tree.
// References "The node and block b-tree".
func (pstFile *File) GetBTreeNodeEntryCount(btreeNodeOffset int, formatType string) (int, error) {
	var nodeEntryCountOffset int
	var nodeEntryCountBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		nodeEntryCountOffset = btreeNodeOffset + 488
		nodeEntryCountBufferSize = 1
		break
	case FormatTypeUnicode4k:
		nodeEntryCountOffset = btreeNodeOffset + 4056
		nodeEntryCountBufferSize = 2
		break
	case FormatTypeANSI:
		nodeEntryCountOffset = btreeNodeOffset + 496
		nodeEntryCountBufferSize = 1
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	nodeEntryCount, err := pstFile.Read(nodeEntryCountBufferSize, nodeEntryCountOffset)

	if err != nil {
		return -1, err
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint16([]byte{nodeEntryCount[0], 0})), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint16(nodeEntryCount)), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint16([]byte{nodeEntryCount[0], 0})), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetBTreeNodeEntrySize returns the size of an entry in the b-tree.
// References "The node and block b-tree".
func (pstFile *File) GetBTreeNodeEntrySize(btreeNodeOffset int, formatType string) (int, error) {
	var nodeEntrySizeOffset int

	switch formatType {
	case FormatTypeUnicode:
		nodeEntrySizeOffset = btreeNodeOffset + 490
		break
	case FormatTypeUnicode4k:
		nodeEntrySizeOffset = btreeNodeOffset + 4060
		break
	case FormatTypeANSI:
		nodeEntrySizeOffset = btreeNodeOffset + 498
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	nodeEntrySize, err := pstFile.Read(1, nodeEntrySizeOffset)

	if err != nil {
		return -1, err
	}

	return int(binary.LittleEndian.Uint16([]byte{nodeEntrySize[0], 0})), nil
}

// GetBTreeNodeLevel returns the level of the b-tree node.
// References "The node and block b-tree".
func (pstFile *File) GetBTreeNodeLevel(btreeNodeOffset int, formatType string) (int, error) {
	var nodeLevelOffset int

	switch formatType {
	case FormatTypeUnicode:
		nodeLevelOffset = btreeNodeOffset + 491
		break
	case FormatTypeUnicode4k:
		nodeLevelOffset = btreeNodeOffset + 4061
		break
	case FormatTypeANSI:
		nodeLevelOffset = btreeNodeOffset + 499
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	nodeLevel, err := pstFile.Read(1, nodeLevelOffset)

	if err != nil {
		return -1, err
	}

	return int(binary.LittleEndian.Uint16([]byte{nodeLevel[0], 0})), nil
}

// BTreeNodeEntry represents an entry in a b-tree node.
type BTreeNodeEntry struct {
	Identifier                 int
	IdentifierType             int
	FileOffset                 int
	DataIdentifier             int
	LocalDescriptorsIdentifier int
	Size                       int
	NodeLevel                  int
}

// NewBTreeNodeEntry creates a new b-tree node entry.
func NewBTreeNodeEntry(nodeEntryData []byte, formatType string, nodeLevel int) (BTreeNodeEntry, error) {
	identifier, err := GetBTreeNodeEntryIdentifier(nodeEntryData, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	identifierType, err := GetBTreeNodeEntryIdentifierType(nodeEntryData, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	fileOffset, err := GetBTreeNodeEntryFileOffset(nodeEntryData, formatType, nodeLevel > 0)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	dataIdentifier, err := GetBTreeNodeEntryDataIdentifier(nodeEntryData, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	localDescriptorsIdentifier, err := GetBTreeNodeEntryLocalDescriptorsIdentifier(nodeEntryData, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	size, err := GetBTreeNodeEntrySize(nodeEntryData, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	return BTreeNodeEntry{
		Identifier:                 identifier,
		IdentifierType:             identifierType,
		FileOffset:                 fileOffset,
		DataIdentifier:             dataIdentifier,
		LocalDescriptorsIdentifier: localDescriptorsIdentifier,
		Size:                       size,
		NodeLevel:                  nodeLevel,
	}, nil
}

// Less tests whether the current item is less than the given argument.
//
// This must provide a strict weak ordering.
// If !a.Less(b) && !b.Less(a), we treat this to mean a == b (i.e. we can only
// hold one of either a or b in the tree).
func (btreeNodeEntry BTreeNodeEntry) Less(than BTreeNodeEntry) bool {
	if btreeNodeEntry.Identifier == than.Identifier {
		// We don't return the first identifier because there may be two or more nodes with
		// the same identifier so prefer leaf nodes (which there is always only one of).
		return btreeNodeEntry.NodeLevel < than.NodeLevel
	} else {
		return btreeNodeEntry.Identifier < than.Identifier
	}
}

// GetBTreeNodeEntries returns the entries in the b-tree node.
// References "The node and block b-tree".
func (pstFile *File) GetBTreeNodeEntries(btreeNodeOffset int, formatType string, nodeLevel int) ([]BTreeNodeEntry, error) {
	var nodeEntriesBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		nodeEntriesBufferSize = 488
		break
	case FormatTypeUnicode4k:
		nodeEntriesBufferSize = 4056
		break
	case FormatTypeANSI:
		nodeEntriesBufferSize = 496
		break
	default:
		return nil, errors.New("unsupported format type")
	}

	nodeEntries, err := pstFile.Read(nodeEntriesBufferSize, btreeNodeOffset)

	if err != nil {
		return nil, err
	}

	nodeEntryCount, err := pstFile.GetBTreeNodeEntryCount(btreeNodeOffset, formatType)

	if err != nil {
		return nil, err
	}

	nodeEntrySize, err := pstFile.GetBTreeNodeEntrySize(btreeNodeOffset, formatType)

	if err != nil {
		return nil, err
	}

	entries := make([]BTreeNodeEntry, nodeEntryCount)

	for i := 0; i < nodeEntryCount; i++ {
		nodeEntryData := nodeEntries[(i * nodeEntrySize) : (i*nodeEntrySize)+nodeEntrySize]

		btreeNodeEntry, err := NewBTreeNodeEntry(nodeEntryData, formatType, nodeLevel)

		if err != nil {
			return nil, err
		}

		entries[i] = btreeNodeEntry
	}

	return entries, nil
}

// GetBTreeNodeEntryIdentifier returns the identifier of this b-tree node entry.
// References "The b-tree entries".
func GetBTreeNodeEntryIdentifier(nodeEntryData []byte, formatType string) (int, error) {
	var identifierBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		identifierBufferSize = 8
		break
	case FormatTypeUnicode4k:
		identifierBufferSize = 8
		break
	case FormatTypeANSI:
		identifierBufferSize = 4
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	return int(binary.LittleEndian.Uint32(nodeEntryData[:identifierBufferSize])), nil
}

// Constants defining the identifier types.
// References "Identifier types".
const (
	IdentifierTypeHID                     = 0
	IdentifierTypeInternal                = 1
	IdentifierTypeNormalFolder            = 2
	IdentifierTypeSearchFolder            = 3
	IdentifierTypeNormalMessage           = 4
	IdentifierTypeAttachment              = 5
	IdentifierTypeSearchUpdateQueue       = 6
	IdentifierTypeSearchCriteriaObject    = 7
	IdentifierTypeAssociatedMessage       = 8
	IdentifierTypeContentsTableIndex      = 10
	IdentifierTypeReceiveFolderTable      = 11
	IdentifierTypeOutgoingQueueTable      = 12
	IdentifierTypeHierarchyTable          = 13
	IdentifierTypeContentsTable           = 14
	IdentifierTypeAssociatedContentsTable = 15
	IdentifierTypeSearchContentsTable     = 16
	IdentifierTypeAttachmentTable         = 17
	IdentifierTypeRecipientTable          = 18
	IdentifierTypeSearchTableIndex        = 19
	IdentifierTypeLTP                     = 31

	IdentifierTypeRootFolder   = 290
	IdentifierTypeMessageStore = 33
	IdentifierTypeNameToIDMap  = 97
)

// GetBTreeNodeEntryIdentifierType returns the b-tree node entry identifier type.
// References "The b-tree entries", "Identifier".
func GetBTreeNodeEntryIdentifierType(nodeEntryData []byte, formatType string) (int, error) {
	identifier, err := GetBTreeNodeEntryIdentifier(nodeEntryData, formatType)

	if err != nil {
		return -1, err
	}

	// Bit masking:
	// Use bitwise ANDing in order to extract a subset of the bits in the value.
	// 11111 (binary) = 0x1F (hex), which with bitwise ANDing extracts the first 5 bits.
	// See: https://www.rapidtables.com/convert/number/binary-to-hex.html
	return identifier & 0x1F, nil
}

// GetBTreeNodeEntryFileOffset returns the file offset for this b-tree branch or leaf node.
// References "The b-tree entries".
func GetBTreeNodeEntryFileOffset(nodeEntryData []byte, formatType string, isBranchNode bool) (int, error) {
	var nodeOffsetOffset int
	var nodeOffsetBufferSize int

	if isBranchNode {
		switch formatType {
		case FormatTypeUnicode:
			nodeOffsetOffset = 16
			nodeOffsetBufferSize = 8
			break
		case FormatTypeUnicode4k:
			nodeOffsetOffset = 16
			nodeOffsetBufferSize = 8
			break
		case FormatTypeANSI:
			nodeOffsetOffset = 8
			nodeOffsetBufferSize = 4
			break
		default:
			return -1, errors.New("unsupported format type")
		}
	} else {
		switch formatType {
		case FormatTypeUnicode:
			nodeOffsetOffset = 8
			nodeOffsetBufferSize = 8
			break
		case FormatTypeUnicode4k:
			nodeOffsetOffset = 8
			nodeOffsetBufferSize = 8
			break
		case FormatTypeANSI:
			nodeOffsetOffset = 4
			nodeOffsetBufferSize = 4
			break
		default:
			return -1, errors.New("unsupported format type")
		}
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint64(nodeEntryData[nodeOffsetOffset:(nodeOffsetOffset + nodeOffsetBufferSize)])), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(nodeEntryData[nodeOffsetOffset:(nodeOffsetOffset + nodeOffsetBufferSize)])), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(nodeEntryData[nodeOffsetOffset:(nodeOffsetOffset + nodeOffsetBufferSize)])), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetBTreeNodeEntryDataIdentifier returns the node identifier of the data (in the block b-tree).
// References "The b-tree entries".
func GetBTreeNodeEntryDataIdentifier(nodeEntryData []byte, formatType string) (int, error) {
	var dataOffset int
	var dataBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		dataOffset = 8
		dataBufferSize = 8
		break
	case FormatTypeUnicode4k:
		dataOffset = 8
		dataBufferSize = 8
		break
	case FormatTypeANSI:
		dataOffset = 4
		dataBufferSize = 4
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint64(nodeEntryData[dataOffset:(dataOffset + dataBufferSize)])), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(nodeEntryData[dataOffset:(dataOffset + dataBufferSize)])), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(nodeEntryData[dataOffset:(dataOffset + dataBufferSize)])), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetBTreeNodeEntryLocalDescriptorsIdentifier returns the identifier to the local descriptors in the block b-tree.
func GetBTreeNodeEntryLocalDescriptorsIdentifier(nodeEntryData []byte, formatType string) (int, error) {
	var localDescriptorsOffset int
	var localDescriptorsBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		localDescriptorsOffset = 16
		localDescriptorsBufferSize = 8
		break
	case FormatTypeUnicode4k:
		localDescriptorsOffset = 16
		localDescriptorsBufferSize = 8
	case FormatTypeANSI:
		localDescriptorsOffset = 8
		localDescriptorsBufferSize = 4
	default:
		return -1, errors.New("unsupported format type")
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint64(nodeEntryData[localDescriptorsOffset : localDescriptorsOffset+localDescriptorsBufferSize])), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(nodeEntryData[localDescriptorsOffset : localDescriptorsOffset+localDescriptorsBufferSize])), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(nodeEntryData[localDescriptorsOffset : localDescriptorsOffset+localDescriptorsBufferSize])), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetBTreeNodeEntrySize returns the size of the data in the block b-tree leaf node entry.
// References "The b-tree entries".
func GetBTreeNodeEntrySize(nodeEntryData []byte, formatType string) (int, error) {
	var nodeSizeOffset int
	var nodeSizeBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		nodeSizeOffset = 16
		nodeSizeBufferSize = 2
		break
	case FormatTypeUnicode4k:
		nodeSizeOffset = 16
		nodeSizeBufferSize = 2
		break
	case FormatTypeANSI:
		nodeSizeOffset = 8
		nodeSizeBufferSize = 2
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	return int(binary.LittleEndian.Uint16(nodeEntryData[nodeSizeOffset:(nodeSizeOffset + nodeSizeBufferSize)])), nil
}

// Constants defining the b-tree types.
const (
	BTreeTypeNode  = 0
	BTreeTypeBlock = 1
)

// BTreeDegree defines the node and block b-tree degree.
var BTreeDegree = 6

// InitializeBTree walks the b-tree and finds the node with the given identifier.
func (pstFile *File) InitializeBTree(btreeType int, formatType string) error {
	if btreeType == BTreeTypeNode && pstFile.NodeBTree != nil || btreeType == BTreeTypeBlock && pstFile.BlockBTree != nil {
		return errors.New("b-tree is already initialized")
	}

	switch btreeType {
	case BTreeTypeNode:
		pstFile.NodeBTree = btree.New[BTreeNodeEntry](BTreeDegree)

		nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

		if err != nil {
			return err
		}

		return pstFile.WalkAndCreateBTree(pstFile.NodeBTree, nodeBTreeOffset, formatType)
	case BTreeTypeBlock:
		pstFile.BlockBTree = btree.New[BTreeNodeEntry](BTreeDegree)

		blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

		if err != nil {
			return err
		}

		return pstFile.WalkAndCreateBTree(pstFile.BlockBTree, blockBTreeOffset, formatType)
	default:
		return errors.New("invalid b-tree type")
	}
}

// InitializeBTrees initializes the node and block b-tree.
func (pstFile *File) InitializeBTrees(formatType string) error {
	err := pstFile.InitializeBTree(BTreeTypeNode, formatType)

	if err != nil {
		return err
	}

	err = pstFile.InitializeBTree(BTreeTypeBlock, formatType)

	if err != nil {
		return err
	}

	return nil
}

// WalkAndCreateBTree walks the b-tree and updates the given nodeEntryBTree.
func (pstFile *File) WalkAndCreateBTree(nodeEntryBTree *btree.BTree[BTreeNodeEntry], btreeOffset int, formatType string) error {
	nodeLevel, err := pstFile.GetBTreeNodeLevel(btreeOffset, formatType)

	if err != nil {
		return err
	}

	nodeEntries, err := pstFile.GetBTreeNodeEntries(btreeOffset, formatType, nodeLevel)

	if err != nil {
		return err
	}

	if nodeLevel > 0 {
		// Branch node entries.
		for i := 0; i < len(nodeEntries); i++ {
			nodeEntry := nodeEntries[i]

			nodeEntryBTree.ReplaceOrInsert(nodeEntry)

			err = pstFile.WalkAndCreateBTree(nodeEntryBTree, nodeEntry.FileOffset, formatType)

			if err != nil {
				return err
			}
		}
	} else {
		// Leaf node entries
		for i := 0; i < len(nodeEntries); i++ {
			nodeEntry := nodeEntries[i]

			nodeEntryBTree.ReplaceOrInsert(nodeEntry)
		}
	}

	return nil
}

// FindBTreeNode returns the node in the node or block b-tree with the given identifier.
func (pstFile *File) FindBTreeNode(btreeType int, identifier int) (BTreeNodeEntry, error) {
	switch btreeType {
	case BTreeTypeNode:
		btreeNodeEntry, found := pstFile.NodeBTree.Get(BTreeNodeEntry{Identifier: identifier})

		if !found {
			return BTreeNodeEntry{}, errors.New("failed to find b-tree node")
		}

		return btreeNodeEntry, nil
	case BTreeTypeBlock:
		btreeNodeEntry, found := pstFile.BlockBTree.Get(BTreeNodeEntry{Identifier: identifier})

		if !found {
			return BTreeNodeEntry{}, errors.New("failed to find b-tree node")
		}

		return btreeNodeEntry, nil
	default:
		return BTreeNodeEntry{}, errors.New("invalid b-tree type")
	}
}

// GetNodeBTreeNode returns the node with the given identifier in the node b-tree.
func (pstFile *File) GetNodeBTreeNode(identifier int) (BTreeNodeEntry, error) {
	return pstFile.FindBTreeNode(BTreeTypeNode, identifier)
}

// GetBlockBTreeNode returns the node with the given identifier in the block b-tree.
func (pstFile *File) GetBlockBTreeNode(identifier int) (BTreeNodeEntry, error) {
	// Clear the LSB, which is reserved, but sometimes set.
	return pstFile.FindBTreeNode(BTreeTypeBlock, identifier&0xfffffffe)
}

// GetDataBTreeNode searches the identifier in the node b-tree, then searches the data identifier in the block b-tree.
func (pstFile *File) GetDataBTreeNode(identifier int) (BTreeNodeEntry, error) {
	nodeBTreeNode, err := pstFile.GetNodeBTreeNode(identifier)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	return pstFile.GetBlockBTreeNode(nodeBTreeNode.DataIdentifier)
}
