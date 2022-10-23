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

	"github.com/pkg/errors"
)

// GetBlockSize returns the size of a block.
// References "Blocks".
func (file *File) GetBlockSize() (int, error) {
	switch file.FormatType {
	case FormatTypeUnicode:
		return 8192, nil
	case FormatTypeUnicode4k:
		return 65536, nil
	case FormatTypeANSI:
		return 8192, nil
	default:
		return 0, errors.WithStack(ErrFormatTypeUnsupported)
	}
}

// GetBlockTrailerSize returns the size of a block trailer.
// References "Blocks".
func (file *File) GetBlockTrailerSize() (int, error) {
	switch file.FormatType {
	case FormatTypeUnicode:
		return 16, nil
	case FormatTypeUnicode4k:
		return 16, nil
	case FormatTypeANSI:
		return 12, nil
	default:
		return 0, errors.WithStack(ErrFormatTypeUnsupported)
	}
}

// BlockType represents a XBlock or XXBlock.
type BlockType uint8

// Constants defining the block types.
const (
	BlockTypeXBlock  BlockType = 1
	BlockTypeXXBlock BlockType = 2
)

// GetBlocks returns all blocks (XBlock/XXBlock) from a Heap-on-Node.
// Internal identifiers have blocks.
//
// References:
// - https://github.com/mooijtech/go-pst/tree/master/docs#xblock
// - https://github.com/mooijtech/go-pst/tree/master/docs#xxblock
func (file *File) GetBlocks(btreeNodeHeapOnNodeOffset int64) ([]BTreeNode, error) {
	blockSignature := make([]byte, 1)

	if _, err := file.ReadAt(blockSignature, btreeNodeHeapOnNodeOffset); err != nil {
		return nil, errors.WithStack(err)
	} else if blockSignature[0] != 1 {
		return nil, errors.WithStack(ErrBlockSignatureInvalid)
	}

	// The number of block b-tree identifiers in this XBlock or XXBlock.
	entryCount := make([]byte, 2)

	if _, err := file.ReadAt(entryCount, btreeNodeHeapOnNodeOffset+2); err != nil {
		return nil, errors.WithStack(err)
	}

	blockLevel := make([]byte, 1)

	if _, err := file.ReadAt(blockLevel, btreeNodeHeapOnNodeOffset+1); err != nil {
		return nil, errors.WithStack(err)
	}

	var blocks []BTreeNode

	switch BlockType(blockLevel[0]) {
	case BlockTypeXBlock:
		// XBlock
		blockIdentifierOffset := int64(8)

		for i := 0; i < int(binary.LittleEndian.Uint16(entryCount)); i++ {
			blockIdentifier := make([]byte, GetIdentifierSize(file.FormatType))

			if _, err := file.ReadAt(blockIdentifier, btreeNodeHeapOnNodeOffset+blockIdentifierOffset); err != nil {
				return nil, errors.WithStack(err)
			}

			blockBTreeNode, err := file.GetBlockBTreeNode(GetIdentifierFromBytes(blockIdentifier, file.FormatType))

			if err != nil {
				return nil, errors.WithStack(err)
			}

			blocks = append(blocks, blockBTreeNode)
			blockIdentifierOffset += int64(GetIdentifierSize(file.FormatType))
		}
	case BlockTypeXXBlock:
		// XXBlock
		blockIdentifierOffset := int64(8)

		for i := 0; i < int(binary.LittleEndian.Uint16(entryCount)); i++ {
			blockIdentifier := make([]byte, GetIdentifierSize(file.FormatType))

			if _, err := file.ReadAt(blockIdentifier, btreeNodeHeapOnNodeOffset+blockIdentifierOffset); err != nil {
				return nil, errors.WithStack(err)
			}

			blockBTreeNode, err := file.GetBlockBTreeNode(GetIdentifierFromBytes(blockIdentifier, file.FormatType))

			if err != nil {
				return nil, errors.WithStack(err)
			}

			blockBTreeNodeBlocks, err := file.GetBlocks(blockBTreeNode.FileOffset)

			if err != nil {
				return nil, errors.WithStack(err)
			}

			blocks = append(blocks, blockBTreeNodeBlocks...)
			blockIdentifierOffset += int64(GetIdentifierSize(file.FormatType))
		}
	default:
		return nil, errors.WithStack(ErrBlockTypeInvalid)
	}

	return blocks, nil
}

// GetBlocksTotalSize returns the size of the external data referenced by the XBlock or XXBlock.
func (file *File) GetBlocksTotalSize(nodeEntryHeapOnNodeOffset int64) (uint32, error) {
	blocksTotalSize := make([]byte, 4)

	if _, err := file.ReadAt(blocksTotalSize, nodeEntryHeapOnNodeOffset+4); err != nil {
		return 0, errors.WithStack(err)
	}

	return binary.LittleEndian.Uint32(blocksTotalSize), nil
}
