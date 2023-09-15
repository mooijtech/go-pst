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
	// streamWriter represents the Go channel for writing sub-folders.
	streamWriter *StreamWriter
	// folderWriteCallback sends the write response to the parent to eventually calculate the total size.
	folderWriteCallback chan int64
	// formatType represents the FormatType used while writing.
	formatType FormatType
	// identifier represents the identifier of this folder.
	// This identifier is used to find the folder in the B-Tree.
	identifier Identifier
	// parentFolderIdentifier represents the identifier of the parent folder.
	parentFolderIdentifier Identifier
	// propertyContextWriter represents the writer for properties of this folder (see properties.Folder).
	propertyContextWriter *PropertyContextWriter
	// subFoldersTableContextWriter represents the sub-folders TableContextWriter of this folder.
	// Contains references to sub-folders (Identifier).
	subFoldersTableContextWriter *TableContextWriter
	// messageWriter represents the writer for messages.
	// Callback is used to add the message identifiers to the MessageTableContextWriter of this folder.
	messageWriter *MessageWriter
	// messageTableContextWriter represents the message TableContextWriter of the folder.
	// Contains references to messages (Identifier).
	messageTableContextWriter *TableContextWriter
}

// NewFolderWriter creates a new FolderWriter.
// folderWriteCallback is used by the caller to calculate the total PST file size written.
func NewFolderWriter(outputFile io.WriteSeeker, writeGroup *errgroup.Group, formatType FormatType, folderWriteCallback chan int64, parentFolderIdentifier Identifier) (*FolderWriter, error) {
	// The sub-folders Table Context (containing identifiers) can be found at parent identifier + 12.
	// References:
	// - https://github.com/mooijtech/go-pst/blob/main/docs/README.md#locating-sub-folder-object-nodes
	// - https://github.com/mooijtech/go-pst/blob/main/docs/README.md#adding-a-sub-folder-object
	subFoldersTableContextWriter, err := NewTableContextWriter(outputFile, writeGroup, parentFolderIdentifier+12, formatType)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create table context writer")
	}

	// Folder identifier for the B-Tree.
	folderIdentifier, err := NewIdentifier(formatType)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create identifier")
	}

	// Message writer (adds messages to this folder).
	messageWriter, err := NewMessageWriter(outputFile, writeGroup, folderIdentifier, formatType)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create message writer")
	}

	// Table Context which contains identifiers to the messages of this folder.
	messageTableContextWriter, err := NewTableContextWriter(outputFile, writeGroup, formatType)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create new table context writer")
	}

	// TODO - Maybe better as one New
	messageTableContextWriter.WithParentIdentifier(parentFolderIdentifier)

	// Writer for properties (see properties.Folder).
	propertyContextWriteCallback := make(chan int64)
	propertyContextWriter, err := NewPropertyContextWriter(outputFile, writeGroup, propertyContextWriteCallback, formatType)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create Property Context writer")
	}

	// Handles writing folders (and sub-folders).
	streamWriter := NewStreamWriter[*FolderWriter, FolderWriteResponse](outputFile, writeGroup)

	// Create the folder writer.
	folderWriter := &FolderWriter{
		streamWriter:                 streamWriter,
		folderWriteCallback:          folderWriteCallback,
		formatType:                   formatType,
		identifier:                   folderIdentifier,
		parentFolderIdentifier:       parentFolderIdentifier,
		propertyContextWriter:        propertyContextWriter,
		subFoldersTableContextWriter: subFoldersTableContextWriter,
		messageWriter:                messageWriter,
		messageTableContextWriter:    messageTableContextWriter,
	}

	// Start the stream writer for writing folders.
	streamWriter.StartWriteChannel()
	// Handle write responses.
	//streamWriter.RegisterCallback(folderWriter.HandleFolderWriteCallback())

	return folderWriter, nil
}

// NewRootFolderWriter creates a new FolderWriter with Identifier 0.
func NewRootFolderWriter(outputFile io.WriteSeeker, writeGroup *errgroup.Group, formatType FormatType, folderWriteCallback chan int64) (*FolderWriter, error) {
	return NewFolderWriter(outputFile, writeGroup, formatType, folderWriteCallback, Identifier(0))
}

// FolderWriteResponse represents a Go channel response for when a sub-folder is written.
type FolderWriteResponse struct {
	// Identifier represents the Identifier of the written folder.
	Identifier Identifier
	// Written represents the written byte size of the folder.
	Written int64
}

// AddSubFolders adds the FolderWriter to the write queue.
func (folderWriter *FolderWriter) AddSubFolders(subFolders ...*FolderWriter) {
	for _, folder := range subFolders {
		folderWriter.streamWriter.Send(folder)
	}
}

// AddMessages the messages to the MessageWriter write queue of this folder.
// See StartMessageWriteCallbackChannel.
func (folderWriter *FolderWriter) AddMessages(messages ...*MessageWriter) {
	folderWriter.messageWriter.AddMessages(folderWriter.identifier, messages...)
}

// HandleFolderWriteCallback handles folder write callbacks.
// Add the written folder to the FolderTableContextWriter.
// Send write responses to parent writer.
func (folderWriter *FolderWriter) handleFolderWriteCallback(folderWriteResponse FolderWriteResponse) error {
	// Add the folder identifier to the folder Table Context so this folder can find the sub-folders.
	folderWriter.subFoldersTableContextWriter.AddIdentifier(folderWriteResponse.Identifier)
	// Send to parent so the total size of the PST file can be calculated.
	folderWriter.folderWriteCallback <- folderWriteResponse.Written

	return nil
}

// SetIdentifier sets the identifier of the folder.
// This is mainly used for the pst.IdentifierRootFolder.
// Usually the identifier is set by NewFolderWriter.
func (folderWriter *FolderWriter) SetIdentifier(identifier Identifier) {
	folderWriter.identifier = identifier
}

// GetIdentifier returns the identifier of this folder.
// Used to reference parent/child folders and messages.
func (folderWriter *FolderWriter) GetIdentifier() Identifier {
	return folderWriter.identifier
}

// UpdateTableContext updates the TableContext of the folder to reference the message identifiers.
func (folderWriter *FolderWriter) UpdateTableContext(messages ...proto.Message) {
	folderWriter.messageTableContextWriter.Add(messages...)
}

func (folderWriter *FolderWriter) AddProperties(properties ...proto.Message) {
	// TODO - folderWriter.propertyContextWriter.AddProperties(properties)
}

// WriteTo writes the folder containing messages.
// Returns the amount of bytes written to the output buffer.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#folders
func (folderWriter *FolderWriter) WriteTo(writer io.Writer) (int64, error) {
	// Everything is automatically started writing add soon as the Add is called.
	// TODO - We can block here if the user wants to wait for WriteTo call.

	// TODO - Moved to New? Make this written folder findable in the B-Tree.
	//if err := folderWriter.UpdateIdentifier(); err != nil {
	//	return 0, eris.Wrap(err, "failed to update identifier")
	//}

	//return tableContextWrittenSize + messagesWrittenSize, nil
	return 0, nil
}
