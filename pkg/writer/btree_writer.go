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
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
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

// Write writes the B-Tree.
// References TODO
func (btreeWriter *BTreeWriter) Write() error {
	return nil
}

// WriteBTree writes the node- and block b-tree.
// References TODO
func (btreeWriter *BTreeWriter) WriteBTree(btreeType pst.BTreeType, level int) error {
	btree := make([]byte, 512) // Same size for Unicode and ANSI.

	if err := btreeWriter.WriteBTreeNode(); err != nil {
		return eris.Wrap(err, "failed to write b-tree node")
	}

	// The number of BTree entries stored in the page data.
	if level > 0 {
		// Branch
		WriteBuffer([]byte{}, btree)
	} else {
		// Leaf
		WriteBuffer([]byte{}, btree)
	}

	// The maximum number of entries that can fit inside the page data.
	WriteBuffer([]byte{255}, btree)

	// The size of each BTree entry, in bytes.
	if btreeType == pst.BTreeTypeNode && level == 0 {
		switch btreeWriter.FormatType {
		case pst.FormatTypeUnicode:
			WriteBuffer([]byte{32}, btree)
		case pst.FormatTypeANSI:
			WriteBuffer([]byte{16}, btree)
		default:
			return pst.ErrFormatTypeUnsupported
		}
	} else {
		switch btreeWriter.FormatType {
		case pst.FormatTypeUnicode:
			WriteBuffer([]byte{24}, btree)
		case pst.FormatTypeANSI:
			WriteBuffer([]byte{12}, btree)
		default:
			return pst.ErrFormatTypeUnsupported
		}
	}

	// The depth level of this page. Leaf pages have a level of zero, whereas intermediate pages have a level greater than 0.

	if btreeWriter.FormatType == pst.FormatTypeUnicode {
		// Padding; MUST be set to zero.
		WriteBuffer(make([]byte, 4), btree)
	}

	// A PAGETRAILER structure (section 2.2.2.7.1).

	return nil
}

func (btreeWriter *BTreeWriter) WriteBTreeNode() error {
	var btreeNodeSize int

	switch btreeWriter.FormatType {
	case pst.FormatTypeUnicode:
		btreeNodeSize = 488
	case pst.FormatTypeANSI:
		btreeNodeSize = 496
	default:
		return pst.ErrFormatTypeUnsupported
	}

	btreeNode := make([]byte, btreeNodeSize)

	//

	return nil
}
