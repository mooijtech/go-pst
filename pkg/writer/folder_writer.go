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
	"context"
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
	"io"
	"sync/atomic"
)

// FolderWriter represents a writer for folders.
type FolderWriter struct {
	// FormatType represents the FormatType used while writing.
	FormatType pst.FormatType
	// FolderWriteChannel represents the Go channel for writing sub-folders.
	FolderWriteChannel chan *FolderWriter
	// MessageWriteChannel represents the Go channel for writing messages to this folder.
	MessageWriteChannel chan *MessageWriter
	// TableContextWriter writes the pst.TableContext of the pst.Folder.
	TableContextWriter *TableContextWriter
	// PropertyWriteChannel represents the Go channel for writing []pst.Property.
	PropertyWriteChannel chan *PropertyWriter
	// TotalSize represents the total size of the PST file so far.
	TotalSize atomic.Int64
}

// NewFolderWriter creates a new FolderWriter.
func NewFolderWriter(writer io.Writer, writeContext context.Context, writeLimit int, writeGroup *errgroup.Group, formatType pst.FormatType) (*FolderWriter, error) {
	folderWriter := &FolderWriter{
		FormatType:           formatType,
		FolderWriteChannel:   make(chan *FolderWriter),
		MessageWriteChannel:  make(chan *MessageWriter),
		TableContextWriter:   NewTableContextWriter(writer, writeContext, writeLimit, formatType),
		PropertyWriteChannel: make(chan *PropertyWriter),
	}

	// Start channels.
	folderWriteChannelErrGroup, folderWriteChannelContext := folderWriter.StartFolderWriteChannel(writeContext)
	messageWriteChannelErrGroup, messageWriteChannelContext := folderWriter.StartMessageWriteChannel(writeContext)

	return folderWriter, subFolderChannel
}

// SetProperties writes the properties of this folder.
func (folderWriter *FolderWriter) SetProperties(properties *properties.Folder) {

}

// AddMessages adds a message to the channel to be written.
func (folderWriter *FolderWriter) AddMessages(messageWriters ...*MessageWriter) {
	for _, messageWriter := range messageWriters {
		folderWriter.MessageWriteChannel <- messageWriter
	}
}

func (folderWriter *FolderWriter) StartMessageWriteChannel(writeContext context.Context) (*errgroup.Group, context.Context) {
	messageWriteChannelErrGroup, messageWriteChannelContext := errgroup.WithContext(writeContext)

	messageWriteChannelErrGroup.Go(func() error {
		return nil
	})

	return messageWriteChannelErrGroup, messageWriteChannelContext
}

// AddFolder queues the folder to be written, picked up by a Go channel.
// This is added to the TableContextWriter, so we can find these folders.
func (folderWriter *FolderWriter) AddFolder(folders ...*FolderWriter) {
	for _, folder := range folders {
		folderWriter.FolderWriteChannel <- folder
	}
}

// StartFolderWriteChannel listens for sub-folders to write.
// The called is responsible for starting the write channel.
func (folderWriter *FolderWriter) StartFolderWriteChannel(folderChannelContext context.Context) (*errgroup.Group, context.Context) {
	folderWriteChannelErrGroup, folderWriteChannelContext := errgroup.WithContext(folderChannelContext)

	folderWriteChannelErrGroup.Go(func() error {
		for receivedFolder := range folderWriter.FolderWriteChannel {
			writtenSize, err := folderWriter.TableContextWriter.AddFolder(receivedFolder)

			if err != nil {
				return eris.Wrap(err, "failed to add folder to Table Context")
			}

			// Callback
			folderWriter
		}

		return nil
	})

	return folderWriteChannelErrGroup, folderWriteChannelContext
}

// WriteTo writes the folder containing messages.
// Returns the amount of bytes written to the output buffer.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#folders
func (folderWriter *FolderWriter) WriteTo(writer io.Writer) (int64, error) {
	tableContextWrittenSize, err := folderWriter.TableContextWriter.WriteTo(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write Table Context")
	}

	messagesWrittenSize, err := folderWriter.WriteMessages(writer)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write messages")
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
