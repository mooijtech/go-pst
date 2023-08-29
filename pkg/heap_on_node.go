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
	"encoding/binary"
	"io"

	"github.com/rotisserie/eris"
)

// HeapOnNode represents a Heap-on-Node.
type HeapOnNode struct {
	Reader *HeapOnNodeReader
}

// SignatureType represents a signature type (for headers).
type SignatureType uint8

// Constants defining the signature types.
// TODO - Use in read checks, change type to byte for checks?
const (
	SignatureTypeTableContext    SignatureType = 124
	SignatureTypeBTreeOnHeap     SignatureType = 181
	SignatureTypePropertyContext SignatureType = 188
)

// NewHeapOnNode creates a new HeapOnNode.
func NewHeapOnNode(reader *HeapOnNodeReader) *HeapOnNode {
	return &HeapOnNode{Reader: reader}
}

// IsValidSignature returns true if the signature of the block matches 0xEC (236).
// References "Heap-on-Node header".
func (heapOnNode *HeapOnNode) IsValidSignature() (bool, error) {
	signature := make([]byte, 1)

	if _, err := heapOnNode.Reader.ReadAt(signature, 2); err != nil {
		return false, eris.Wrap(err, "failed to read signature")
	}

	return signature[0] == 236, nil
}

// GetTableType returns the table type.
// References "Heap-on-Node header", "Table types".
func (heapOnNode *HeapOnNode) GetTableType() (uint8, error) {
	tableType := make([]byte, 1)

	if _, err := heapOnNode.Reader.ReadAt(tableType, 3); err != nil {
		return 0, eris.Wrap(err, "failed to read table type")
	}

	return tableType[0], nil
}

// GetHIDUserRoot returns the HID user root.
// References "Heap-on-Node header".
func (heapOnNode *HeapOnNode) GetHIDUserRoot() (Identifier, error) {
	hidUserRoot := make([]byte, 4)

	if _, err := heapOnNode.Reader.ReadAt(hidUserRoot, 4); err != nil {
		return 0, eris.Wrap(err, "failed to read HID user root")
	}

	return Identifier(binary.LittleEndian.Uint32(hidUserRoot)), nil
}

// GetHeapOnNode returns the Heap-on-Node of the b-tree node.
func (file *File) GetHeapOnNode(btreeNode BTreeNode) (*HeapOnNode, error) {
	// Internal identifiers have blocks (XBlock or XXBlock).
	// This is a list of block identifiers that point to block b-tree entries (where the data is).
	isInternal := btreeNode.Identifier&0x02 != 0

	if isInternal {
		blocks, err := file.GetBlocks(btreeNode.FileOffset)

		if err != nil {
			return nil, eris.Wrap(err, "failed to get blocks")
		}

		blocksTotalSize, err := file.GetBlocksTotalSize(btreeNode.FileOffset)

		if err != nil {
			return nil, eris.Wrap(err, "failed to get total blocks size")
		}

		blockReaders := make([]io.SectionReader, len(blocks))
		blockReaderTotalSize := 0

		for i, block := range blocks {
			blockReaderTotalSize += int(block.Size)
			blockReaders[i] = *NewBTreeNodeReader(block, file.Reader)
		}

		if blocksTotalSize != uint32(blockReaderTotalSize) {
			return nil, ErrTotalBlocksSizeMismatch
		}

		return &HeapOnNode{Reader: NewHeapOnNodeReader(file.EncryptionType, blockReaders...)}, nil
	}

	return &HeapOnNode{Reader: NewHeapOnNodeReader(file.EncryptionType, *io.NewSectionReader(file.Reader, btreeNode.FileOffset, int64(btreeNode.Size)))}, nil
}

// GetHeapOnNodeReaderFromHNID returns the Heap-on-Node reader from the specified HNID (heap or node identifier).
// Note this doesn't keep track of all the passed HeapOnNodeReader blocks.
func (file *File) GetHeapOnNodeReaderFromHNID(hnid Identifier, heapOnNodeReader HeapOnNodeReader, localDescriptors ...LocalDescriptor) (*HeapOnNodeReader, error) {
	if len(localDescriptors) > 0 {
		localDescriptor, err := FindLocalDescriptor(hnid, localDescriptors)

		if err == nil {
			localDescriptorHeapOnNode, err := file.GetHeapOnNodeFromLocalDescriptor(localDescriptor)

			if err != nil {
				return nil, eris.Wrap(err, "failed to get Heap-on-Node from local descriptor")
			}

			return localDescriptorHeapOnNode.Reader, nil
		}
	}

	return file.GetHeapOnNodeReaderFromHID(hnid, heapOnNodeReader)
}

// GetHeapOnNodeReaderFromHID returns the Heap-on-Node reader from the heap ID.
func (file *File) GetHeapOnNodeReaderFromHID(hid Identifier, heapOnNodeReader HeapOnNodeReader) (*HeapOnNodeReader, error) {
	if hid.GetType() != IdentifierTypeHID {
		// The data is in the local descriptors (when the HNID matches a local descriptor identifier).
		// This gives us a data identifier that points to a node in the block b-tree (another Heap-on-Node).
		// Maybe there were no local descriptors specified in GetHeapOnNodeReaderFromHNID?
		return nil, ErrHeapOnNodeExternalNode
	}

	blockIndex := int(hid) >> 16
	blockOffset := int64(0)

	if blockIndex > 0 {
		if blockIndex > len(heapOnNodeReader.Blocks) {
			return nil, ErrBlockIndexNotFound
		}

		blockOffset = heapOnNodeReader.BlockOffsets[blockIndex]
	}

	pageMapOffset := make([]byte, 2)

	if _, err := heapOnNodeReader.ReadAt(pageMapOffset, blockOffset); err != nil {
		return nil, eris.Wrap(err, "failed to read page map offset")
	}

	allocationIndex := int64((hid & 0xFFFF) >> 5)
	allocationOffset := (blockOffset + int64(binary.LittleEndian.Uint16(pageMapOffset))) + (2 * allocationIndex) + 2

	startOffset := make([]byte, 2)

	if _, err := heapOnNodeReader.ReadAt(startOffset, allocationOffset); err != nil {
		return nil, eris.Wrap(err, "failed to read start offset")
	}

	endOffset := make([]byte, 2)

	if _, err := heapOnNodeReader.ReadAt(endOffset, allocationOffset+2); err != nil {
		return nil, eris.Wrap(err, "failed to read end offset")
	}

	// Note that the block start offset is only for this singular block, not across all blocks.
	blockStartOffset := int64(binary.LittleEndian.Uint16(startOffset))
	blockEndOffset := int64(binary.LittleEndian.Uint16(endOffset))

	return NewHeapOnNodeReader(file.EncryptionType, *io.NewSectionReader(&heapOnNodeReader.Blocks[blockIndex], blockStartOffset, blockEndOffset-blockStartOffset)), nil
}

// GetHeapOnNodeFromLocalDescriptor creates a Heap-on-Node from the local descriptor.
func (file *File) GetHeapOnNodeFromLocalDescriptor(localDescriptor LocalDescriptor) (*HeapOnNode, error) {
	localDescriptorDataNode, err := file.GetBlockBTreeNode(localDescriptor.DataIdentifier)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get block b-tree node")
	}

	return file.GetHeapOnNode(localDescriptorDataNode)
}
