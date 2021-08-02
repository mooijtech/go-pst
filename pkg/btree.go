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
	var entryCountOffset int
	var entryCountBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		entryCountOffset = btreeNodeOffset + 488
		entryCountBufferSize = 1
		break
	case FormatTypeUnicode4k:
		entryCountOffset = btreeNodeOffset + 4056
		entryCountBufferSize = 2
		break
	case FormatTypeANSI:
		entryCountOffset = btreeNodeOffset + 496
		entryCountBufferSize = 1
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	entryCount, err := pstFile.Read(entryCountBufferSize, entryCountOffset)

	if err != nil {
		return -1, err
	}

	switch formatType {
	case FormatTypeUnicode:
		return int(binary.LittleEndian.Uint16([]byte{entryCount[0], 0})), nil
	case FormatTypeUnicode4k:
		return int(binary.LittleEndian.Uint16(entryCount)), nil
	case FormatTypeANSI:
		return int(binary.LittleEndian.Uint16([]byte{entryCount[0], 0})), nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetBTreeNodeEntrySize returns the size of an entry in the b-tree.
// References "The node and block b-tree".
func (pstFile *File) GetBTreeNodeEntrySize(btreeNodeOffset int, formatType string) (int, error) {
	var entrySizeOffset int

	switch formatType {
	case FormatTypeUnicode:
		entrySizeOffset = btreeNodeOffset + 490
		break
	case FormatTypeUnicode4k:
		entrySizeOffset = btreeNodeOffset + 4060
		break
	case FormatTypeANSI:
		entrySizeOffset = btreeNodeOffset + 498
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	entrySize, err := pstFile.Read(1, entrySizeOffset)

	if err != nil {
		return -1, err
	}

	return int(binary.LittleEndian.Uint16([]byte{entrySize[0], 0})), nil
}

// GetBTreeNodeLevel returns the level of the b-tree node.
// References "The node and block b-tree".
func (pstFile *File) GetBTreeNodeLevel(btreeNodeOffset int, formatType string) (int, error) {
	var levelOffset int

	switch formatType {
	case FormatTypeUnicode:
		levelOffset = 491
		break
	case FormatTypeUnicode4k:
		levelOffset = 4061
		break
	case FormatTypeANSI:
		levelOffset = 499
		break
	default:
		return -1, errors.New("unsupported format type")
	}

	level, err := pstFile.Read(1, levelOffset)

	if err != nil {
		return -1, err
	}

	return int(binary.LittleEndian.Uint16([]byte{level[0], 0})), nil
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
	var entriesBufferSize int

	switch formatType {
	case FormatTypeUnicode:
		entriesBufferSize = 488
		break
	case FormatTypeUnicode4k:
		entriesBufferSize = 4056
		break
	case FormatTypeANSI:
		entriesBufferSize = 496
		break
	default:
		return nil, errors.New("unsupported format type")
	}

	entries, err := pstFile.Read(entriesBufferSize, btreeNodeOffset)

	if err != nil {
		return nil, err
	}

	entryCount, err := pstFile.GetBTreeNodeEntryCount(btreeNodeOffset, formatType)

	if err != nil {
		return nil, err
	}

	entrySize, err := pstFile.GetBTreeNodeEntrySize(btreeNodeOffset, formatType)

	if err != nil {
		return nil, err
	}

	nodeEntries := make([]BTreeNodeEntry, entryCount)

	for i := 0; i < entryCount; i++ {
		entry := entries[(i * entrySize):(i * entrySize) + entrySize]

		nodeEntries[i] = NewBTreeNodeEntry(entry)
	}

	return nodeEntries, nil
}