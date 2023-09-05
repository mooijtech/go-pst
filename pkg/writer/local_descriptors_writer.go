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
	"crypto/rand"
	"encoding/binary"
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
	"io"
)

// LocalDescriptorsWriter represents a writer for Local Descriptors (B-Tree Nodes pointing to other B-Tree Nodes).
// The LocalDescriptorsWriter can be used multiple times to create multiple local descriptors.
type LocalDescriptorsWriter struct {
	// FormatType represents the FormatType used while writing.
	FormatType pst.FormatType
	// BTreeWriter represents the BTreeWriter.
	BTreeWriter *BTreeWriter
	// BlockWriter represents the BlocKWriter.
	BlockWriter *BlockWriter
	// Identifier represents the BTree node pst.Identifier of the Local Descriptor created after WriteTo has been called.
	// Set by UpdateIdentifier, called after writing the Local Descriptor WriteTo.
	Identifier pst.Identifier
}

// NewLocalDescriptorsWriter creates a new LocalDescriptorsWriter.
func NewLocalDescriptorsWriter(formatType pst.FormatType, btreeType pst.BTreeType, btreeNodes []pst.Identifier) *LocalDescriptorsWriter {
	return &LocalDescriptorsWriter{
		FormatType:  formatType,
		BTreeWriter: NewBTreeWriter(formatType, btreeType, btreeNodes),
		BlockWriter: NewBlockWriter(formatType, btreeNodes),
	}
}

// WriteTo writes the Local Descriptors.
func (localDescriptorsWriter *LocalDescriptorsWriter) WriteTo(writer io.Writer) (int64, error) {
	// Create B-Tree nodes for the Local Descriptors.
	if _, err := localDescriptorsWriter.BTreeWriter.WriteTo(writer); err != nil {
		return 0, eris.Wrap(err, "failed to write B-Tree nodes for the Local Descriptors")
	}

	return 0, nil
}

// GetIdentifier returns the identifier (pst.Identifier) of the written local descriptor.
// This will return an error if called before WriteTo has been called.
// References
func (localDescriptorsWriter *LocalDescriptorsWriter) UpdateIdentifier() (pst.Identifier, error) {
	var identifierSize int

	switch localDescriptorsWriter.FormatType {
	case pst.FormatTypeUnicode:
		identifierSize = 8
	case pst.FormatTypeANSI:
		identifierSize = 4
	default:
		return 0, pst.ErrFormatTypeUnsupported
	}

	identifierBytes := make([]byte, identifierSize)

	if _, err := rand.Read(identifierBytes); err != nil {
		return 0, eris.Wrap(err, "failed to read random bytes")
	}

	var identifier int64

	switch localDescriptorsWriter.FormatType {
	case pst.FormatTypeUnicode:
		identifier = int64(binary.LittleEndian.Uint64(identifierBytes))
	case pst.FormatTypeANSI:
		identifier = int64(binary.LittleEndian.Uint32(identifierBytes))
	}

	return pst.Identifier(identifier), nil
}
