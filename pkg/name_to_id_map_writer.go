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

// NameToIDMapWriter defines a writer for the Name-to-ID-Map.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#name-to-id-map
type NameToIDMapWriter struct {
	PropertyContextWriter *PropertyContextWriter
}

// NewNameToIDMapWriter creates a new NameToIDMapWriter.
func NewNameToIDMapWriter(writer io.WriteSeeker, writeGroup *errgroup.Group, propertyContextWriteCallback chan int64, formatType FormatType) (*NameToIDMapWriter, error) {
	propertyContextWriter, err := NewPropertyContextWriter(writer, writeGroup, propertyContextWriteCallback, formatType)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create Property Context writer")
	}

	return &NameToIDMapWriter{
		PropertyContextWriter: propertyContextWriter,
	}, nil
}

func (nameToIDMapWriter *NameToIDMapWriter) WriteTo(writer io.Writer) (int64, error) {
	// The minimum requirement for the Name-to-ID Map is a PC node with a single property PidTagNameidBucketCount set to a value of 251 (0xFB)
	//return nameToIDMapWriter.PropertyContextWriter.
	return 0, nil
}
