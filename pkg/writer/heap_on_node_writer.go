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

// HeapOnNodeWriter represents a writer for pst.HeapOnNode.
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

// WriteTo writes the Heap-on-Node.
// References
// - https://github.com/mooijtech/go-pst/blob/main/docs/README.md#creating-an-hn
// - https://github.com/mooijtech/go-pst/blob/main/docs/README.md#creating-a-new-node
func (heapOnNodeWriter *HeapOnNodeWriter) WriteTo(writer io.Writer) (int64, error) {
	headerWrittenSize, err := heapOnNodeWriter.WriteHeader(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write header")
	}

	pageMapWrittenSize, err := heapOnNodeWriter.WritePageMap(writer, []byte{})

	if err != nil {
		return 0, eris.Wrap(err, "failed to write page map")
	}

	return headerWrittenSize + pageMapWrittenSize, nil
}

// WriteHeader writes the Heap-on-Node header.
// Returns the amount of bytes copied to the output buffer.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#hnhdr
func (heapOnNodeWriter *HeapOnNodeWriter) WriteHeader(writer io.Writer) (int64, error) {
	// 2+1+1+4+4
	header := bytes.NewBuffer(make([]byte, 12))

	// The byte offset to the page map record (see WritePageMap).
	// Offset starts in respect to the beginning of this header.
	header.Write([]byte{12})

	// Block signature.
	// MUST be set to 236 to indicate a Heap-on-Node.
	header.Write([]byte{236})

	// Client signature.
	// This value describes the higher-level structure that is implemented on top of the Heap-on-Node.
	header.WriteByte(byte(heapOnNodeWriter.SignatureType))

	// HID that points to the User Root record.
	// The User Root record contains data that is specific to the higher level.
	header.Write([]byte{32}) // 0x20 TODO check this

	// Per-block Fill Level Map.
	// This array consists of eight 4-bit values that indicate the fill level for each of the first 8 data blocks (including this header block).
	// If the HN has fewer than 8 data blocks, then the values corresponding to the non-existent data blocks MUST be set to zero.
	header.Write(make([]byte, 4)) // TODO

	return header.WriteTo(writer)
}

// WritePageMap writes the Heap-on-Node page map.
// Returns the amount of bytes copied to the output buffer.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#hnpagemap
func (heapOnNodeWriter *HeapOnNodeWriter) WritePageMap(writer io.Writer, allocations []byte) (int64, error) {
	// 2+2+allocations
	pageMap := bytes.NewBuffer(make([]byte, 4+len(allocations)))

	// Allocation count.
	pageMap.Write(GetUint16(2))

	// Free count.
	pageMap.Write(make([]byte, 2))

	// Allocations.
	pageMap.Write(allocations)

	return pageMap.WriteTo(writer)
}
