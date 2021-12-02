// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
	"fmt"
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

// GetMessageString returns the string value of the property.
func (pstFile *File) GetMessageString(message Message, propertyID int, formatType string, encryptionType string) (string, error) {
	propertyContextItem, err := FindPropertyContextItem(message.PropertyContext, propertyID)

	if err != nil {
		return "", err
	}

	if !propertyContextItem.IsExternalValueReference {
		if propertyID == 4096 || propertyID == 4115 { // Only the message body uses the specified encoding as far as I know.
			return DecodeMessageBytesToString(message, propertyContextItem.Data)
		} else {
			return DecodeBytesToUTF16String(propertyContextItem.Data)
		}
	} else {
		// External value reference (data is stored in a separate node).
		propertyLocalDescriptor, err := FindLocalDescriptor(message.LocalDescriptors, propertyContextItem.ReferenceHNID, formatType)

		if err != nil {
			return "", err
		}

		propertyHeapOnNode, err := pstFile.NewHeapOnNodeFromLocalDescriptor(propertyLocalDescriptor, formatType, encryptionType)

		if err != nil {
			return "", err
		}

		data, err := propertyHeapOnNode.InputStream.Read(propertyHeapOnNode.InputStream.Size, 0)

		if err != nil {
			return "", err
		}

		return DecodeMessageBytesToString(message, data)
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

// GetMessageDate returns the date value of the property.
func (pstFile *File) GetMessageDate(message Message, propertyID int) (time.Time, error) {
	propertyContextItem, err := FindPropertyContextItem(message.PropertyContext, propertyID)

	if err != nil {
		return time.Time{}, err
	}

	// References https://stackoverflow.com/a/57903746
	dateInteger := binary.LittleEndian.Uint64(propertyContextItem.Data)

	t := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	d := time.Duration(dateInteger)

	for i := 0; i < 100; i++ {
		t = t.Add(d)
	}

	return t, nil
}

// GetSubject returns the subject of this message.
func (pstFile *File) GetMessageSubject(message Message, formatType string, encryptionType string) (string, error) {
	return pstFile.GetMessageString(message, 55, formatType, encryptionType)
}

// GetMessageClass returns the message class.
func (pstFile *File) GetMessageClass(message Message, formatType string, encryptionType string) (string, error) {
	return pstFile.GetMessageString(message, 26, formatType, encryptionType)
}

// GetMessageID returns the message ID.
func (pstFile *File) GetMessageID(message Message, formatType string, encryptionType string) (string, error) {
	return pstFile.GetMessageString(message, 4149, formatType, encryptionType)
}

// GetMessageHeaders return the message headers.
func (pstFile *File) GetMessageHeaders(message Message, formatType string, encryptionType string) (string, error) {
	return pstFile.GetMessageString(message, 125, formatType, encryptionType)
}

// GetFrom returns the "From" header.
func (pstFile *File) GetMessageFrom(message Message, formatType string, encryptionType string) (string, error) {
	return pstFile.GetMessageString(message, 3103, formatType, encryptionType)
}

// GetMessageTo returns the "To" header.
func (pstFile *File) GetMessageTo(message Message, formatType string, encryptionType string) (string, error) {
	return pstFile.GetMessageString(message, 3588, formatType, encryptionType)
}

// GetMessageCC returns the "CC" header.
func (pstFile *File) GetMessageCC(message Message, formatType string, encryptionType string) (string, error) {
	return pstFile.GetMessageString(message, 3587, formatType, encryptionType)
}

// GetMessageBCC returns the BCC of this message.
func (pstFile *File) GetMessageBCC(message Message, formatType string, encryptionType string) (string, error) {
	originalDisplayBCC, err := pstFile.GetMessageString(message, 114, formatType, encryptionType)

	if err == nil && originalDisplayBCC != "" {
		return originalDisplayBCC, nil
	}

	return pstFile.GetMessageString(message, 3586, formatType, encryptionType)
}

// GetMessageReceivedDate returns the date this message was received.
func (pstFile *File) GetMessageReceivedDate(message Message) (time.Time, error) {
	return pstFile.GetMessageDate(message, 3590)
}

// GetMessageBody returns the plaintext body of the message.
func (pstFile *File) GetMessageBody(message Message, formatType string, encryptionType string) (string, error) {
	return pstFile.GetMessageString(message, 4096, formatType, encryptionType)
}

// GetBodyHTML returns the HTML body of this message.
func (pstFile *File) GetBodyHTML(message Message, formatType string, encryptionType string) (string, error) {
	return pstFile.GetMessageString(message, 4115, formatType, encryptionType)
}