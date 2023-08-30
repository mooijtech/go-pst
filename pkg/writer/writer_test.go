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
	"github.com/mooijtech/go-pst/v6/pkg"
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"google.golang.org/protobuf/proto"
	"os"
	"testing"
)

// TestWritePSTFile writes a new PST file.
func TestWritePSTFile(t *testing.T) {
	outputFile, err := os.Create("1337.pst")

	if err != nil {
		t.Fatalf("Failed to create output file: %+v", err)
	}

	writeOptions := NewWriteOptions(pst.FormatTypeUnicode, pst.EncryptionTypePermute)
	writer := NewWriter(writeOptions)

	// Create messages to write.
	messageProperties := &properties.Message{
		Subject: proto.String("Hello world!"),
		From:    proto.String("info@mooijtech.com"),
		Body:    proto.String("go-pst now supports writing PST files."),
	}
	attachmentProperties := &properties.Attachment{
		AttachFilename:     proto.String("nudes.png"),
		AttachLongFilename: proto.String("nudes.png"),
	}
	messageAttachments := []*AttachmentWriter{NewAttachmentWriter(attachmentProperties)}

	message := NewMessageWriter(messageProperties, messageAttachments)

	folderProperties := NewFolderProperties("root")
	rootFolder := NewFolderWriter(folderProperties, []*MessageWriter{message})

	writer.AddFolder(rootFolder)

	if _, err := writer.WriteTo(outputFile); err != nil {
		t.Fatalf("Failed to write PST file: %+v", err)
	}
}

// TestWritePSTFileFromExisting writes a PST file from an existing PST file.
func TestWritePSTFileFromExisting(t *testing.T) {

}

// TestReadWrittenPSTFile tests if go-pst can read the written PST file without errors.
func TestReadWrittenPSTFile(t *testing.T) {
	TestWritePSTFile(t)

	// TODO - Read
}
