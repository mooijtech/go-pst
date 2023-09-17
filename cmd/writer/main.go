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
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/mooijtech/concurrent-writer/concurrent"
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Command-line flags.
	outputName := *flag.String("output", "1337.pst", "output path of the PST file")

	flag.Parse()

	slog.Info("Starting go-pst writer...")

	startTime := time.Now()

	// Write PST file.
	written, err := WritePSTFile(outputName)

	if err != nil {
		panic(fmt.Sprintf("Failed to write PST file: %+v", err))
	}

	fmt.Printf("Wrote %d bytes in %s.", written, time.Since(startTime).String())
}

// WritePSTFile writes a PST file containing folders, messages and attachments.
func WritePSTFile(outputName string) (int64, error) {
	// Output PST file.
	outputFile, err := os.Create(outputName)

	if err != nil {
		return 0, eris.Wrap(err, "failed to create output file")
	}

	// 4KiB is the default I/O buffer size on Linux.
	// Ideally all writes should be aligned on this boundary (FormatTypeUnicode4k).
	// We could also add support for Linux I/O URing (https://en.wikipedia.org/wiki/Io_uring).
	// You can use your own io.WriteSeeker.
	// At the end of the day the bottleneck is still being limited by disk I/O.
	concurrentWriter := concurrent.NewWriterAutoFlush(outputFile, 4096, 0.75)

	// Write options.
	formatType := pst.FormatTypeUnicode4k
	encryptionType := pst.EncryptionTypePermute
	options := pst.NewOptions(formatType, encryptionType)

	// Write group for Goroutines (all writers run here).
	writeCancelContext, writeCancelFunc := context.WithCancel(context.Background())
	writeGroup, _ := errgroup.WithContext(writeCancelContext)

	defer writeCancelFunc()

	// Writer.
	writer, err := pst.NewWriter(concurrentWriter, writeGroup, options)

	if err != nil {
		return 0, eris.Wrap(err, "failed to create writer")
	}

	// Let's write some folders with messages containing attachments.
	rootFolder, err := writer.GetRootFolder()

	if err != nil {
		return 0, eris.Wrap(err, "failed to get root folder writer")
	}

	// Callbacks are used to calculate the total size of the PST file.
	// As soon as this becomes larger than 50GB, we overflow by creating a new PST file.
	folderWriteCallback := make(chan int64)

	// Add sub-folders to the root folder.
	for i := 0; i < 3; i++ {
		// Create folder writer.
		folderWriter, err := pst.NewFolderWriter(concurrentWriter, writeGroup, formatType, folderWriteCallback, rootFolder.GetIdentifier())

		if err != nil {
			return 0, eris.Wrap(err, "failed to create folder writer")
		}

		// Add properties to the folder.
		folderWriter.AddProperties(&properties.Folder{
			Name: fmt.Sprintf("Sub folder #%d", i),
			// TODO - Extend from generate.go new output.
		})

		// Add messages to the folder.
		for i := 0; i < 6; i++ {
			// Create a message to write.
			message, err := pst.NewMessageWriter(concurrentWriter, writeGroup, folderWriter.GetIdentifier(), formatType)

			if err != nil {
				return 0, eris.Wrap(err, "failed to create message writer")
			}

			// Add properties to the message.
			message.AddProperties(&properties.Message{
				Subject: proto.String("[Go Forensics]: Goodbye, world!"),
				From:    proto.String("info@mooijtech.com"),
				Body:    proto.String("https://goforensics.io/"),
				// See all other available properties.
			})

			// You can create many property types, for example, a contact:
			message.AddProperties(&properties.Contact{
				GivenName: proto.String("Marten Mooij"),
			})

			// See the properties package for more message types.

			// Add attachments to the message.
			for i := 0; i < 9; i++ {
				attachmentWriter, err := pst.NewAttachmentWriter(concurrentWriter, writeGroup, formatType)

				if err != nil {
					return 0, eris.Wrap(err, "failed to create attachment writer")
				}

				// Set attachment input buffer.
				if err := attachmentWriter.AddFile("example-attachment.txt"); err != nil {
					return 0, eris.Wrap(err, "failed to add attachment to message")
				}

				message.AddAttachments(attachmentWriter)
			}

			// Add the message to the folder.
			folderWriter.AddMessages(message)
		}

		// Add sub-folder to the root folder.
		rootFolder.AddSubFolders(folderWriter)
	}

	// See the documentation of WriteTo.
	written, err := writer.WriteTo(outputFile)

	if errors.Is(err, pst.ErrOverflow) {
		// Handle edge case where the PST file is >= 50GB.
		// Redirect the channels to a new output file.
		randomBytes := make([]byte, 8)

		if _, err := rand.Read(randomBytes); err != nil {
			return 0, eris.Wrap(err, "failed to read random bytes from crypto/rand")
		}

		outputExtension := filepath.Ext(outputName)
		outputNameWithoutExtension := strings.ReplaceAll(outputName, outputExtension, "")
		newOutputName := fmt.Sprintf("%s-%s%s", outputNameWithoutExtension, hex.EncodeToString(randomBytes), outputExtension)
		newOutputFile, err := os.Create(newOutputName)

		if err != nil {
			return 0, eris.Wrap(err, "failed to create new output file for overflow")
		}

		// Redirect output.
		writer.OverflowTo(newOutputFile)
	} else if err != nil {
		return 0, eris.Wrap(err, "failed to write PST file")
	}

	return written, nil
}
