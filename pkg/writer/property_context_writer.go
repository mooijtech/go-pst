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
	"google.golang.org/protobuf/proto"
)

// PropertyContextWriter represents a writer for a pst.PropertyContext.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#property-context-pc
type PropertyContextWriter struct {
	// Properties represents the properties in the pst.PropertyContext.
	// See properties.Message, properties.Attachment etc.
	Properties proto.Message
	// BTreeOnHeapWriter represents the BTreeOnHeapWriter.
	BTreeOnHeapWriter *BTreeOnHeapWriter
}

// NewPropertyContextWriter creates a new PropertyContextWriter.
func NewPropertyContextWriter(properties proto.Message) *PropertyContextWriter {
	heapOnNodeWriter := NewHeapOnNodeWriter(pst.SignatureTypePropertyContext)
	btreeOnHeapWriter := NewBTreeOnHeapWriter(heapOnNodeWriter)

	return &PropertyContextWriter{
		Properties:        properties,
		BTreeOnHeapWriter: btreeOnHeapWriter,
	}
}

// Write writes the pst.PropertyContext.
func (propertyContextWriter *PropertyContextWriter) Write() error {
	if err := propertyContextWriter.BTreeOnHeapWriter.Write(); err != nil {
		return eris.Wrap(err, "failed to write Heap-on-Node")
	}

	return nil
}
