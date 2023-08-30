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
type BTreeWriter struct {
	// FormatType represents the FormatType to use during writing.
	FormatType pst.FormatType
}

// NewBTreeWriter creates a new BTreeWriter.
func NewBTreeWriter(formatType pst.FormatType) *BTreeWriter {
	return &BTreeWriter{FormatType: formatType}
}

// WriteTo writes the B-Tree.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#btpage
func (btreeWriter *BTreeWriter) WriteTo(writer io.Writer, btreeType pst.BTreeType, btreeEntries [][]byte, level int) (int64, error) {
	btree := bytes.NewBuffer(make([]byte, 512)) // Same for Unicode and ANSI.

	// Entries
	// TODO make a structure for this.
	// TODO Check maximum b-tree entry size and return error on overflow
	for _, btreeEntry := range btreeEntries {
		btree.Write(btreeEntry)
	}

	// The number of BTree entries stored in the page data.
	btree.WriteByte(byte(len(btreeEntries)))

	// The maximum number of entries that can fit inside the page data.
	btree.Write(make([]byte, 1)) // TODO

	// The size of each BTree entry, in bytes.
	if btreeType == pst.BTreeTypeNode && level == 0 {
		switch btreeWriter.FormatType {
		case pst.FormatTypeUnicode:
			btree.Write([]byte{32})
		case pst.FormatTypeANSI:
			btree.Write([]byte{16})
		default:
			panic(pst.ErrFormatTypeUnsupported)
		}
	} else {
		switch btreeWriter.FormatType {
		case pst.FormatTypeUnicode:
			btree.Write([]byte{24})
		case pst.FormatTypeANSI:
			btree.Write([]byte{12})
		default:
			panic(pst.ErrFormatTypeUnsupported)
		}
	}

	// The depth level of this page.
	btree.WriteByte(byte(level))

	if btreeWriter.FormatType == pst.FormatTypeUnicode {
		// Padding; MUST be set to zero.
		// Unicode only.
		btree.Write(make([]byte, 4))
	}

	// Page trailer.
	if _, err := btreeWriter.WritePageTrailer(btree); err != nil {
		return 0, eris.Wrap(err, "failed to write page trailer")
	}

	return btree.WriteTo(writer)
}

// WritePageTrailer writes the page tailer of the b-tree.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#pagetrailer
func (btreeWriter *BTreeWriter) WritePageTrailer(writer io.Writer) (int64, error) {
	return 0, nil
}

// WriteBTreeNode writes the b-tree node entry.
func (btreeWriter *BTreeWriter) WriteBTreeNode(writer io.Writer) (int64, error) {
	var btreeNodeSize int

	switch btreeWriter.FormatType {
	case pst.FormatTypeUnicode:
		btreeNodeSize = 488
	case pst.FormatTypeANSI:
		btreeNodeSize = 496
	default:
		panic(pst.ErrFormatTypeUnsupported)
	}

	btreeNode := bytes.NewBuffer(make([]byte, btreeNodeSize))

	// TODO -

	return btreeNode.WriteTo(writer)
}
