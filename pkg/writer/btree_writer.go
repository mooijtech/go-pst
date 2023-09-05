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

package writer

import (
	"bytes"
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
	"io"
)

// BTreeWriter represents a writer for B-Trees.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#btrees
type BTreeWriter struct {
	// FormatType represents the FormatType to use during writing.
	FormatType pst.FormatType
	// BTreeType represents the type of b-tree to write (node or block).
	BTreeType pst.BTreeType
	// BTreeNodes represents the B-Tree nodes to write.
	BTreeNodes []pst.BTreeNode
}

// NewBTreeWriter creates a new BTreeWriter.
func NewBTreeWriter(formatType pst.FormatType, btreeType pst.BTreeType, btreeNodes []pst.Identifier) *BTreeWriter {
	// Make writable B-Tree nodes.
	var btreeNodes

	for _, btreeNodeIdentifier := range btreeNodes {

	}

	return &BTreeWriter{
		FormatType: formatType,
		BTreeType:  btreeType,
		BTreeNodes: btreeNodes,
	}
}

// WriteTo writes the B-Tree.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#btpage
func (btreeWriter *BTreeWriter) WriteTo(writer io.Writer) (int64, error) {
	btree := bytes.NewBuffer(make([]byte, 512)) // Same for Unicode and ANSI.

	// This section contains entries of the BTree array.
	// If "cLevel" is 0, each entry is either of type "BBTENTRY" or "NBTENTRY" based on the "ptype" of the page.
	// TODO Check maximum b-tree entry size and return error on overflow
	for _, btreeEntry := range btreeWriter.BTreeNodes {
		if _, err := btreeEntry.WriteTo(btree); err != nil {
			return 0, eris.Wrap(err, "failed to write b-tree node")
		}
	}

	// The number of BTree entries stored in the page data.
	// The entries depend on the value of this field.
	btree.WriteByte(byte(len(btreeWriter.BTreeNodes)))

	// The maximum number of entries that can fit inside the page data.
	btree.Write(make([]byte, 1)) // TODO

	// The size of each BTree entry, in bytes.
	if btreeWriter.BTreeType == pst.BTreeTypeNode && level == 0 {
		switch btreeWriter.FormatType {
		case pst.FormatTypeUnicode:
			btree.Write([]byte{32})
		case pst.FormatTypeANSI:
			btree.Write([]byte{16})
		default:
			return 0, pst.ErrFormatTypeUnsupported
		}
	} else {
		switch btreeWriter.FormatType {
		case pst.FormatTypeUnicode:
			btree.Write([]byte{24})
		case pst.FormatTypeANSI:
			btree.Write([]byte{12})
		default:
			return 0, pst.ErrFormatTypeUnsupported
		}
	}

	// The depth level of this page.
	// Leaf pages have a level of 0, while intermediate pages have a level greater than 0.
	// This value determines the type of entries.
	btree.WriteByte(byte(level))

	if btreeWriter.FormatType == pst.FormatTypeUnicode {
		// Padding that should be set to zero.
		// Note that there is no padding in the ANSI version of the structure.
		btree.Write(make([]byte, 4))
	}

	// Page trailer.
	// A PAGETRAILER structure with specific subfield values.
	// The "ptype" subfield of "pageTrailer" should be set to "ptypeBBT" for a Block BTree page or "ptypeNBT" for a Node BTree page.
	// The other subfields of "pageTrailer" should be set as specified in the documentation.
	if _, err := btreeWriter.WritePageTrailer(btree, btreeType); err != nil {
		return 0, eris.Wrap(err, "failed to write page trailer")
	}

	return btree.WriteTo(writer)
}

// WritePageTrailer writes the page tailer of the b-tree.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#pagetrailer
func (btreeWriter *BTreeWriter) WritePageTrailer(writer io.Writer, btreeType pst.BTreeType) (int64, error) {
	pageTrailerBuffer := bytes.NewBuffer(make([]byte, 0))

	// This value indicates the type of data contained within the page.
	pageTrailerBuffer.WriteByte(byte(btreeWriter.BTreeType))
	// MUST be set to the same value as ptype.
	pageTrailerBuffer.WriteByte(byte(btreeWriter.BTreeType))
	// Page signature. This value depends on the value of the ptype field.
	// This value is zero (0x0000) for AMap, PMap, FMap, and FPMap pages.
	// For BBT, NBT, and DList pages, a page / block signature is computed (see section 5.5).
	// TODO - pageTrailerBuffer.Write()
	// 32-bit CRC of the page data, excluding the page trailer.
	// See section 5.3 for the CRC algorithm.
	// Note the locations of the dwCRC and bid are differs between the Unicode and ANSI version of this structure.
	// TODO - pageTrailerBuffer.Write()
	// The BID of the page's block.
	// AMap, PMap, FMap, and FPMap pages have a special convention where their BID is assigned the same value as their IB (that is, the absolute file offset of the page).
	// The bidIndex for other page types are allocated from the special bidNextP counter in the HEADER structure.
	// TODO - pageTrailerBuffer.Write()

	return 0, nil
}
