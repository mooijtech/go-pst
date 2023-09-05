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

// FolderWriter represents a writer for folders.
type FolderWriter struct {
	// Writer represents the io.Writer to write to.
	Writer io.Writer
	// FormatType represents the FormatType used while writing.
	FormatType FormatType
	// FolderWriteChannel represents the Go channel for writing sub-folders.
	FolderWriteChannel chan *FolderWriter
	// FolderWriteCallback represents the callback which is called after writing a folder.
	FolderWriteCallback chan WriteCallbackResponse
	// MessageWriter represents the writer for messages.
	MessageWriter *MessageWriter
	// TableContextWriter writes the pst.TableContext of the pst.Folder.
	TableContextWriter *TableContextWriter
	// PropertyWriter represents the writer for properties.
	PropertyWriter *PropertyWriter
	// PropertyWriteChannel represents the Go channel for writing []pst.Property.
	//PropertyWriteChannel chan *PropertyWriter
	// Identifier represents the identifier of this folder.
	Identifier Identifier
}

// NewFolderWriter creates a new FolderWriter.
func NewFolderWriter(writer io.Writer, writeGroup *errgroup.Group, formatType FormatType) *FolderWriter {
	folderWriter := &FolderWriter{
		Writer:             writer,
		FormatType:         formatType,
		MessageWriter:      NewMessageWriter(writeGroup, formatType),
		TableContextWriter: NewTableContextWriter(writer, writeGroup, formatType),
	}

	// Start channels for writing folders and messages.
	go folderWriter.StartFolderWriteChannel(writeGroup)
	go folderWriter.StartMessageWriteChannel(writeGroup)

	return folderWriter
}

// SetIdentifier sets the identifier of the folder.
// This is mainly used for the pst.IdentifierRootFolder.
// Usually the identifier is automatically set by UpdateIdentifier after a WriteTo call.
func (folderWriter *FolderWriter) SetIdentifier(identifier Identifier) {
	folderWriter.Identifier = identifier
}

// UpdateIdentifier sets the identifier of the folder, so it can be identified in the B-Tree.
// Called after WriteTo.
func (folderWriter *FolderWriter) UpdateIdentifier() error {
	identifier, err := NewIdentifier(folderWriter.FormatType)

	if err != nil {
		return eris.Wrap(err, "failed to create identifier")
	}

	folderWriter.Identifier = identifier

	return nil
}

// AddProperties writes the properties of this folder.
// TODO - Extend properties.Folder to include everything in [MS-OXCFOLD]: Folder Object Protocol.
func (folderWriter *FolderWriter) AddProperties(properties ...proto.Message) {
	folderWriter.TableContextWriter.AddProperties(properties...)
}

// AddMessages adds a message to the channel to be written.
func (folderWriter *FolderWriter) AddMessages(messages ...proto.Message) {
	folderWriter.MessageWriter.AddProperties(messages...)
}

// AddFolder queues the folder to be written, picked up by a Go channel.
// The folders are added to the TableContextWriter.
func (folderWriter *FolderWriter) AddFolder(folders ...*FolderWriter) {
	for _, folder := range folders {
		folderWriter.FolderWriteChannel <- folder
	}
}

// StartFolderWriteChannel listens for sub-folders to write.
// The called is responsible for starting the write channel.
func (folderWriter *FolderWriter) StartFolderWriteChannel(writeGroup *errgroup.Group) {
	for receivedFolder := range folderWriter.FolderWriteChannel {
		// Listen for folders to write.
		writeGroup.Go(func() error {
			// Add folder to TableContextWriter write queue.
			folderWriter.TableContextWriter.AddFolder(receivedFolder)

			return nil
		})
	}
}

// WriteTo writes the folder containing messages.
// Returns the amount of bytes written to the output buffer.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#folders
func (folderWriter *FolderWriter) WriteTo(writer io.Writer) (int64, error) {
	// Write TableContext.
	tableContextWrittenSize, err := folderWriter.TableContextWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Table Context")
	}

	// Write messages.
	messagesWrittenSize, err := folderWriter.WriteMessages(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write messages")
	}

	// Make this written folder findable in the B-Tree.
	if err := folderWriter.UpdateIdentifier(); err != nil {
		return 0, eris.Wrap(err, "failed to update identifier")
	}

	return tableContextWrittenSize + messagesWrittenSize, nil
}

// WriteMessages writes the messages of the folder.
func (folderWriter *FolderWriter) WriteMessages(writer io.Writer) (int64, error) {
	var totalSize int64

	for _, messageWriter := range folderWriter.Messages {
		written, err := messageWriter.WriteTo(writer)

		if err != nil {
			return 0, eris.Wrap(err, "failed to write message")
		}

		totalSize += written
	}

	return totalSize, nil
}
