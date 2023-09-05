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
	"github.com/mooijtech/go-pst/v6/pkg"
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"google.golang.org/protobuf/proto"
	"os"
	"testing"
)

// TestWritePSTFile writes a new PST file.
func TestWritePSTFile(t *testing.T) {
	// Output file.
	outputFile, err := os.Create("1337.pst")

	if err != nil {
		t.Fatalf("Failed to create output file: %+v", err)
	}

	// TODO - Unsupported Unicode4k, note that go-pst does not write OST files (I also don't have a test OST).
	formatType := pst.FormatTypeUnicode
	encryptionType := pst.EncryptionTypePermute

	// Define writer.
	writeContext, writeCancel := context.WithCancel(context.Background())
	writeOptions := NewWriteOptions(formatType, encryptionType)
	writer := NewWriter(writeOptions)

	// Create messages to write.
	var messageProperties []*properties.Message

	messageProperties = append(messageProperties, &properties.Message{
		Subject: proto.String("Hello world!"),
		From:    proto.String("info@mooijtech.com"),
		Body:    proto.String("go-pst now supports writing PST files."),
	})

	// Attachment properties.
	var attachmentProperties []*properties.Attachment

	attachmentProperties = append(attachmentProperties, &properties.Attachment{
		AttachFilename:     proto.String("nudes.png"),
		AttachLongFilename: proto.String("nudes.png"),
	})

	// Message attachments.
	messageAttachments := []*AttachmentWriter{NewAttachmentWriter(attachmentProperties)}

	// Message
	message := NewMessageWriter(messageProperties, messageAttachments)

	// Folder
	rootFolder, err := NewFolderWriter(outputFile, context.Background(), -1, formatType)

	if err != nil {
		t.Fatalf("Failed to create folder writer: %+v", err)
	}

	// Writer
	writer.AddFolder(rootFolder)

	if _, err := writer.WriteTo(outputFile); err != nil {
		t.Fatalf("Failed to write PST file: %+v", err)
	}
}

func TestWriteTableContext(t *testing.T) {
	attachmentProperties := properties.Attachment{
		AttachLongFilename: proto.String("nudes.png"),
	}

	tableContextWriter := NewTableContextWriter(pst.FormatTypeUnicode, &attachmentProperties)

	if _, err := tableContextWriter.WriteTo(os.Stdout); err != nil {
		t.Fatalf("Failed to write Table Context: %+v", err)
	}
}
