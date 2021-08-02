// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
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
	return BTreeNodeEntry {
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
		nodeEntry := nodeEntries[(i * nodeEntrySize):(i * nodeEntrySize) + nodeEntrySize]

		entries[i] = NewBTreeNodeEntry(nodeEntry)
	}

	return entries, nil
}

// GetIdentifier returns the identifier of this b-tree node entry.
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

// GetFileOffset returns the file offset for this b-tree branch or leaf node.
// References "The b-tree entries".
func (btreeNodeEntry *BTreeNodeEntry) GetFileOffset(isBranchNode bool, formatType string) (int, error) {
	var nodeOffsetBufferSize int
	var nodeOffsetOffset int

	if isBranchNode {
		switch formatType {
		case FormatTypeUnicode:
			nodeOffsetBufferSize = 8
			nodeOffsetOffset = 16
			break
		case FormatTypeUnicode4k:
			nodeOffsetBufferSize = 8
			nodeOffsetOffset = 16
			break
		case FormatTypeANSI:
			nodeOffsetBufferSize = 4
			nodeOffsetOffset = 8
			break
		default:
			return -1, errors.New("unsupported format type")
		}
	} else {
		switch formatType {
		case FormatTypeUnicode:
			nodeOffsetBufferSize = 8
			nodeOffsetOffset = 16
			break
		case FormatTypeUnicode4k:
			nodeOffsetBufferSize = 8
			nodeOffsetOffset = 16
		case FormatTypeANSI:
			nodeOffsetBufferSize = 4
			nodeOffsetOffset = 4
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

// FindBTreeNode walks the b-tree and finds the node with the given identifier.
func (pstFile *File) FindBTreeNode(btreeNodeOffset int, identifier int, formatType string) (BTreeNodeEntry, error) {
	nodeEntries, err := pstFile.GetBTreeNodeEntries(btreeNodeOffset, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	nodeLevel, err := pstFile.GetBTreeNodeLevel(btreeNodeOffset, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	if nodeLevel > 0 {
		// Branch node entries.
		for i := 0; i < len(nodeEntries); i++ {
			nodeEntry := nodeEntries[i]

			nodeEntryIdentifier, err := nodeEntry.GetIdentifier(formatType)

			if err != nil {
				return BTreeNodeEntry{}, err
			}

			if nodeEntryIdentifier == identifier {
				return nodeEntry, nil
			}

			// Recursively walk through the branch node entries.
			recursiveNodeOffset, err := nodeEntry.GetFileOffset(true, formatType)

			if err != nil {
				return BTreeNodeEntry{}, err
			}

			recursiveNodeEntry, err := pstFile.FindBTreeNode(recursiveNodeOffset, identifier, formatType)

			if err != nil {
				return BTreeNodeEntry{}, err
			}

			recursiveNodeEntryIdentifier, err := recursiveNodeEntry.GetIdentifier(formatType)
			
			if err != nil {
				return BTreeNodeEntry{}, err
			}

			if recursiveNodeEntryIdentifier == identifier {
				return recursiveNodeEntry, nil
			}

		}
	} else {
		// Leaf node entries
		for i := 0; i < len(nodeEntries); i++ {
			nodeEntry := nodeEntries[i]

			nodeEntryIdentifier, err := nodeEntry.GetIdentifier(formatType)

			if err != nil {
				return BTreeNodeEntry{}, err
			}

			if nodeEntryIdentifier == identifier {
				return nodeEntry, nil
			}
		}
	}

	return BTreeNodeEntry{}, errors.New("failed to find b-tree node")
}