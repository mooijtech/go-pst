// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
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
	Data []byte
}

// NewBTreeNodeEntry is a constructor for creating node entries.
func NewBTreeNodeEntry(data []byte) BTreeNodeEntry {
	return BTreeNodeEntry{
		Data: data,
	}
}

// GetBTreeNodeEntries returns the entries in the b-tree node.
// References "The node and block b-tree".
func (pstFile *File) GetBTreeNodeEntries(btreeNodeOffset int, formatType string) ([]BTreeNodeEntry, error) {
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
		nodeEntry := nodeEntries[(i * nodeEntrySize) : (i*nodeEntrySize)+nodeEntrySize]

		entries[i] = NewBTreeNodeEntry(nodeEntry)
	}

	return entries, nil
}

// GetIdentifier returns the identifier of this b-tree node entry.
// References "The b-tree entries".
func (btreeNodeEntry *BTreeNodeEntry) GetIdentifier(formatType string) (int, error) {
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

	return int(binary.LittleEndian.Uint32(btreeNodeEntry.Data[:identifierBufferSize])), nil
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
)

// GetIdentifierType returns the b-tree node entry identifier type.
// References "The b-tree entries", "Identifier".
func (btreeNodeEntry *BTreeNodeEntry) GetIdentifierType(formatType string) (int, error) {
	nodeEntryIdentifier, err := btreeNodeEntry.GetIdentifier(formatType)

	if err != nil {
		return -1, err
	}

	// Bit masking:
	// Use bitwise ANDing in order to extract a subset of the bits in the value.
	// 11111 (binary) = 0x1F (hex), which with bitwise ANDing extracts the first 5 bits.
	// See: https://www.rapidtables.com/convert/number/binary-to-hex.html
	return nodeEntryIdentifier & 0x1F, nil
}

// GetFileOffset returns the file offset for this b-tree branch or leaf node.
// References "The b-tree entries".
func (btreeNodeEntry *BTreeNodeEntry) GetFileOffset(isBranchNode bool, formatType string) (int, error) {
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
		return int(binary.LittleEndian.Uint64(btreeNodeEntry.Data[nodeOffsetOffset:(nodeOffsetOffset + nodeOffsetBufferSize)])), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(btreeNodeEntry.Data[nodeOffsetOffset:(nodeOffsetOffset + nodeOffsetBufferSize)])), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(btreeNodeEntry.Data[nodeOffsetOffset:(nodeOffsetOffset + nodeOffsetBufferSize)])), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetDataIdentifier returns the node identifier of the data (in the node b-tree).
// References "The b-tree entries".
func (btreeNodeEntry *BTreeNodeEntry) GetDataIdentifier(formatType string) (int, error) {
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
		return int(binary.LittleEndian.Uint64(btreeNodeEntry.Data[dataOffset:(dataOffset + dataBufferSize)])), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(btreeNodeEntry.Data[dataOffset:(dataOffset + dataBufferSize)])), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(btreeNodeEntry.Data[dataOffset:(dataOffset + dataBufferSize)])), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetLocalDescriptorsIdentifier returns the identifier to the local descriptors in the block b-tree.
func (btreeNodeEntry *BTreeNodeEntry) GetLocalDescriptorsIdentifier(formatType string) (int, error) {
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
		return int(binary.LittleEndian.Uint64(btreeNodeEntry.Data[localDescriptorsOffset : localDescriptorsOffset+localDescriptorsBufferSize])), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint64(btreeNodeEntry.Data[localDescriptorsOffset : localDescriptorsOffset+localDescriptorsBufferSize])), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint32(btreeNodeEntry.Data[localDescriptorsOffset : localDescriptorsOffset+localDescriptorsBufferSize])), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetSize returns the size of the data in the block b-tree leaf node entry.
// References "The b-tree entries".
func (btreeNodeEntry *BTreeNodeEntry) GetSize(formatType string) (int, error) {
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

	return int(binary.LittleEndian.Uint16(btreeNodeEntry.Data[nodeSizeOffset:(nodeSizeOffset + nodeSizeBufferSize)])), nil
}

// Constants defining the b-tree types.
const (
	BTreeTypeNode = 0
	BTreeTypeBlock = 1
)

// Variables defining the node and block b-tree which need to be initialized.
var (
	NodeBTree []BTreeNodeEntry
	BlockBTree []BTreeNodeEntry
)

// InitializeBTree walks the b-tree and finds the node with the given identifier.
func (pstFile *File) InitializeBTree(btreeType int, formatType string) error {
	if len(NodeBTree) > 0 && btreeType == BTreeTypeNode || len(BlockBTree) > 0 && btreeType == BTreeTypeBlock {
		return errors.New("b-tree is already initialized")
	}

	switch btreeType {
	case BTreeTypeNode:
		nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

		if err != nil {
			return err
		}

		nodeBTreeNodes, err := pstFile.WalkBTree(nodeBTreeOffset, formatType)

		if err != nil {
			return err
		}

		NodeBTree = nodeBTreeNodes
		return nil
	case BTreeTypeBlock:
		blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

		if err != nil {
			return err
		}

		blockBTreeNodes, err := pstFile.WalkBTree(blockBTreeOffset, formatType)

		if err != nil {
			return err
		}

		BlockBTree = blockBTreeNodes
		return nil
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

// WalkBTree walks the b-tree and returns all nodes.
func (pstFile *File) WalkBTree(btreeOffset int, formatType string) ([]BTreeNodeEntry, error) {
	nodeEntries, err := pstFile.GetBTreeNodeEntries(btreeOffset, formatType)

	if err != nil {
		return []BTreeNodeEntry{}, err
	}

	nodeLevel, err := pstFile.GetBTreeNodeLevel(btreeOffset, formatType)

	if err != nil {
		return []BTreeNodeEntry{}, err
	}

	var btreeNodeEntries []BTreeNodeEntry

	if nodeLevel > 0 {
		// Branch node entries.
		for i := 0; i < len(nodeEntries); i++ {
			nodeEntry := nodeEntries[i]

			btreeNodeEntries = append(btreeNodeEntries, nodeEntry)

			// Recursively walk through the branch node entries.
			recursiveNodeOffset, err := nodeEntry.GetFileOffset(true, formatType)

			if err != nil {
				return []BTreeNodeEntry{}, err
			}

			recursiveNodeEntries, err := pstFile.WalkBTree(recursiveNodeOffset, formatType)

			if err != nil {
				log.Fatal("Got here")
				return []BTreeNodeEntry{}, err
			}

			for _, recursiveNodeEntry := range recursiveNodeEntries {
				btreeNodeEntries = append(btreeNodeEntries, recursiveNodeEntry)
			}
		}
	} else {
		// Leaf node entries
		for i := 0; i < len(nodeEntries); i++ {
			nodeEntry := nodeEntries[i]

			btreeNodeEntries = append(btreeNodeEntries, nodeEntry)
		}
	}

	return btreeNodeEntries, nil
}

// FindBTreeNode returns the node in the node or block b-tree with the given identifier.
// Note that we don't return the first identifier because there may be two nodes with the same identifier so prefer leaf nodes.
func (pstFile *File) FindBTreeNode(btreeType int, identifier int, formatType string) (BTreeNodeEntry, error) {
	switch btreeType {
	case BTreeTypeNode:
		var nodeBTreeNode BTreeNodeEntry

		for _, btreeNode := range NodeBTree {
			btreeNodeIdentifier, err := btreeNode.GetIdentifier(formatType)

			if err != nil {
				return BTreeNodeEntry{}, err
			}

			if btreeNodeIdentifier == identifier {
				nodeBTreeNode = btreeNode
			}
		}

		return nodeBTreeNode, nil
	case BTreeTypeBlock:
		var blockBTreeNode BTreeNodeEntry

		for _, btreeNode := range BlockBTree {
			btreeNodeIdentifier, err := btreeNode.GetIdentifier(formatType)

			if err != nil {
				return BTreeNodeEntry{}, err
			}

			if btreeNodeIdentifier == identifier {
				blockBTreeNode = btreeNode
			}
		}

		return blockBTreeNode, nil
	default:
		return BTreeNodeEntry{}, errors.New("invalid b-tree type")
	}
}

// GetNodeBTreeNode returns the node with the given identifier in the node b-tree.
func (pstFile *File) GetNodeBTreeNode(identifier int, formatType string) (BTreeNodeEntry, error) {
	nodeBTreeNode, err := pstFile.FindBTreeNode(BTreeTypeNode, identifier, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	return nodeBTreeNode, nil
}

// GetBlockBTreeNode returns the node with the given identifier in the block b-tree.
func (pstFile *File) GetBlockBTreeNode(identifier int, formatType string) (BTreeNodeEntry, error) {
	nodeBTreeNode, err := pstFile.FindBTreeNode(BTreeTypeBlock, identifier, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	return nodeBTreeNode, nil
}

// GetDataBTreeNode searches the identifier in the node b-tree, then searches the data identifier in the block b-tree.
func (pstFile *File) GetDataBTreeNode(identifier int, formatType string) (BTreeNodeEntry, error) {
	nodeBTreeNode, err := pstFile.GetNodeBTreeNode(identifier, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	dataIdentifier, err := nodeBTreeNode.GetDataIdentifier(formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	blockBTreeNode, err := pstFile.GetBlockBTreeNode(dataIdentifier, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	return blockBTreeNode, nil
}