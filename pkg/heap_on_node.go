// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
)

// GetHeapOnNode decrypts the Heap-on-Node using compressible encryption.
// References "Compressible encryption".
func (pstFile *File) GetHeapOnNode(btreeNodeEntry BTreeNodeEntry, formatType string) (BTreeNodeEntry, error) {
	nodeEntryHeapOnNodeOffset, err := btreeNodeEntry.GetFileOffset(false, formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	nodeEntryHeapOnNodeSize, err := btreeNodeEntry.GetSize(formatType)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	nodeEntryHeapOnNode, err := pstFile.Read(nodeEntryHeapOnNodeSize, nodeEntryHeapOnNodeOffset)

	if err != nil {
		return BTreeNodeEntry{}, err
	}

	compressibleEncryption := []int {
		0x47, 0xf1, 0xb4, 0xe6, 0x0b, 0x6a, 0x72, 0x48, 0x85, 0x4e, 0x9e, 0xeb, 0xe2, 0xf8, 0x94, 0x53, 0xe0,
		0xbb, 0xa0, 0x02, 0xe8, 0x5a, 0x09, 0xab, 0xdb, 0xe3, 0xba, 0xc6, 0x7c, 0xc3, 0x10, 0xdd, 0x39, 0x05,
		0x96, 0x30, 0xf5, 0x37, 0x60, 0x82, 0x8c, 0xc9, 0x13, 0x4a, 0x6b, 0x1d, 0xf3, 0xfb, 0x8f, 0x26, 0x97,
		0xca, 0x91, 0x17, 0x01, 0xc4, 0x32, 0x2d, 0x6e, 0x31, 0x95, 0xff, 0xd9, 0x23, 0xd1, 0x00, 0x5e, 0x79,
		0xdc, 0x44, 0x3b, 0x1a, 0x28, 0xc5, 0x61, 0x57, 0x20, 0x90, 0x3d, 0x83, 0xb9, 0x43, 0xbe, 0x67, 0xd2,
		0x46, 0x42, 0x76, 0xc0, 0x6d, 0x5b, 0x7e, 0xb2, 0x0f, 0x16, 0x29, 0x3c, 0xa9, 0x03, 0x54, 0x0d, 0xda,
		0x5d, 0xdf, 0xf6, 0xb7, 0xc7, 0x62, 0xcd, 0x8d, 0x06, 0xd3, 0x69, 0x5c, 0x86, 0xd6, 0x14, 0xf7, 0xa5,
		0x66, 0x75, 0xac, 0xb1, 0xe9, 0x45, 0x21, 0x70, 0x0c, 0x87, 0x9f, 0x74, 0xa4, 0x22, 0x4c, 0x6f, 0xbf,
		0x1f, 0x56, 0xaa, 0x2e, 0xb3, 0x78, 0x33, 0x50, 0xb0, 0xa3, 0x92, 0xbc, 0xcf, 0x19, 0x1c, 0xa7, 0x63,
		0xcb, 0x1e, 0x4d, 0x3e, 0x4b, 0x1b, 0x9b, 0x4f, 0xe7, 0xf0, 0xee, 0xad, 0x3a, 0xb5, 0x59, 0x04, 0xea,
		0x40, 0x55, 0x25, 0x51, 0xe5, 0x7a, 0x89, 0x38, 0x68, 0x52, 0x7b, 0xfc, 0x27, 0xae, 0xd7, 0xbd, 0xfa,
		0x07, 0xf4, 0xcc, 0x8e, 0x5f, 0xef, 0x35, 0x9c, 0x84, 0x2b, 0x15, 0xd5, 0x77, 0x34, 0x49, 0xb6, 0x12,
		0x0a, 0x7f, 0x71, 0x88, 0xfd, 0x9d, 0x18, 0x41, 0x7d, 0x93, 0xd8, 0x58, 0x2c, 0xce, 0xfe, 0x24, 0xaf,
		0xde, 0xb8, 0x36, 0xc8, 0xa1, 0x80, 0xa6, 0x99, 0x98, 0xa8, 0x2f, 0x0e, 0x81, 0x65, 0x73, 0xe4, 0xc2,
		0xa2, 0x8a, 0xd4, 0xe1, 0x11, 0xd0, 0x08, 0x8b, 0x2a, 0xf2, 0xed, 0x9a, 0x64, 0x3f, 0xc1, 0x6c, 0xf9, 0xec,
	}

	for i := 0; i < len(nodeEntryHeapOnNode); i++ {
		temp := nodeEntryHeapOnNode[i] & 0xff
		nodeEntryHeapOnNode[i] = byte(compressibleEncryption[temp])
	}

	return NewBTreeNodeEntry(nodeEntryHeapOnNode), nil
}

// IsValidHeapOnNodeSignature returns true if the signature of the block matches 0xEC (236).
// References "Heap-on-Node header".
func (btreeNodeEntry *BTreeNodeEntry) IsValidHeapOnNodeSignature() bool {
	return binary.LittleEndian.Uint16([]byte{btreeNodeEntry.Data[2], 0}) == 236
}

// GetHeapOnNodeTableType returns the table type.
// References "Heap-on-Node header", "Table types".
func (btreeNodeEntry *BTreeNodeEntry) GetHeapOnNodeTableType() int {
	return int(binary.LittleEndian.Uint16([]byte{btreeNodeEntry.Data[3], 0}))
}

// GetHeapOnNodeHIDUserRoot returns the HID user root.
// References "Heap-on-Node header".
func (btreeNodeEntry *BTreeNodeEntry) GetHeapOnNodeHIDUserRoot() int {
	return int(binary.LittleEndian.Uint16(btreeNodeEntry.Data[4:8]))
}

// GetHeapOnNodeHIDUserRootType MUST be set to 0 (IdentifierTypeHID) to indicate a valid HID.
func (btreeNodeEntry *BTreeNodeEntry) GetHeapOnNodeHIDUserRootType() int {
	return btreeNodeEntry.GetHeapOnNodeHIDUserRoot() & 0x1F
}

func (btreeNodeEntry *BTreeNodeEntry) GetHeapOnNodeHIDUserRootIndex() int {
	return btreeNodeEntry.GetHeapOnNodeHIDUserRoot() >> 5
}

// GetHeapOnNodeHIDUserRootBlockIndex returns in which block (the index) the Heap-on-Node item resides.
func (btreeNodeEntry *BTreeNodeEntry) GetHeapOnNodeHIDUserRootBlockIndex() int {
	return btreeNodeEntry.GetHeapOnNodeHIDUserRoot() >> 16
}

func (btreeNodeEntry *BTreeNodeEntry) GetHeapOnNodePageMap() int {
	return int(binary.LittleEndian.Uint16(btreeNodeEntry.Data[:2]))
}

func (btreeNodeEntry *BTreeNodeEntry) GetHeapOnNodePageMapAllocationCount() int {
	pageMapOffset := btreeNodeEntry.GetHeapOnNodePageMap()

	return int(binary.LittleEndian.Uint16(btreeNodeEntry.Data[pageMapOffset:pageMapOffset + 2]))
}

// HeapOnNodeBlock represents a Heap-on-Node block.
// References "Heap-on-Node".
type HeapOnNodeBlock struct {
	BlockIndex int
	Buffer []byte
}

// NewHeapOnNodeBlock is a constructor for creating Heap-on-Node blocks.
func NewHeapOnNodeBlock(blockIndex int, buffer []byte) HeapOnNodeBlock {
	return HeapOnNodeBlock {
		BlockIndex: blockIndex,
		Buffer: buffer,
	}
}

// GetHeapOnNodeBlocks reads the heap on node blocks.
// References "Heap-on-Node"
func (pstFile *File) GetHeapOnNodeBlocks(btreeNodeEntryHeapOnNode BTreeNodeEntry, formatType string) ([]HeapOnNodeBlock, error) {
	if !btreeNodeEntryHeapOnNode.IsValidHeapOnNodeSignature() {
		// The Heap-on-Node was probably not decrypted via GetHeapOnNode.
		return nil, errors.New("invalid heap-on-node signature")
	}

	var nodeEntryBlocks [][]byte

	nodeEntryIdentifierType, err := btreeNodeEntryHeapOnNode.GetIdentifierType(formatType)

	if err != nil {
		return nil, err
	}

	if nodeEntryIdentifierType == IdentifierTypeInternal {
		// TODO - XBlock and XXBlock
		return nil, errors.New("not implemented yet")
	} else {
		// TODO - Key for cyclic algorithm is the low 32 bits of the node identifier.
		nodeEntryBlocks = append(nodeEntryBlocks, btreeNodeEntryHeapOnNode.Data)
	}

	var blocks []HeapOnNodeBlock

	for i, nodeEntryBlock := range nodeEntryBlocks {
		if i == 0 {
			// The first block contains the Heap-on-Node header.
			blocks = append(blocks, NewHeapOnNodeBlock(i, nodeEntryBlock))
		} else if i == 8 || i >= 138 && (i - 8) / 128 == 0 {
			// Blocks 8, 136, then every 128th contains the Heap-on-Node bitmap header
			blocks = append(blocks, NewHeapOnNodeBlock(i, nodeEntryBlock))
		} else {
			// All other blocks contain the Heap-on-Node page header
			blocks = append(blocks, NewHeapOnNodeBlock(i, nodeEntryBlock))
		}
	}

	return blocks, nil
}
