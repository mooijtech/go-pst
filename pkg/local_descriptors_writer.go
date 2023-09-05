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
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
	"io"
)

// LocalDescriptorsWriter represents a writer for Local Descriptors (B-Tree Nodes pointing to other B-Tree Nodes).
// The LocalDescriptorsWriter can be used multiple times to create multiple local descriptors.
type LocalDescriptorsWriter struct {
	// Writer represents the io.Writer used while writing.
	Writer io.Writer
	// WriteGroup represents writers running in Goroutines.
	WriteGroup *errgroup.Group
	// FormatType represents the FormatType used while writing.
	FormatType FormatType
	// BTreeWriter represents the BTreeWriter.
	BTreeWriter *BTreeWriter
	// BlockWriter represents the BlocKWriter.
	BlockWriter *BlockWriter
	// Identifier represents the BTree node pst.Identifier of the Local Descriptor created after WriteTo has been called.
	// Set by UpdateIdentifier, called after writing the Local Descriptor using WriteTo.
	Identifier Identifier
}

// NewLocalDescriptorsWriter creates a new LocalDescriptorsWriter.
func NewLocalDescriptorsWriter(writer io.Writer, writeGroup *errgroup.Group, formatType FormatType, btreeType BTreeType) *LocalDescriptorsWriter {
	btreeWriteCallback := make(chan WriteCallbackResponse)
	btreeWriter := NewBTreeWriter(writer, writeGroup, btreeWriteCallback, formatType, btreeType)
	blockWriter := NewBlockWriter(formatType)

	return &LocalDescriptorsWriter{
		Writer:      writer,
		WriteGroup:  writeGroup,
		FormatType:  formatType,
		BTreeWriter: btreeWriter,
		BlockWriter: blockWriter,
	}
}

// AddBTreeNodes adds B-Tree nodes to the Local Descriptor.
func (localDescriptorsWriter *LocalDescriptorsWriter) AddBTreeNodes(btreeNode ...BTreeNode) {
	localDescriptorsWriter.BTreeWriter.AddBTreeNodes(btreeNode...)
}

// AddProperty adds a Property to the Local Descriptors.
func (localDescriptorsWriter *LocalDescriptorsWriter) AddProperty(property Property) {
	// Create a B-Tree node for this property.
}

// WriteTo writes the Local Descriptors.
func (localDescriptorsWriter *LocalDescriptorsWriter) WriteTo(writer io.Writer) (int64, error) {
	// Set the Local Descriptors identifier.
	if err := localDescriptorsWriter.UpdateIdentifier(); err != nil {
		return 0, eris.Wrap(err, "failed to update identifier")
	}

	return 0, nil
}

// UpdateIdentifier sets the identifier of the local descriptors so it can be found in the B-Tree.
// Called after WriteTo.
func (localDescriptorsWriter *LocalDescriptorsWriter) UpdateIdentifier() error {
	identifier, err := NewIdentifier(localDescriptorsWriter.FormatType)

	if err != nil {
		return eris.Wrap(err, "failed to create identifier")
	}

	localDescriptorsWriter.Identifier = identifier

	return nil
}
