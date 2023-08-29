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

// HeapOnNodeWriter represents a writer for pst.HeapOnNode.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#hn-heap-on-node
type HeapOnNodeWriter struct {
	// SignatureType represents the higher level data structure of this Heap-on-Node.
	SignatureType pst.SignatureType
}

// NewHeapOnNodeWriter creates a new HeapOnNodeWriter.
func NewHeapOnNodeWriter(signatureType pst.SignatureType) *HeapOnNodeWriter {
	return &HeapOnNodeWriter{
		SignatureType: signatureType,
	}
}

// Write writes the Heap-on-Node.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#creating-an-hn
func (heapOnNodeWriter *HeapOnNodeWriter) Write() error {
	if err := heapOnNodeWriter.WriteHeader(); err != nil {
		return eris.Wrap(err, "failed to write Heap-on-Node header")
	}

	return nil
}

// WriteHeader writes the Heap-on-Node header.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#hnhdr
func (heapOnNodeWriter *HeapOnNodeWriter) WriteHeader() error {
	// NOTE: This is the Heap-on-Node NOT BTree-On-Heap
	header := make([]byte, 12)

	// The byte offset to the HN page Map record (section 2.3.1.5), with respect to the beginning of the HNHDR structure.
	WriteBuffer(make([]byte, 2), header) // TODO

	// Block signature; MUST be set to 0xEC to indicate an HN.
	WriteBuffer([]byte{236}, header)

	// Client signature. This value describes the higher-level structure that is implemented on top of the HN.
	WriteBuffer([]byte{byte(heapOnNodeWriter.SignatureType)}, header)

	// HID that points to the User Root record.
	// The User Root record contains data that is specific to the higher level.
	WriteBuffer(make([]byte, 4), header) // TODO

	// Per-block Fill Level Map.
	// This array consists of eight 4-bit values that indicate the fill level for each of the first 8 data blocks (including this header block).
	// If the HN has fewer than 8 data blocks, then the values corresponding to the non-existent data blocks MUST be set to zero.
	WriteBuffer(make([]byte, 4), header) // TODO

	return nil
}
