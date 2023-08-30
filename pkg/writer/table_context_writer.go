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

// TableContextWriter represents a writer for a pst.TableContext.
type TableContextWriter struct {
	// BTreeOnHeapWriter represents the BTreeOnHeapWriter.
	BTreeOnHeapWriter *BTreeOnHeapWriter
	// Properties represents the pst.TableContext properties.
	Properties [][]pst.Property // TODO - Init ??
}

// NewTableContextWriter creates a new TableContextWriter.
func NewTableContextWriter() *TableContextWriter {
	heapOnNodeWriter := NewHeapOnNodeWriter(pst.SignatureTypeTableContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)

	return &TableContextWriter{
		BTreeOnHeapWriter: btreeOnHeapWriter,
	}
}

// WriteTo writes the pst.TableContext.
func (tableContextWriter *TableContextWriter) WriteTo(writer io.Writer) (int64, error) {
	btreeOnHeapWrittenSize, err := tableContextWriter.BTreeOnHeapWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write BTree-on-Heap")
	}

	headerWrittenSize, err := tableContextWriter.WriteHeader(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Table Context header")
	}

	return btreeOnHeapWrittenSize + headerWrittenSize, nil
}

// WriteHeader writes the pst.TableContext header.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcinfo
func (tableContextWriter *TableContextWriter) WriteHeader(writer io.Writer) (int64, error) {
	// ?+?+?+entries
	header := bytes.NewBuffer(make([]byte, 0)) // TODO - fix size

	// MUST be set to bTypeTC
	header.Write([]byte{byte(pst.SignatureTypeTableContext)})
	// Column count
	header.Write([]byte{byte(len(tableContextWriter.Properties))})

	// Array of Column Descriptors.

	return header.WriteTo(writer)
}
