// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
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

	emailsNode, err := pstFile.GetNodeBTreeNode(emailsIdentifier)

	if err != nil {
		return nil, err
	}

	localDescriptors, err := pstFile.GetLocalDescriptors(emailsNode, formatType)

	if err != nil {
		return nil, err
	}

	emailsDataNode, err := pstFile.GetDataBTreeNode(emailsIdentifier)

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

	messageNode, err := pstFile.GetNodeBTreeNode(identifier)

	if err != nil {
		return Message{}, err
	}

	messageDataNode, err := pstFile.GetBlockBTreeNode(messageNode.DataIdentifier)

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
func (message *Message) GetString(propertyID int, pstFile *File, formatType string, encryptionType string) (string, error) {
	propertyContextItem, err := FindPropertyContextItem(message.PropertyContext, propertyID)

	if err != nil {
		return "", err
	}

	encoding, err := message.GetEncoding()

	if err != nil {
		return "", err
	}

	return propertyContextItem.GetString(encoding, message.LocalDescriptors, pstFile, formatType, encryptionType)
}

// GetInteger returns the integer value of the property.
func (message *Message) GetInteger(propertyID int) (int, error) {
	propertyContextItem, err := FindPropertyContextItem(message.PropertyContext, propertyID)

	if err != nil {
		return -1, err
	}

	return propertyContextItem.GetInteger(), nil
}

// GetDate returns the date value of the property.
func (message *Message) GetDate(propertyID int) (time.Time, error) {
	propertyContextItem, err := FindPropertyContextItem(message.PropertyContext, propertyID)

	if err != nil {
		return time.Time{}, err
	}

	return propertyContextItem.GetDate(), nil
}

// GetSubject returns the subject of this message.
func (message *Message) GetSubject(pstFile *File, formatType string, encryptionType string) (string, error) {
	return message.GetString(55, pstFile, formatType, encryptionType)
}

// GetMessageClass returns the message class.
func (message *Message) GetMessageClass(pstFile *File, formatType string, encryptionType string) (string, error) {
	return message.GetString(26, pstFile, formatType, encryptionType)
}

// GetMessageID returns the message ID.
func (message *Message) GetMessageID(pstFile *File, formatType string, encryptionType string) (string, error) {
	return message.GetString(4149, pstFile, formatType, encryptionType)
}

// GetHeaders return the message headers.
func (message *Message) GetHeaders(pstFile *File, formatType string, encryptionType string) (string, error) {
	return message.GetString(125, pstFile, formatType, encryptionType)
}

// GetFrom returns the "From" header.
func (message *Message) GetFrom(pstFile *File, formatType string, encryptionType string) (string, error) {
	return message.GetString(3103, pstFile, formatType, encryptionType)
}

// GetTo returns the "To" header.
func (message *Message) GetTo(pstFile *File, formatType string, encryptionType string) (string, error) {
	return message.GetString(3588, pstFile, formatType, encryptionType)
}

// GetCC returns the "CC" header.
func (message *Message) GetCC(pstFile *File, formatType string, encryptionType string) (string, error) {
	return message.GetString(3587, pstFile, formatType, encryptionType)
}

// GetBCC returns the BCC of this message.
func (message *Message) GetBCC(pstFile *File, formatType string, encryptionType string) (string, error) {
	originalDisplayBCC, err := message.GetString(114, pstFile, formatType, encryptionType)

	if err == nil && originalDisplayBCC != "" {
		return originalDisplayBCC, nil
	}

	return message.GetString(3586, pstFile, formatType, encryptionType)
}

// GetReceivedDate returns the date this message was received.
func (message *Message) GetReceivedDate() (time.Time, error) {
	return message.GetDate(3590)
}

// GetBody returns the plaintext body of the message.
func (message *Message) GetBody(pstFile *File, formatType string, encryptionType string) (string, error) {
	return message.GetString(4096, pstFile, formatType, encryptionType)
}

// GetBodyHTML returns the HTML body of this message.
func (message *Message) GetBodyHTML(pstFile *File, formatType string, encryptionType string) (string, error) {
	return message.GetString(4115, pstFile, formatType, encryptionType)
}