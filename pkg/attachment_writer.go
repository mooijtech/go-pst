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

// AttachmentWriter represents a writer for attachments.
type AttachmentWriter struct {
	// streamWriter represents the Go channel for writing attachments.
	streamWriter *StreamWriter
	// propertyContextWriter represents the PropertyContextWriter.
	propertyContextWriter *PropertyContextWriter
	// attachmentWriteCallbackChannel is the callback for when attachments have been written.
	attachmentWriteCallbackChannel chan int64
}

// NewAttachmentWriter creates a new AttachmentWriter.
func NewAttachmentWriter(outputFile io.WriteSeeker, writeGroup *errgroup.Group, formatType FormatType) (*AttachmentWriter, error) {
	// Stream writer used to write the attachments.
	streamWriter := NewStreamWriter[io.WriterTo, int64](outputFile, writeGroup)

	// Property context writer.
	propertyContextWriteCallback := make(chan int64)
	propertyContextWriter, err := NewPropertyContextWriter(outputFile, writeGroup, propertyContextWriteCallback, formatType)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create property context writer")
	}

	// Attachment writer.
	attachmentWriter := &AttachmentWriter{
		propertyContextWriter: propertyContextWriter,
		streamWriter:          streamWriter,
		// TODO - attachmentWriteCallbackChannel: attachmentWriteCallbackChannel,
	}

	// Start the stream writer for writing attachments.
	streamWriter.StartWriteChannel()
	// Send write responses to the parent callback so the total write size can be calculated.
	// TODO - streamWriter.RegisterCallback(attachmentWriteCallbackChannel)

	return attachmentWriter, nil
}

// AddFile adds the properties of the attachment (properties.Attachment).
func (attachmentWriter *AttachmentWriter) AddFile(names ...string) error {
	//// Send to the write channel.
	//for _, name := range names {
	//	attachmentWriter.streamWriter.Send(NewWritableAttachment(name))
	//
	//	// TODO - Use AttachMethods :)
	//	//attachmentWriter.StreamWriter.Send(attachment)
	//}

	return nil
}

// AddProperties add the properties of the attachment to write.
func (attachmentWriter *AttachmentWriter) AddProperties(protoMessages ...proto.Message) error {
	for _, protoMessage := range protoMessages {
		if err := attachmentWriter.propertyContextWriter.AddProperties(protoMessage); err != nil {
			return eris.Wrap(err, "failed to add properties")
		}
	}

	return nil
}

// WriteTo writes the attachment.
func (attachmentWriter *AttachmentWriter) WriteTo(writer io.Writer) (int64, error) {
	propertyContextWrittenSize, err := attachmentWriter.propertyContextWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Table Context")
	}

	// Wait for attachments to be written.
	var attachmentWrittenSize int64

	for streamResponse := range attachmentWriter.streamWriter.callbackChannel {
		attachmentWrittenSize += streamResponse
	}

	// TODO - All writers need to wait for the callback channel.
	// TODO - Pass this up again?

	return propertyContextWrittenSize, nil
}

// Wait for the attachments to be written.
// Blocking call.
func (attachmentWriter *AttachmentWriter) Wait() int64 {
	var totalSize int64

	for streamResponse := range attachmentWriter.streamWriter.callbackChannel {
		totalSize += streamResponse
	}

	return totalSize
}
