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

package pst

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"github.com/mooijtech/go-pst/v5/pkg/properties"
	"github.com/pkg/errors"
	"strings"
)

// Message represents a message.
type Message struct {
	PropertyContext        *PropertyContext
	AttachmentTableContext *TableContext
	LocalDescriptors       []LocalDescriptor // Used by the PropertyContext and TableContext.
	File                   *File
	properties.Message
}

// GetMessageTableContext returns the message table context of this folder which contains references to all messages.
// Note this only returns the identifier of each message.
func (folder *Folder) GetMessageTableContext() (TableContext, error) {
	emailsIdentifier := folder.Identifier + 12

	emailsNode, err := folder.File.GetNodeBTreeNode(emailsIdentifier)

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	localDescriptors, err := folder.File.GetLocalDescriptors(emailsNode)

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	emailsDataNode, err := folder.File.GetDataBTreeNode(emailsIdentifier)

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	emailsHeapOnNode, err := folder.File.GetHeapOnNode(emailsDataNode)

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	// 26610 is a message property HNID.
	tableContext, err := folder.File.GetTableContext(emailsHeapOnNode, localDescriptors, 26610)

	if err != nil {
		return TableContext{}, errors.WithStack(err)
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
			messageIterator.err = errors.WithStack(err)
			return false
		}

		messageIdentifier, err := propertyReader.GetInteger32()

		if err != nil {
			messageIterator.err = errors.WithStack(err)
			return false
		}

		message, err := messageIterator.file.GetMessage(Identifier(messageIdentifier))

		if err != nil {
			messageIterator.err = errors.WithStack(err)
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
		return MessageIterator{}, errors.WithStack(err)
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
		return nil, errors.WithStack(errors.WithMessage(ErrMessageIdentifierTypeInvalid, fmt.Sprintf("Identifier type: %d", identifier.GetType())))
	}

	messageNode, err := file.GetNodeBTreeNode(identifier)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	messageDataNode, err := file.GetBlockBTreeNode(messageNode.DataIdentifier)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	messageHeapOnNode, err := file.GetHeapOnNode(messageDataNode)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	localDescriptors, err := file.GetLocalDescriptors(messageNode)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	propertyContext, err := file.GetPropertyContext(messageHeapOnNode)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	message := &Message{
		File:             file,
		PropertyContext:  propertyContext,
		LocalDescriptors: localDescriptors,
	}

	if err := propertyContext.LoadProperties(message, localDescriptors...); err != nil {
		return nil, errors.WithStack(err)
	}

	return message, nil
}

//go:embed properties.csv
var PropertyMapCSV string

// PropertyMap maps the property ID to the struct JSON name.
var PropertyMap = make(map[string]string)

func init() {
	propertyMapReader := csv.NewReader(strings.NewReader(PropertyMapCSV))

	csvProperties, err := propertyMapReader.ReadAll()

	if err != nil {
		panic("go-pst: failed to initialize property map")
	}

	for _, row := range csvProperties {
		PropertyMap[row[0]] = row[1]
	}
}
