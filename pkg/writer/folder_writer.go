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

import "github.com/rotisserie/eris"

// FolderWriter represents a writer for folders.
type FolderWriter struct {
	// Properties represents the FolderProperties.
	Properties FolderProperties
	// Messages represents the messages in this folder.
	Messages []*MessageWriter
	// TableContextWriter writes the pst.TableContext of the pst.Folder.
	TableContextWriter *TableContextWriter
}

// FolderProperties represents the properties of a pst.Folder.
// TODO - Move to properties.Folder.
type FolderProperties struct {
	Name string
}

// NewFolderProperties creates a new FolderProperties.
func NewFolderProperties(name string) FolderProperties {
	return FolderProperties{
		Name: name,
	}
}

// NewFolderWriter creates a new FolderWriter.
func NewFolderWriter(folderProperties FolderProperties, messages []*MessageWriter, tableContextWriter *TableContextWriter) *FolderWriter {
	return &FolderWriter{
		Properties:         folderProperties,
		Messages:           messages,
		TableContextWriter: tableContextWriter,
	}
}

// Write writes the folder containing messages.
func (folderWriter *FolderWriter) Write() error {
	if err := folderWriter.TableContextWriter.Write(); err != nil {
		return eris.Wrap(err, "failed to write Table Context")
	}

	if err := folderWriter.WriteMessages(); err != nil {
		return eris.Wrap(err, "failed to write messages")
	}

	return nil
}

// WriteFolders writes the pst.TableContext of the folders.
func (folderWriter *FolderWriter) WriteFolders() error {
	return nil
}

// WriteMessages writes the messages of the folder.
func (folderWriter *FolderWriter) WriteMessages() error {
	for _, messageWriter := range folderWriter.Messages {
		if err := messageWriter.Write(); err != nil {
			return eris.Wrap(err, "failed to write message")
		}
	}

	return nil
}
