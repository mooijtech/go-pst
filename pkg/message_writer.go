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
// References
type MessageWriter struct {
	// streamWriter represents the Go channel used to process writing messages.
	// Has a callback called for each written message.
	streamWriter *StreamWriter
	// formatType represents the FormatType used while writing.
	formatType FormatType
	// attachmentWriter represents a writer for attachments.
	attachmentWriter *AttachmentWriter
	// propertyContextWriter writes the pst.PropertyContext of a pst.Message.
	propertyContextWriter *PropertyContextWriter
	// identifier represents the identifier of this message, which is used in the B-Tree.
	identifier Identifier
}

// NewMessageWriter creates a new MessageWriter.
func NewMessageWriter(outputFile io.WriteSeeker, writeGroup *errgroup.Group, parentFolderIdentifier Identifier, formatType FormatType) (*MessageWriter, error) {
	streamWriter := NewStreamWriter(outputFile, writeGroup)

	// Start the stream writer which is used by the MessageWriter.
	streamWriter.StartWriteChannel()

	//
	attachmentWriter, err := NewAttachmentWriter(outputFile, writeGroup, formatType)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create attachment writer")
	}

	propertyContextWriter, err := NewPropertyContextWriter(outputFile, writeGroup, propertyContextCallback, formatType)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create property context writer")
	}

	messageWriter := &MessageWriter{
		formatType:            formatType,
		streamWriter:          streamWriter,
		attachmentWriter:      attachmentWriter,
		propertyContextWriter: propertyContextWriter,
		// Identifier is set below.
	}

	// Set the identifier so this message can be found in the B-Tree.
	if err := messageWriter.UpdateIdentifier(parentFolderIdentifier); err != nil {
		return nil, eris.Wrap(err, "failed to update identifier")
	}

	return messageWriter, nil
}

// AddMessages adds the MessageWriter to the write queue.
func (messageWriter *MessageWriter) AddMessages(folderIdentifier Identifier, messages ...*MessageWriter) {
	// TODO - identifier = folderIdentifier + 12 --- Identifier of what?
	for _, message := range messages {
		messageWriter.streamWriter.Send(message)
	}
}

// AddAttachments adds AttachmentWriter to the write queue.
func (messageWriter *MessageWriter) AddAttachments(attachments ...*AttachmentWriter) {
	// TODO -
	//messageWriter.attachmentWriter.AddAttachments(attachments...)
}

func (messageWriter *MessageWriter) AddProperties(properties ...proto.Message) {
	//messageWriter.propertyContextWriter.AddProperties(properties)
}

// UpdateIdentifier
// References
func (messageWriter *MessageWriter) UpdateIdentifier(parentFolderIdentifier Identifier) error {
	messageWriter.identifier = parentFolderIdentifier + 12
	//identifier, err := NewIdentifier(messageWriter.FormatType)
	//
	//if err != nil {
	//	return eris.Wrap(err, "failed to create identifier")
	//}
	//
	//messageWriter.Identifier = identifier
}

// WriteTo writes the message property context.
func (messageWriter *MessageWriter) WriteTo(writer io.Writer) (int64, error) {
	// Write Property Context.
	propertyContextWrittenSize, err := messageWriter.propertyContextWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Property Context")
	}

	// Wait for attachments to write.
	messageWriter.attachmentWriter.Wait()

	// TODO - Wait for StreamWriter here?

	var attachmentsWrittenSize int64

	// TODO - Receive total written size from the callback.

	// TODO - Moved to New? Make this message findable in the B-Tree.
	//if err := messageWriter.UpdateIdentifier(); err != nil {
	//	return 0, eris.Wrap(err, "failed to update identifier")
	//}

	return propertyContextWrittenSize + attachmentsWrittenSize, nil
}
