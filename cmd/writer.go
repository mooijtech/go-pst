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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/mooijtech/concurrent-writer/concurrent"
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"os"
	"time"
)

func main() {
	// Command-line flags.
	outputName := *flag.String("output", "1337.pst", "output path of the PST file")

	flag.Parse()

	slog.Info("Starting go-pst writer...")

	startTime := time.Now()

	// Output PST file.
	outputFile, err := os.Create(outputName)

	if err != nil {
		panic(fmt.Sprintf("Failed to create output file: %+v", err))
	}

	// 4KiB is the default I/O buffer size on Linux.
	// Ideally all writes should be aligned on this boundary (FormatTypeUnicode4k).
	// We could also add support for Linux I/O URing (https://en.wikipedia.org/wiki/Io_uring).
	// You can use your own io.WriteSeeker.
	concurrentWriter := concurrent.NewWriterAutoFlush(outputFile, 4096, 0.75)

	// Write options.
	formatType := pst.FormatTypeUnicode4k
	encryptionType := pst.EncryptionTypePermute
	writeOptions := pst.NewWriteOptions(formatType, encryptionType)

	// Write group for Goroutines.
	writeCancelContext, writeCancelFunc := context.WithCancel(context.Background())
	writeGroup, _ := errgroup.WithContext(writeCancelContext)

	defer writeCancelFunc()

	// Writer.
	writer, err := pst.NewWriter(concurrentWriter, writeGroup, writeOptions)

	if err != nil {
		panic(fmt.Sprintf("Failed to create writer: %+v", err))
	}

	// Write folders.

	rootFolder := writer.AddRootFolder()

	// Create sub-folders.
	writer.Add

	folderWriteCallback := make(chan int64)
	folderWriter, err := pst.NewFolderWriter(concurrentWriter, writeGroup, formatType, folderWriteCallback)

	if err != nil {
		panic(fmt.Sprintf("Failed to create FolderWriter: %+v", err))
	}

	rootFolder := writer.AddRootFolder()

	// Add sub-folders.
	for i := 0; i < 6; i++ {
		subFolder, err := pst.NewFolderWriter(outputFile, writeGroup, formatType)

		if err != nil {
			panic(fmt.Sprintf("Failed to create FolderWriter: %+v", err))
		}

		subFolderProperties := &properties.Folder{
			Name: fmt.Sprintf("Sub-folder #%d", i),
		}

		if err := subFolder.Add(subFolderProperties); err != nil {
			panic(fmt.Sprintf("Failed to add folder: %+v", err))
		}

		// Add messages to the sub-folder.
		message := NewMessageWriter(formatType, writeGroup)

		message.AddProperties(&properties.Message{
			Subject: proto.String("Goodbye, world!"),
		})

		// Add attachments to message.
		for x := 0; x < 9; x++ {
			attachment := NewAttachmentWriter()

			attachment.AddProperties(&properties.Attachment{
				AttachFilename:     proto.String(fmt.Sprintf("nudes%d.png", x)),
				AttachLongFilename: proto.String(fmt.Sprintf("nudes%d.png", x)),
			})

			message.AddAttachments(attachment)
		}

		// Add sub-folders with messages containing attachments to root folder.
		rootFolder.(subFolder)
	}

	// Writer which starts everything (has the PST file header).
	writer.AddFolders(rootFolder)

	// See the documentation of WriteTo for the path followed.
	bytesWritten, err := writer.WriteTo(outputFile)

	if err != nil {
		panic(fmt.Sprintf("Failed to write PST file: %+v", err))
	}

	// Wait for writers to finish.
	if err := writeGroup.Wait(); err != nil {
		panic(fmt.Sprintf("Failed to write PST file: %+v", err))
	}

	// humanize doesn't currently support Duration.
	fmt.Printf("Done! Wrote %s in %s.", humanize.Bytes(uint64(bytesWritten)), humanize.Time(time.Now().Add(-time.Since(startTime))))
}
