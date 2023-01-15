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
	"github.com/mooijtech/go-pst/v5/pkg/properties"
	"github.com/rotisserie/eris"
	"os"
	"testing"
	"time"

	pst "github.com/mooijtech/go-pst/v5/pkg"
)

func TestExample(t *testing.T) {
	startTime := time.Now()

	fmt.Println("Initializing...")

	reader, err := os.Open("/home/bot/Documents/Projects/go-pst/data/s.woon.pst")

	if err != nil {
		panic(fmt.Sprintf("Failed to open PST file: %+v\n", err))
	}

	pstFile, err := pst.New(reader)

	if err != nil {
		panic(fmt.Sprintf("Failed to open PST file: %+v\n", err))
	}

	defer func() {
		pstFile.Cleanup()

		if errClosing := reader.Close(); errClosing != nil {
			panic(fmt.Sprintf("Failed to close PST file: %+v\n", err))
		}
	}()

	// Walk through folders.
	if err := pstFile.WalkFolders(func(folder pst.Folder) error {
		fmt.Printf("Walking folder: %s\n", folder.Name)

		messageIterator, err := folder.GetMessageIterator()

		if eris.Is(err, pst.ErrMessagesNotFound) {
			// Folder has no messages.
			return nil
		} else if err != nil {
			return err
		}

		// Iterate through messages.
		for messageIterator.Next() {
			message := messageIterator.Value()

			switch messageProperties := message.Properties.(type) {
			case *properties.Appointment:
				fmt.Printf("Appointment: %s - %s\n", messageProperties.String(), folder.Name)
			case *properties.Contact:
				fmt.Printf("Contact: %s - %s\n", messageProperties.String(), folder.Name)
			case *properties.Task:
				fmt.Printf("Task: %s\n", messageProperties.GetTaskAssigner())
			case *properties.RSS:
				fmt.Printf("RSS: %s\n", messageProperties.GetPostRssChannelLink())
			case *properties.AddressBook:
				fmt.Printf("Address book: %s\n", messageProperties.GetAccount())
			case *properties.Message:
				//fmt.Printf("Message: %s\n", messageProperties.GetSubject())
			case *properties.Note:
				fmt.Printf("Note: %d\n", messageProperties.GetNoteColor())
			default:
				fmt.Printf("Unknown message type\n")
			}

			//fmt.Printf("Got: %s\n", message.GetOriginalMessageClass())

			//attachmentIterator, err := message.GetAttachmentIterator()
			//
			//if eris.Is(err, pst.ErrAttachmentsNotFound) {
			//	// This message has no attachments.
			//	continue
			//} else if err != nil {
			//	return err
			//}
			//
			//// Iterate through attachments.
			//for attachmentIterator.Next() {
			//	attachment := attachmentIterator.Value()
			//
			//	fmt.Printf("Attachment: %s\n", attachment.GetAttachFilename())
			//
			//	attachmentOutput, err := os.Create(fmt.Sprintf("attachments/%s", attachment.GetAttachFilename()))
			//
			//	if err != nil {
			//		return err
			//	} else if _, err := attachment.WriteTo(attachmentOutput); err != nil {
			//		return err
			//	}
			//
			//	if err := attachmentOutput.Close(); err != nil {
			//		return err
			//	}
			//}
			//
			//if attachmentIterator.Err() != nil {
			//	return attachmentIterator.Err()
			//}
		}

		return messageIterator.Err()
	}); err != nil {
		panic(fmt.Sprintf("Failed to walk folders: %+v\n", err))
	}

	fmt.Printf("Time: %s\n", time.Since(startTime).String())
}
