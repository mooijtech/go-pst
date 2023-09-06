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
	"github.com/mooijtech/go-pst/v6/pkg/properties"
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
	FolderWriteChannel chan *Folder
	// FolderWriteCallback represents the callback which is called after writing a folder.
	FolderWriteCallback chan int64
	// FolderTableContextWriter writes the pst.TableContext of the pst.Folder.
	// A TableContext has identifiers for finding other folders.
	FolderTableContextWriter *TableContextWriter
	// PropertyContextWriter represents the writer for properties of this folder (see properties.Folder).
	PropertyContextWriter *PropertyContextWriter
	// MessageWriter represents the writer for messages.
	MessageWriter *MessageWriter
	// MessageWriteCallbackChannel represents the callback which is called after a message has been written.
	// The message is then added to the MessageTableContextWriter of this folder.
	MessageWriteCallbackChannel chan Identifier
	// MessageTableContextWriter represents the message TableContext of the folder (identifiers for messages in the B-Tree).
	// Property ID 26610 contains the message property identifier.
	MessageTableContextWriter *TableContextWriter
	// Identifier represents the identifier of this folder.
	// The identifier is used to find this folder in the B-Tree.
	Identifier Identifier
}

// NewFolderWriter creates a new FolderWriter.
// folderWriteCallback is used by the caller to calculate the total PST file size written.
func NewFolderWriter(writer io.Writer, writeGroup *errgroup.Group, formatType FormatType, folderWriteCallback chan int64) (*FolderWriter, error) {
	// propertyWriteCallback is started (StartFolderWriteChannel) below to send the writable properties to the PropertyContext.
	propertyWriteCallback := make(chan int64)
	// Adds the message identifiers to the message TableContext of the folder (located at folderIdentifier + 12)
	messageWriteCallback := make(chan int64)

	// Create the folder writer.
	folderWriter := &FolderWriter{
		Writer:                    writer,
		FormatType:                formatType,
		FolderWriteChannel:        make(chan *Folder),
		FolderWriteCallback:       folderWriteCallback,
		MessageWriter:             NewMessageWriter(writer, writeGroup, messageWriteCallback, formatType),
		FolderTableContextWriter:  NewTableContextWriter(writer, writeGroup, formatType),
		MessageTableContextWriter: NewTableContextWriter(writer, writeGroup, formatType),
		PropertyContextWriter:     NewPropertyContextWriter(writer, writeGroup, propertyWriteCallback, formatType, BTreeTypeBlock),
		Identifier:                folderIdentifier,
	}

	// Start channel for writing folders.
	// The MessageWriter starts a channel for writing messages.
	go folderWriter.StartFolderWriteChannel(writeGroup)

	return folderWriter, nil
}

// SetIdentifier sets the identifier of the folder.
// This is mainly used for the pst.IdentifierRootFolder.
// Usually the identifier is automatically set by NewFolderWriter.
func (folderWriter *FolderWriter) SetIdentifier(identifier Identifier) {
	folderWriter.Identifier = identifier
}

// Add adds the folder or message to this folder.
func (folderWriter *FolderWriter) Add(protoMessages ...proto.Message) error {
	for _, protoMessage := range protoMessages {
		switch writableProperties := protoMessage.(type) {
		case *properties.Folder:
			// Add a folder.
			folderWriter.FolderTableContextWriter.Add(writableProperties)
		case *properties.Message:
			// Add a message.
			folderWriter.MessageTableContextWriter.Add(writableProperties)
		default:
			return eris.New("unsupported properties passed to Add")
		}
	}

	return nil
}

// AddFolders adds the Folder to the write queue.
func (folderWriter *FolderWriter) AddFolders(folders ...*Folder) {
	for _, folder := range folders {
		folderWriter.FolderWriteChannel <- folder
	}
}

// AddMessages adds a message to the MessageWriter, picked up by Goroutines.
func (folderWriter *FolderWriter) AddMessages(messages ...*MessageWriter) {
	folderWriter.MessageWriter.Add(messages...)
	folderWriter.UpdateTableContext(messages)
}

// UpdateTableContext updates the TableContext of the folder to reference the message identifiers.
func (folderWriter *FolderWriter) UpdateTableContext(messages ...proto.Message) {

}

// StartFolderWriteChannel listens for sub-folders to write.
// The called is responsible for starting the write channel.
func (folderWriter *FolderWriter) StartFolderWriteChannel(writeGroup *errgroup.Group) {
	for receivedFolder := range folderWriter.FolderWriteChannel {
		// Listen for folders to write.
		writeGroup.Go(func() error {
			// Add folder to TableContextWriter write queue.
			folderWriter.TableContextWriter.AddFolders(receivedFolder)

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

	// Make this written folder findable in the B-Tree.
	if err := folderWriter.UpdateIdentifier(); err != nil {
		return 0, eris.Wrap(err, "failed to update identifier")
	}

	return tableContextWrittenSize + messagesWrittenSize, nil
}
