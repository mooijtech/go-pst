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
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"github.com/rotisserie/eris"
	"io"
)

// AttachmentWriter represents a writer for attachments.
type AttachmentWriter struct {
	// PropertyContextWriter represents the PropertyContextWriter.
	PropertyContextWriter *PropertyContextWriter
}

// NewAttachmentWriter creates a new AttachmentWriter.
func NewAttachmentWriter(properties *properties.Attachment) *AttachmentWriter {
	return &AttachmentWriter{
		PropertyContextWriter: NewPropertyContextWriter(properties),
	}
}

// WriteTo writes the attachment.
func (attachmentWriter *AttachmentWriter) WriteTo(writer io.Writer) (int64, error) {
	propertyContextWrittenSize, err := attachmentWriter.PropertyContextWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Table Context")
	}

	return propertyContextWrittenSize, nil
}
