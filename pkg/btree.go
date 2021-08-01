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