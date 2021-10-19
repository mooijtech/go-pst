// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
)

// GetBlockSize returns the size of a block.
// References "Blocks".
func (pstFile *File) GetBlockSize(formatType string) (int, error) {
	switch formatType {
	case FormatTypeUnicode:
		return 8192, nil
	case FormatTypeUnicode4k:
		return 65536, nil
	case FormatTypeANSI:
		return 8192, nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetBlockTrailerSize returns the size of a block trailer.
// References "Blocks".
func (pstFile *File) GetBlockTrailerSize(formatType string) (int, error) {
	switch formatType {
	case FormatTypeUnicode:
		return 16, nil
	case FormatTypeUnicode4k:
		return 16, nil
	case FormatTypeANSI:
		return 12, nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// Constants defining the block types.
const (
	BlockTypeXBlock  = 1
	BlockTypeXXBlock = 2
)

// GetBlocks parses the XBlock and XXBlock from a Heap-on-Node.
// Used by the NodeInputStream (internal identifiers have blocks).
func (pstFile *File) GetBlocks(nodeEntryHeapOnNodeOffset int, formatType string) ([]BTreeNodeEntry, error) {
	blockSignature, err := pstFile.Read(1, nodeEntryHeapOnNodeOffset)

	if err != nil {
		return nil, err
	}

	if int(binary.LittleEndian.Uint16([]byte{blockSignature[0], 0})) != 1 {
		return nil, errors.New("invalid block signature (MUST be XBlock or XXBlock)")
	}

	// The number of block b-tree identifiers in this XBlock or XXBlock.
	entryCount, err := pstFile.Read(2, nodeEntryHeapOnNodeOffset+2)

	if err != nil {
		return nil, err
	}

	// The identifier size of a block identifier stored in the XBlock or XXBlock.
	var identifierSize int

	switch formatType {
	case FormatTypeUnicode:
		identifierSize = 8
		break
	case FormatTypeUnicode4k:
		identifierSize = 8
		break
	case FormatTypeANSI:
		identifierSize = 4
		break
	default:
		return nil, errors.New("unsupported format type")
	}

	blockLevel, err := pstFile.Read(1, nodeEntryHeapOnNodeOffset+1)

	var btreeNodeEntries []BTreeNodeEntry

	switch int(binary.LittleEndian.Uint16([]byte{blockLevel[0], 0})) {
	case BlockTypeXBlock:
		// XBlock
		offset := 8 // Start of the array of identifiers that reference data blocks.

		for i := 0; i < int(binary.LittleEndian.Uint16(entryCount)); i++ {
			blockIdentifier, err := pstFile.Read(identifierSize, nodeEntryHeapOnNodeOffset+offset)

			if err != nil {
				return nil, err
			}

			blockBTreeNode, err := pstFile.GetBlockBTreeNode(int(binary.LittleEndian.Uint32(blockIdentifier)), formatType)

			if err != nil {
				return nil, err
			}

			btreeNodeEntries = append(btreeNodeEntries, blockBTreeNode)

			offset += identifierSize
		}
		break
	case BlockTypeXXBlock:
		// XXBlock
		return nil, errors.New("XXBlock is not implemented yet, please open an issue on GitHub")
	default:
		return nil, errors.New("unsupported block type")
	}

	return btreeNodeEntries, nil
}

// GetBlocksTotalSize returns the size of the external data referenced by the XBlock or XXBlock.
func (pstFile *File) GetBlocksTotalSize(nodeEntryHeapOnNodeOffset int) (int, error) {
	totalDataSize, err := pstFile.Read(4, nodeEntryHeapOnNodeOffset+4)

	if err != nil {
		return -1, err
	}

	return int(binary.LittleEndian.Uint32(totalDataSize)), nil
}
