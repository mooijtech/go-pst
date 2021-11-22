// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/unicode"
	"time"
)

// Message represents a message.
type Message struct {
	PropertyContext         []PropertyContextItem
	LocalDescriptors        []LocalDescriptor
	AttachmentsTableContext [][]TableContextItem
}

// GetMessageTableContext returns the message table context of this folder.
func (pstFile *File) GetMessageTableContext(folder Folder, formatType string, encryptionType string) ([][]TableContextItem, error) {
	emailsIdentifier := folder.Identifier + 12

	emailsNode, err := pstFile.GetNodeBTreeNode(emailsIdentifier, formatType)

	if err != nil {
		return nil, err
	}

	localDescriptors, err := pstFile.GetLocalDescriptors(emailsNode, formatType)

	if err != nil {
		return nil, err
	}

	emailsDataNode, err := pstFile.GetDataBTreeNode(emailsIdentifier, formatType)

	if err != nil {
		return nil, err
	}

	emailsHeapOnNode, err := pstFile.NewHeapOnNodeFromNode(emailsDataNode, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	tableContext, err := pstFile.GetTableContext(emailsHeapOnNode, localDescriptors, formatType, encryptionType, -1, -1, 26610)

	if err != nil {
		return nil, err
	}

	return tableContext, nil
}

// GetMessages returns an array of messages from the message table context.
func (pstFile *File) GetMessages(folder Folder, formatType string, encryptionType string) ([]Message, error) {
	if folder.MessageCount == 0 {
		return nil, nil
	}

	identifierType := folder.Identifier & 0x1F

	if identifierType == IdentifierTypeSearchFolder {
		return nil, nil
	}

	messageTableContext, err := pstFile.GetMessageTableContext(folder, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	var messages []Message

	for _, messageTableContextRow := range messageTableContext {
		for _, messageTableContextColumn := range messageTableContextRow {
			if messageTableContextColumn.PropertyID == 26610 {
				message, err := pstFile.GetMessage(messageTableContextColumn.ReferenceHNID, formatType, encryptionType)

				if err != nil {
					// There may be other messages.
					fmt.Printf("Failed to get message (%d): %s\n", messageTableContextColumn.ReferenceHNID, err)
					continue
				}

				messages = append(messages, message)
			}
		}
	}

	return messages, nil
}

// GetMessage returns the message of the identifier.
func (pstFile *File) GetMessage(identifier int, formatType string, encryptionType string) (Message, error) {
	identifierType := identifier & 0x1F

	if identifierType != IdentifierTypeNormalMessage {
		return Message{}, errors.New("invalid identifier type")
	}

	messageNode, err := pstFile.GetNodeBTreeNode(identifier, formatType)

	if err != nil {
		return Message{}, err
	}

	messageNodeDataIdentifier, err := messageNode.GetDataIdentifier(formatType)

	if err != nil {
		return Message{}, err
	}

	messageDataNode, err := pstFile.GetBlockBTreeNode(messageNodeDataIdentifier, formatType)

	if err != nil {
		return Message{}, err
	}

	messageHeapOnNode, err := pstFile.NewHeapOnNodeFromNode(messageDataNode, formatType, encryptionType)

	if err != nil {
		return Message{}, err
	}

	localDescriptors, err := pstFile.GetLocalDescriptors(messageNode, formatType)

	if err != nil {
		return Message{}, err
	}

	propertyContext, err := pstFile.GetPropertyContext(messageHeapOnNode, formatType, encryptionType)

	if err != nil {
		return Message{}, err
	}

	message := Message{
		PropertyContext:  propertyContext,
		LocalDescriptors: localDescriptors,
	}

	return message, nil
}

// GetString returns the string value of the property.
func (message *Message) GetString(propertyID int) string {
	propertyContextItem, err := FindPropertyContextItem(message.PropertyContext, propertyID)

	if err != nil {
		return ""
	}

	if propertyID == 4096 || propertyID == 4115 { // Only the message body uses the specified encoding as far as I know.
		encoding, err := message.GetEncoding()

		if err != nil {
			return err.Error()
		}

		mimeEncoding, err := ianaindex.MIME.Encoding(encoding.Name)

		if err != nil {
			return err.Error()
		}

		inputReader, err := mimeEncoding.NewDecoder().Bytes(propertyContextItem.Data)

		if err != nil {
			return err.Error()
		}

		return string(inputReader)
	} else {
		// The libpff documentation states:
		// "Unicode strings are stored in UTF-16 little-endian without the byte order mark (BOM)."
		decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

		utf16String, err := decoder.String(string(propertyContextItem.Data))

		if err != nil {
			return err.Error()
		}

		return utf16String
	}
}

// GetInteger returns the integer value of the property or -1 if it was not found.
func (message *Message) GetInteger(propertyID int) int {
	propertyContextItem, err := FindPropertyContextItem(message.PropertyContext, propertyID)

	if err != nil {
		return -1
	}

	return propertyContextItem.ReferenceHNID
}

// GetDate returns the date value of the property.
// References https://stackoverflow.com/a/57903746
func (message *Message) GetDate(propertyID int) time.Time {
	propertyContextItem, err := FindPropertyContextItem(message.PropertyContext, propertyID)

	if err != nil {
		return time.Time{}
	}

	dateInteger := binary.LittleEndian.Uint64(propertyContextItem.Data)

	t := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	d := time.Duration(dateInteger)

	for i := 0; i < 100; i++ {
		t = t.Add(d)
	}

	return t
}

// GetSubject returns the subject of this message.
func (message *Message) GetSubject() string {
	return message.GetString(55)
}

// GetMessageClass returns the message class.
func (message *Message) GetMessageClass() string {
	return message.GetString(26)
}

// GetMessageID returns the message ID.
func (message *Message) GetMessageID() string {
	return message.GetString(4149)
}

// GetHeaders return the message headers.
func (message *Message) GetHeaders() string {
	return message.GetString(125)
}

// GetFrom returns the "From" header.
func (message *Message) GetFrom() string {
	return message.GetString(3103)
}

// GetTo returns the "To" header.
func (message *Message) GetTo() string {
	return message.GetString(3588)
}

// GetCC returns the "CC" header.
func (message *Message) GetCC() string {
	return message.GetString(3587)
}

// GetBCC returns the BCC of this message.
func (message *Message) GetBCC() string {
	originalDisplayBCC := message.GetString(114)

	if originalDisplayBCC != "" {
		return originalDisplayBCC
	}

	return message.GetString(3586)
}

// GetReceivedDate returns the date this message was received.
func (message *Message) GetReceivedDate() time.Time {
	return message.GetDate(3590)
}

// GetBody returns the plaintext body of the message.
func (message *Message) GetBody() string {
	return message.GetString(4096)
}

// GetBodyHTML returns the HTML body of this message.
func (message *Message) GetBodyHTML() string {
	return message.GetString(4115)
}