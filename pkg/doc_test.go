// go-pst is a library for reading Personal Storage Table (.pst) files (written in Go/Golang).
//
// Copyright (C) 2022  Marten Mooij
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package pst_test

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"testing"
	"time"

	pst "github.com/mooijtech/go-pst/v5/pkg"
)

func TestExample(t *testing.T) {
	startTime := time.Now()

	fmt.Println("Initializing...")

	pstFile, err := pst.NewFromFile("../data/enron.pst")

	if err != nil {
		panic(fmt.Sprintf("Failed to open PST file: %+v\n", err))
	}

	defer func() {
		if errClosing := pstFile.Close(); errClosing != nil {
			panic(fmt.Sprintf("Failed to close PST file: %+v\n", err))
		}
	}()

	// Walk through folders.
	if err := pstFile.WalkFolders(func(folder pst.Folder) error {
		fmt.Printf("Walking folder: %s\n", folder.Name)

		messageIterator, err := folder.GetMessageIterator()

		if errors.Is(err, pst.ErrMessagesNotFound) {
			// Folder has no messages.
			return nil
		} else if err != nil {
			return err
		}

		// Iterate through messages.
		for messageIterator.Next() {
			message := messageIterator.Value()

			fmt.Printf("Message subject: %s\n", message.GetSubject())

			attachmentIterator, err := message.GetAttachmentIterator()

			if errors.Is(err, pst.ErrAttachmentsNotFound) {
				// This message has no attachments.
				continue
			} else if err != nil {
				return err
			}

			// Iterate through attachments.
			for attachmentIterator.Next() {
				attachment := attachmentIterator.Value()

				fmt.Printf("Attachment: %s\n", attachment.GetAttachFilename())

				attachmentOutput, err := os.Create(fmt.Sprintf("attachments/%s", attachment.GetAttachFilename()))

				if err != nil {
					return err
				} else if _, err := attachment.WriteTo(attachmentOutput); err != nil {
					return err
				}

				if err := attachmentOutput.Close(); err != nil {
					return err
				}
			}

			if attachmentIterator.Err() != nil {
				return attachmentIterator.Err()
			}
		}

		return messageIterator.Err()
	}); err != nil {
		panic(fmt.Sprintf("Failed to walk folders: %+v\n", err))
	}

	fmt.Printf("Time: %s\n", time.Since(startTime).String())
}
