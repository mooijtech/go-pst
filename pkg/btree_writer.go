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
	"bytes"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
	"io"
)

// BTreeWriter represents a writer for B-Trees.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#btrees
type BTreeWriter struct {
	// Writer represents the io.Writer which is used during writing.
	Writer io.Writer
	// WriteGroup represents writer running in Goroutines.
	WriteGroup *errgroup.Group
	// FormatType represents the FormatType to use during writing.
	FormatType FormatType
	// BTreeType represents the type of b-tree to write (node or block).
	BTreeType BTreeType
	// BTreeWriteChannel represents a Go channel for writing B-Tree nodes.
	BTreeWriteChannel chan BTreeNode
	// BTreeWriteCallback represents the callback which is called once a B-Tree node is written.
	BTreeWriteCallback chan WriteCallbackResponse
	// Identifier represents the identifier of the written B-Tree node so that it can be found in the B-Tree.
	Identifier Identifier
}

// NewBTreeWriter creates a new BTreeWriter.
func NewBTreeWriter(writer io.Writer, writeGroup *errgroup.Group, btreeWriteCallback chan WriteCallbackResponse, formatType FormatType, btreeType BTreeType) *BTreeWriter {
	btreeWriteChannel := make(chan BTreeNode)

	btreeWriter := &BTreeWriter{
		Writer:             writer,
		WriteGroup:         writeGroup,
		FormatType:         formatType,
		BTreeType:          btreeType,
		BTreeWriteChannel:  btreeWriteChannel,
		BTreeWriteCallback: btreeWriteCallback,
	}

	// Start the Go channel for writing B-Tree nodes.
	go btreeWriter.StartBTreeNodeWriteChannel(btreeWriteChannel)

	return btreeWriter
}

// UpdateIdentifier is called after WriteTo so that this B-Tree node can be found in the B-Tree.
func (btreeWriter *BTreeWriter) UpdateIdentifier() error {
	identifier, err := NewIdentifier(btreeWriter.FormatType)

	if err != nil {
		return eris.Wrap(err, "failed to create identifier")
	}

	btreeWriter.Identifier = identifier

	return nil
}

// AddBTreeNodes adds the B-Trees nodes to the write queue.
// Processed by StartBTreeNodeWriteChannel.
func (btreeWriter *BTreeWriter) AddBTreeNodes(btreeNodes ...BTreeNode) {
	for _, btreeNode := range btreeNodes {
		btreeWriter.BTreeWriteChannel <- btreeNode
	}
}

// StartBTreeNodeWriteChannel writes B-Tree nodes using Goroutines.
func (btreeWriter *BTreeWriter) StartBTreeNodeWriteChannel(btreeNodeWriteChannel chan BTreeNode) {
	for btreeNode := range btreeNodeWriteChannel {
		btreeWriter.WriteGroup.Go(func() error {
			// Write the B-Tree node.
			written, err := btreeNode.WriteTo(btreeWriter.Writer)

			if err != nil {
				return eris.Wrap(err, "failed to write B-Tree node")
			}

			// Callback, used to calculate the total size.
			// "Share memory by communicating; don't communicate by sharing memory."
			btreeWriter.BTreeWriteCallback <- NewWriteCallbackResponse(written)

			return nil
		})
	}
}

// WriteTo writes the B-Tree.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#btpage
func (btreeWriter *BTreeWriter) WriteTo(writer io.Writer, btreeNodes [][]byte, btreeNodeLevel int) (int64, error) {
	btree := bytes.NewBuffer(make([]byte, 512)) // Same for Unicode and ANSI.

	// This section contains entries of the BTree array.
	// The node if either a branch or leaf node depending on the node level.
	for _, btreeNode := range btreeNodes {
		if _, err := btree.Write(btreeNode); err != nil {
			return 0, eris.Wrap(err, "failed to write B-Tree node")
		}
	}

	// The number of BTree entries stored in the page data.
	// The entries depend on the value of this field.
	btree.WriteByte(byte(len(btreeNodes)))

	// The maximum number of entries that can fit inside the page data.
	btree.Write(make([]byte, 1)) // TODO

	// The size of each BTree entry, in bytes.
	if btreeWriter.BTreeType == BTreeTypeNode && btreeNodeLevel == 0 {
		switch btreeWriter.FormatType {
		case FormatTypeUnicode:
			btree.Write([]byte{32})
		case FormatTypeANSI:
			btree.Write([]byte{16})
		default:
			return 0, ErrFormatTypeUnsupported
		}
	} else {
		switch btreeWriter.FormatType {
		case FormatTypeUnicode:
			btree.Write([]byte{24})
		case FormatTypeANSI:
			btree.Write([]byte{12})
		default:
			return 0, ErrFormatTypeUnsupported
		}
	}

	// The depth level of this page.
	// Leaf nodes have a level of 0, while branch nodes have a level greater than 0.
	// This value determines the type of B-Tree nodes (branch or leaf).
	btree.WriteByte(byte(btreeNodeLevel))

	// Padding that should be set to zero.
	// Note that there is no padding in the ANSI version of the structure.
	if btreeWriter.FormatType == FormatTypeUnicode {
		btree.Write(make([]byte, 4))
	}

	// Page trailer.
	// A PageTrailer structure with specific subfield values.
	// The "ptype" subfield of "pageTrailer" should be set to "ptypeBBT" for a Block BTree page or "ptypeNBT" for a Node BTree page.
	// The other subfields of "pageTrailer" should be set as specified in the documentation.
	if _, err := btreeWriter.WritePageTrailer(btree, btreeWriter.BTreeType); err != nil {
		return 0, eris.Wrap(err, "failed to write page trailer")
	}

	return btree.WriteTo(writer)
}

// WritePageTrailer writes the page tailer of the b-tree.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#pagetrailer
func (btreeWriter *BTreeWriter) WritePageTrailer(writer io.Writer, btreeType BTreeType, btreeFileOffset int64, btreeNodeIdentifier Identifier) (int64, error) {
	pageTrailerBuffer := bytes.NewBuffer(make([]byte, 0))

	// This value indicates if we are writing a node or block B-Tree.
	pageTrailerBuffer.WriteByte(byte(btreeWriter.BTreeType))

	// MUST be set to the same value as previous BTreeType.
	pageTrailerBuffer.WriteByte(byte(btreeWriter.BTreeType))

	// Page signature.
	if _, err := btreeWriter.WriteBlockSignature(pageTrailerBuffer, btreeFileOffset, btreeNodeIdentifier); err != nil {
		return 0, eris.Wrap(err, "failed to write page signature")
	}

	// 32-bit CRC of the page data, excluding the page trailer.
	// See section 5.3 for the CRC algorithm.
	// TODO - Check if Microsoft uses a custom CRC.
	// Note the locations of the dwCRC and bid are differs between the Unicode and ANSI version of this structure.
	// TODO - pageTrailerBuffer.Write()

	// Write the identifier.
	pageTrailerBuffer.Write(btreeNodeIdentifier.Bytes(btreeWriter.FormatType))

	// The bidIndex for other page types are allocated from the special bidNextP counter in the HEADER structure.
	// TODO - pageTrailerBuffer.Write()

	return 0, nil
}

// WriteBlockSignature writes the block signature.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#block-signature
func (btreeWriter *BTreeWriter) WriteBlockSignature(writer io.Writer, fileOffset int64, identifier Identifier) (int, error) {
	// A WORD is a 16-bit unsigned integer.
	// A DWORD is a 32-bit unsigned integer.
	// The signature is calculated by first obtaining the DWORD XOR result between the absolute file offset of the block and its identifier.
	fileOffset ^= int64(identifier)

	// The WORD signature is then obtained by obtaining the XOR result between the higher and lower 16 bits of the DWORD obtained previously.
	return writer.Write(GetUint16(uint16(fileOffset>>16) ^ uint16(fileOffset)))
}
