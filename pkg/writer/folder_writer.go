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

// FolderWriter represents a writer for folders.
type FolderWriter struct {
	// Properties represents the FolderProperties.
	Properties *properties.Folder
	// Messages represents the messages in this folder.
	Messages []*MessageWriter
	// TableContextWriter writes the pst.TableContext of the pst.Folder.
	TableContextWriter *TableContextWriter
}

// NewFolderWriter creates a new FolderWriter.
func NewFolderWriter(folderProperties *properties.Folder, messages []*MessageWriter) *FolderWriter {
	return &FolderWriter{
		Properties:         folderProperties,
		Messages:           messages,
		TableContextWriter: NewTableContextWriter(folderProperties),
	}
}

// WriteTo writes the folder containing messages.
// Returns the amount of bytes written to the output buffer.
// References TODO
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
