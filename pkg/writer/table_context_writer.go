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

// Write writes the pst.TableContext.
func (tableContextWriter *TableContextWriter) Write() error {
	if err := tableContextWriter.BTreeOnHeapWriter.Write(); err != nil {
		return eris.Wrap(err, "failed to write BTree-on-Heap")
	}

	if err := tableContextWriter.WriteHeader(); err != nil {
		return eris.Wrap(err, "failed to write Table Context header")
	}

	return nil
}

// WriteHeader writes the pst.TableContext header.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#tcinfo
func (tableContextWriter *TableContextWriter) WriteHeader() error {
	// ?+?+?+entries
	header := make([]byte, 0) // TODO - fix size

	WriteBuffer([]byte{byte(pst.SignatureTypeTableContext)}, header)      // MUST be set to bTypeTC
	WriteBuffer([]byte{byte(len(tableContextWriter.Properties))}, header) // Column count

	// Array of Column Descriptors.

	return nil
}
