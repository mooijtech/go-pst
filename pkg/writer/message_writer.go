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

// MessageWriter represents a message that can be written to a PST file.
type MessageWriter struct {
	// Properties represents the properties of a pst.Message.
	Properties *properties.Message
	// Attachments represents the attachments of a pst.Message.
	Attachments []*AttachmentWriter
	// PropertyContextWriter writes the pst.PropertyContext of a pst.Message.
	PropertyContextWriter *PropertyContextWriter
}

// NewMessageWriter creates a new MessageWriter.
func NewMessageWriter(properties *properties.Message, attachments []*AttachmentWriter) *MessageWriter {
	return &MessageWriter{
		Properties:            properties,
		Attachments:           attachments,
		PropertyContextWriter: NewPropertyContextWriter(properties),
	}
}

// WriteTo writes the message property context.
func (messageWriter *MessageWriter) WriteTo(writer io.Writer) (int64, error) {
	// Write Property Context.
	propertyContextWrittenSize, err := messageWriter.PropertyContextWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Property Context")
	}

	// Write attachments.
	var attachmentsWrittenSize int64

	for _, attachmentWriter := range messageWriter.Attachments {
		written, err := attachmentWriter.WriteTo(writer)

		if err != nil {
			return 0, eris.Wrap(err, "failed to write attachment")
		}

		attachmentsWrittenSize += written
	}

	return propertyContextWrittenSize + attachmentsWrittenSize, nil
}
