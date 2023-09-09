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

// FolderWriter represents a writer per folder.
type FolderWriter struct {
	// Writer represents the io.Writer to write to.
	Writer io.Writer
	// FormatType represents the FormatType used while writing.
	FormatType FormatType

	// Identifier represents the identifier of this folder.
	// This identifier is used to find the folder in the B-Tree.
	Identifier Identifier
	// PropertyContextWriter represents the writer for properties of this folder (see properties.Folder).
	PropertyContextWriter *PropertyContextWriter
	// FolderWriteCallbackChannel TODO
	FolderWriteCallbackChannel chan int64

	// SubFoldersWriteChannel represents the Go channel for writing sub-folders.
	SubFoldersWriteChannel chan *FolderWriter
	// SubFoldersWriteCallbackChannel represents the callback which is called after writing a sub-folder.
	SubFoldersWriteCallbackChannel chan SubFolderWriteResponse
	// SubFoldersTableContextWriter represents the sub-folders TableContextWriter of this folder.
	// Contains references to sub-folders (Identifier).
	SubFoldersTableContextWriter *TableContextWriter

	// MessageWriter represents the writer for messages.
	MessageWriter *MessageWriter
	// MessageWriteCallbackChannel represents the callback which is called after a message has been written.
	// The message is then added to the MessageTableContextWriter of this folder.
	MessageWriteCallbackChannel chan Identifier
	// MessageTableContextWriter represents the message TableContextWriter of the folder.
	// Contains references to messages (Identifier).
	MessageTableContextWriter *TableContextWriter
}

// NewFolderWriter creates a new FolderWriter.
// folderWriteCallback is used by the caller to calculate the total PST file size written.
func NewFolderWriter(writer io.Writer, writeGroup *errgroup.Group, formatType FormatType, folderWriteCallback chan int64) (*FolderWriter, error) {
	// propertyWriteCallback is started (StartFolderWriteChannel) below to send the writable properties to the PropertyContext.
	propertyWriteCallback := make(chan int64)
	// Adds the message identifiers to the message TableContext of the folder (located at folderIdentifier + 12)
	messageWriteCallback := make(chan int64)

	// TODO - folderIdentifier + 12?
	//subFoldersTableContext := NewTableContextWriter(identifier)

	// Create the folder writer.
	folderWriter := &FolderWriter{
		Writer:                     writer,
		FormatType:                 formatType,
		FolderWriteChannel:         make(chan *Folder),
		FolderWriteCallbackChannel: folderWriteCallback,
		MessageWriter:              NewMessageWriter(writer, writeGroup, messageWriteCallback, formatType),
		FolderTableContextWriter:   NewTableContextWriter(writer, writeGroup, formatType),
		MessageTableContextWriter:  NewTableContextWriter(writer, writeGroup, formatType),
		PropertyContextWriter:      NewPropertyContextWriter(writer, writeGroup, propertyWriteCallback, formatType, BTreeTypeBlock),
		Identifier:                 folderIdentifier,
	}

	// Start channel for writing folders.
	// The MessageWriter starts a channel for writing messages.
	go folderWriter.StartFolderWriteChannel(writeGroup)

	return folderWriter, nil
}

// SubFolderWriteResponse represents a Go channel response for when a sub-folder is written.
type SubFolderWriteResponse struct {
	// Identifier represents the Identifier of the written folder.
	Identifier Identifier
	// Written represents the written byte size of the folder.
	Written int64
}

// AddSubFolders adds the FolderWriter to the write queue.
// See StartSubFoldersWriteCallbackChannel.
func (folderWriter *FolderWriter) AddSubFolders(subFolders ...*FolderWriter) {
	for _, subFolder := range subFolders {
		folderWriter.SubFoldersWriteChannel <- subFolder
	}
}

// StartSubFoldersWriteCallbackChannel listens for written sub-folders.
// Adds the sub-folder Identifier to the SubFoldersTableContextWriter.
func (folderWriter *FolderWriter) StartSubFoldersWriteCallbackChannel() {
	for subFolderWriteResponse := range folderWriter.SubFoldersWriteCallbackChannel {
		folderWriter.SubFoldersTableContextWriter.AddIdentifier(subFolderWriteResponse.Identifier)
	}
}

// AddMessages the messages to the MessageWriter write queue of this folder.
// See StartMessageWriteCallbackChannel.
func (folderWriter *FolderWriter) AddMessages(messages ...proto.Message) {
	folderWriter.MessageWriter.AddMessages(folderWriter.Identifier, messages...)
}

// StartMessageWriteCallbackChannel listens for written messages.
// Adds the message Identifier to the MessageTableContextWriter.
func (folderWriter *FolderWriter) StartMessageWriteCallbackChannel() {
	for messageIdentifier := range folderWriter.MessageWriteCallbackChannel {
		folderWriter.MessageTableContextWriter.AddIdentifier(messageIdentifier)
	}
}

// SetIdentifier sets the identifier of the folder.
// This is mainly used for the pst.IdentifierRootFolder.
// Usually the identifier is automatically set by NewFolderWriter.
func (folderWriter *FolderWriter) SetIdentifier(identifier Identifier) {
	folderWriter.Identifier = identifier
}

// UpdateTableContext updates the TableContext of the folder to reference the message identifiers.
func (folderWriter *FolderWriter) UpdateTableContext(messages ...proto.Message) {

}

// StartFolderWriteChannel listens for sub-folders to write.
// The called is responsible for starting the write channel.
//func (folderWriter *FolderWriter) StartFolderWriteChannel(writeGroup *errgroup.Group) {
//	for receivedFolder := range folderWriter.FolderWriteChannel {
//		// Listen for folders to write.
//		writeGroup.Go(func() error {
//			// Add folder to TableContextWriter write queue.
//			folderWriter.TableContextWriter.AddFolders(receivedFolder)
//
//			return nil
//		})
//	}
//}

// WriteTo writes the folder containing messages.
// Returns the amount of bytes written to the output buffer.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#folders
func (folderWriter *FolderWriter) WriteTo(writer io.Writer) (int64, error) {
	// Write TableContext.
	tableContextWrittenSize, err := folderWriter.TableContextWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Table Context")
	}

	// Make this written folder findable in the B-Tree.
	if err := folderWriter.UpdateIdentifier(); err != nil {
		return 0, eris.Wrap(err, "failed to update identifier")
	}

	return tableContextWrittenSize + messagesWrittenSize, nil
}
