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

// BTreeOnHeapWriter writes a BTree-On-Heap
type BTreeOnHeapWriter struct {
	// HeapOnNodeWriter represents the HeapOnNodeWriter.
	HeapOnNodeWriter *HeapOnNodeWriter
}

// NewBTreeOnHeapWriter creates a new BTreeOnHeapWriter.
func NewBTreeOnHeapWriter(heapOnNodeWriter *HeapOnNodeWriter) *BTreeOnHeapWriter {
	return &BTreeOnHeapWriter{HeapOnNodeWriter: heapOnNodeWriter}
}

// WriteTo writes the BTree-on-Heap.
// References:
// - https://github.com/mooijtech/go-pst/blob/main/docs/README.md#creating-a-bth
// - https://github.com/mooijtech/go-pst/blob/main/docs/README.md#inserting-into-the-bth
func (btreeOnHeapWriter *BTreeOnHeapWriter) WriteTo(writer io.Writer) (int64, error) {
	heapOnNodeWrittenSize, err := btreeOnHeapWriter.HeapOnNodeWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Heap-on-Node")
	}

	headerWrittenSize, err := btreeOnHeapWriter.WriteHeader(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write BTree-On-Heap header")
	}

	return heapOnNodeWrittenSize + headerWrittenSize, nil
}

// WriteHeader writes the BTree-on-Heap header.
// References:
// - https://github.com/mooijtech/go-pst/blob/main/docs/README.md#bthheader
func (btreeOnHeapWriter *BTreeOnHeapWriter) WriteHeader(writer io.Writer) (int64, error) {
	// 1+1+1+1+4
	header := bytes.NewBuffer(make([]byte, 8))

	// MUST be bTypeBTH.
	header.WriteByte(byte(pst.SignatureTypeBTreeOnHeap))
	// Size of the BTree Key value, in bytes.
	// This value MUST be set to 2, 4, 8, or 16.
	header.Write([]byte{8})
	// Size of the data value, in bytes.
	// This MUST be greater than zero and less than or equal to 32.
	header.Write(make([]byte, 32))
	// Index depth.
	// This number indicates how many levels of intermediate indices exist in the BTH.
	header.Write([]byte{0})
	// This is the HID (heap ID) that points to the entries of this BTree-on-Heap header.
	// The data consists of an array of BTH records.
	header.Write(make([]byte, 4))

	return header.WriteTo(writer)
}
