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

package pst

import (
	_ "embed"
	"fmt"
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	"github.com/tinylib/msgp/msgp"
)

// Message represents a message.
type Message struct {
	Identifier             Identifier
	PropertyContext        *PropertyContext
	AttachmentTableContext *TableContext
	LocalDescriptors       []LocalDescriptor // Used by the PropertyContext and TableContext.
	Properties             msgp.Decodable    // Type properties.Message, properties.Appointment, properties.Contact

	// TODO - Remove pointer so this can be used by the writer separately?
	File *File
}

// NewMessage constructs a new Message.
func NewMessage(file *File, identifier Identifier, localDescriptors []LocalDescriptor, propertyContext *PropertyContext) (*Message, error) {
	var messageProperties msgp.Decodable

	messageClassPropertyReader, err := propertyContext.GetPropertyReader(26, localDescriptors)

	if err != nil {
		fmt.Printf("Failed to get message class property reader, falling back to properties.Message: %+v\n", eris.New(err.Error()))
		messageProperties = &properties.Message{}
	} else {
		messageClass, err := messageClassPropertyReader.GetString()

		if err != nil {
			fmt.Printf("Failed to get message class, falling back to properties.Message: %+v\n", eris.New(err.Error()))
			messageProperties = &properties.Message{}
		} else {
			// https://learn.microsoft.com/en-us/office/vba/outlook/concepts/forms/item-types-and-message-classes
			if messageClass == "IPM.Note" || messageClass == "IPM.Note.SMIME.MultipartSigned" {
				messageProperties = &properties.Message{}
			} else if messageClass == "IPM.Appointment" || messageClass == "IPM.Schedule.Meeting" || messageClass == "IPM.Schedule.Meeting.Request" || messageClass == "IPM.OLE.CLASS.{00061055-0000-0000-C000-000000000046}" {
				messageProperties = &properties.Appointment{}
			} else if messageClass == "IPM.Contact" || messageClass == "IPM.AbchPerson" {
				messageProperties = &properties.Contact{}
			} else if messageClass == "IPM.Task" {
				messageProperties = &properties.Task{}
			} else if messageClass == "IPM.Activity" {
				messageProperties = &properties.Journal{}
			} else if messageClass == "IPM.Post.Rss" {
				messageProperties = &properties.RSS{}
			} else if messageClass == "IPM.DistList" {
				messageProperties = &properties.AddressBook{}
			} else {
				fmt.Printf("Unmapped message class \"%s\", falling back to properties.Message...\n", messageClass)
				messageProperties = &properties.Message{}
			}
		}
	}

	if err := propertyContext.Populate(messageProperties, localDescriptors); err != nil {
		return nil, eris.Wrap(err, "failed to populate message properties")
	}

	return &Message{
		File:             file,
		Identifier:       identifier,
		PropertyContext:  propertyContext,
		LocalDescriptors: localDescriptors,
		Properties:       messageProperties,
	}, nil
}

// GetMessageTableContext returns the message table context of this folder which contains references to all messages.
// Note this only returns the identifier of each message.
func (folder *Folder) GetMessageTableContext(file *File) (TableContext, error) {
	emailsIdentifier := folder.Identifier + 12

	emailsNode, err := folder.File.GetNodeBTreeNode(emailsIdentifier)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to find node b-tree node")
	}

	localDescriptors, err := folder.File.GetLocalDescriptors(emailsNode)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to find local descriptors")
	}

	emailsDataNode, err := folder.File.GetDataBTreeNode(emailsIdentifier)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to find data b-tree node")
	}

	emailsHeapOnNode, err := folder.File.GetHeapOnNode(emailsDataNode)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get Heap-on-Node")
	}

	// 26610 is a message property HNID.
	tableContext, err := file.GetTableContext(emailsHeapOnNode, localDescriptors, 26610)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get table context")
	}

	return tableContext, nil
}

// MessageIterator implements a message iterator.
type MessageIterator struct {
	file                *File
	messageTableContext TableContext

	err            error
	currentIndex   int
	currentMessage *Message
}

// Err return the error cause.
func (messageIterator *MessageIterator) Err() error {
	return messageIterator.err
}

// Next will ensure that Value returns the next item when executed.
// If the next value is not retrievable, Next will return false and Err() will return the error cause.
func (messageIterator *MessageIterator) Next() bool {
	hasNext := len(messageIterator.messageTableContext.Properties) > messageIterator.currentIndex

	if !hasNext {
		return false
	}

	var currentMessage *Message

	for _, property := range messageIterator.messageTableContext.Properties[messageIterator.currentIndex] {
		// We only return the message identifier in GetMessageTableContext,
		// so we don't need to check the property ID here.
		propertyReader, err := messageIterator.messageTableContext.GetPropertyReader(property)

		if err != nil {
			messageIterator.err = eris.Wrap(err, "failed to get property reader")
			return false
		}

		messageIdentifier, err := propertyReader.GetInteger32()

		if err != nil {
			messageIterator.err = eris.Wrap(err, "failed to get message identifier")
			return false
		}

		message, err := messageIterator.file.GetMessage(Identifier(messageIdentifier))

		if err != nil {
			messageIterator.err = eris.Wrapf(err, "failed to find message: %d", messageIdentifier)
			return false
		}

		currentMessage = message
	}

	messageIterator.currentIndex++
	messageIterator.currentMessage = currentMessage

	return true
}

// Value returns the current value in the iterator.
func (messageIterator *MessageIterator) Value() *Message {
	return messageIterator.currentMessage
}

// Size returns the amount of messages in the message iterator.
func (messageIterator *MessageIterator) Size() int {
	return len(messageIterator.messageTableContext.Properties)
}

func (messageIterator *MessageIterator) CurrentIndex() int {
	return messageIterator.currentIndex
}

// GetMessageIterator returns an iterator for messages.
func (folder *Folder) GetMessageIterator() (MessageIterator, error) {
	if folder.MessageCount == 0 {
		return MessageIterator{}, ErrMessagesNotFound
	} else if folder.Identifier.GetType() == IdentifierTypeSearchFolder {
		return MessageIterator{}, ErrMessagesNotFound
	}

	messageTableContext, err := folder.GetMessageTableContext()

	if err != nil {
		return MessageIterator{}, eris.Wrap(err, "failed to get message table context")
	}

	return MessageIterator{
		file:                folder.File,
		messageTableContext: messageTableContext,
	}, nil
}

// GetAllMessages returns an array of all messages from the message table context.
// See GetMessageIterator.
func (folder *Folder) GetAllMessages() ([]*Message, error) {
	messageIterator, err := folder.GetMessageIterator()

	if err != nil {
		return nil, err
	}

	var messages []*Message

	for messageIterator.Next() {
		messages = append(messages, messageIterator.Value())
	}

	return messages, messageIterator.Err()
}

// GetMessage returns the message of the identifier.
func (file *File) GetMessage(identifier Identifier) (*Message, error) {
	if identifier.GetType() != IdentifierTypeNormalMessage {
		return nil, ErrMessageIdentifierTypeInvalid
	}

	messageNode, err := file.GetNodeBTreeNode(identifier)

	if err != nil {
		return nil, eris.Wrap(err, "failed to find node b-tree node")
	}

	messageDataNode, err := file.GetBlockBTreeNode(messageNode.DataIdentifier)

	if err != nil {
		return nil, eris.Wrap(err, "failed to find block b-tree node")
	}

	messageHeapOnNode, err := file.GetHeapOnNode(messageDataNode)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get Heap-on-Node")
	}

	localDescriptors, err := file.GetLocalDescriptors(messageNode)

	if err != nil {
		return nil, eris.Wrap(err, "failed to find local descriptors")
	}

	propertyContext, err := file.GetPropertyContext(messageHeapOnNode)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get property context")
	}

	return NewMessage(file, identifier, localDescriptors, propertyContext)
}

// GetBodyRTF return the RTF body, may be
func (message *Message) GetBodyRTF() (string, error) {
	rtfPropertyReader, err := message.PropertyContext.GetPropertyReader(4105, message.LocalDescriptors)

	if err != nil {
		return "", err
	}

	rtfBody := make([]byte, rtfPropertyReader.Size())

	if _, err := rtfPropertyReader.ReadAt(rtfBody, 0); err != nil {
		return "", errors.WithStack(err)
	}

	return NewRTFDecoder().Decode(rtfBody)
}
