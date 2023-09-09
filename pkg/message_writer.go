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
	"google.golang.org/protobuf/proto"
	"io"
)

// MessageWriter represents a writer for messages.
type MessageWriter struct {
	// Writer represents the io.Writer used while writing.
	Writer io.WriteSeeker
	// WriteGroup represents the writers running in Goroutines.
	WriteGroup *errgroup.Group
	// FormatType represents the FormatType used while writing.
	FormatType FormatType
	// MessageWriteChannel represents the Go channel used to process writing messages.
	MessageWriteChannel chan *WritableMessage
	// MessageWriteCallback represents the callback called for each written message.
	MessageWriteCallback chan int64
	// AttachmentWriter represents a writer for attachments.
	AttachmentWriter *AttachmentWriter
	// PropertyContextWriter writes the pst.PropertyContext of a pst.Message.
	PropertyContextWriter *PropertyContextWriter
	// Identifier represents the identifier of this message, which is used in the B-Tree.
	Identifier Identifier
}

// NewMessageWriter creates a new MessageWriter.
func NewMessageWriter(writer io.WriteSeeker, writeGroup *errgroup.Group, messageWriteCallback chan int64, formatType FormatType) *MessageWriter {
	messageWriter := &MessageWriter{
		Writer:                writer,
		WriteGroup:            writeGroup,
		FormatType:            formatType,
		AttachmentWriter:      NewAttachmentWriter(writer, writeGroup, formatType),
		PropertyContextWriter: NewPropertyContextWriter(writer, writeGroup, formatType),
		MessageWriteCallback:  messageWriteCallback,
	}

	// Start the message write channel which processes writing messages.
	messageWriter.StartMessageWriteChannel()

	return messageWriter
}

// AddMessages adds the WritableMessage to the write queue.
func (messageWriter *MessageWriter) AddMessages(folderIdentifier Identifier, messages ...proto.Message) {
	// TODO - identifier = folderIdentifier + 12
	for _, message := range messages {
		messageWriter.MessageWriteChannel <- message
	}
}

// StartMessageWriteChannel receives messages to write.
func (messageWriter *MessageWriter) StartMessageWriteChannel() {
	messageWriter.WriteGroup

	// The caller already starts a Goroutine.
	for receivedMessage := range messageWriter.MessageWriteChannel {
		writeGroup.Go(func() error {
			// Write the message.
			messageWrittenSize, err := receivedMessage.WriteTo(messageWriter.Writer)

			if err != nil {
				return eris.Wrap(err, "failed to write message")
			}

			// Callback to keep track of the total PST file size.
			messageWriter.MessageWriteCallback <- NewWriteCallbackResponse(messageWrittenSize)

			return nil
		})
	}
}

// AddAttachments adds AttachmentWriter to the write queue.
func (messageWriter *MessageWriter) AddAttachments(attachments ...*AttachmentWriter) {
	messageWriter.AttachmentWriter.AddAttachments(attachments...)
}

// WritableMessage represents a writable message.
// TODO - Maybe merge to Message.
type WritableMessage struct {
	Identifier Identifier
	Properties proto.Message
}

func (messageWriter *MessageWriter) UpdateIdentifier() error {
	identifier, err := NewIdentifier(messageWriter.FormatType)

	if err != nil {
		return eris.Wrap(err, "failed to create identifier")
	}

	messageWriter.Identifier = identifier
}

// WriteTo writes the message property context.
func (messageWriter *MessageWriter) WriteTo(writer io.Writer) (int64, error) {
	// Write Property Context.
	propertyContextWrittenSize, err := messageWriter.PropertyContextWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Property Context")
	}

	// Wait for attachments to write.
	var attachmentsWrittenSize int64

	for _, attachmentWriter := range messageWriter.Attachments {
		written, err := attachmentWriter.WriteTo(writer)

		if err != nil {
			return 0, eris.Wrap(err, "failed to write attachment")
		}

		attachmentsWrittenSize += written
	}

	// Make this message findable in the B-Tree.
	if err := messageWriter.UpdateIdentifier(); err != nil {
		return 0, eris.Wrap(err, "failed to update identifier")
	}

	return propertyContextWrittenSize + attachmentsWrittenSize, nil
}
