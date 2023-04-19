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
	"github.com/mooijtech/go-pst/pkg/properties"
	"io"

	"github.com/rotisserie/eris"
)

// Attachment represents a message attachment.
type Attachment struct {
	PropertyContext  *PropertyContext
	LocalDescriptors []LocalDescriptor
	properties.Attachment
}

// HasAttachments returns true if this message has attachments.
func (message *Message) HasAttachments() (bool, error) {
	reader, err := message.PropertyContext.GetPropertyReader(3591, message.LocalDescriptors)

	if err != nil {
		return false, eris.Wrap(err, "failed to get property reader")
	}

	value, err := reader.GetInteger32()

	if err != nil {
		return false, eris.Wrap(err, "failed to read int32")
	}

	return value&0x10 != 0, nil
}

// GetAttachmentTableContext returns the table context of the attachments of this message.
// Note we only return the attachment identifier property.
func (message *Message) GetAttachmentTableContext() (*TableContext, error) {
	hasAttachments, err := message.HasAttachments()

	if err != nil {
		return nil, eris.Wrap(err, "failed to check if there are attachments")
	}

	if !hasAttachments {
		return nil, ErrAttachmentsNotFound
	}

	if message.AttachmentTableContext == nil {
		// Initialize the attachments table context.
		attachmentLocalDescriptor, err := FindLocalDescriptor(1649, message.LocalDescriptors)

		if err != nil {
			return nil, eris.Wrap(err, "failed to find attachment local descriptor")
		}

		attachmentHeapOnNode, err := message.File.GetHeapOnNodeFromLocalDescriptor(attachmentLocalDescriptor)

		if err != nil {
			return nil, eris.Wrap(err, "failed to get attachment Heap-on-Node")
		}

		attachmentLocalDescriptors, err := message.File.GetLocalDescriptorsFromIdentifier(attachmentLocalDescriptor.LocalDescriptorsIdentifier)

		if err != nil {
			return nil, eris.Wrap(err, "failed to get attachment local descriptors")
		}

		attachmentTableContext, err := message.File.GetTableContext(attachmentHeapOnNode, attachmentLocalDescriptors, 26610)

		if err != nil {
			return nil, eris.Wrap(err, "failed to get attachment table context")
		}

		message.AttachmentTableContext = &attachmentTableContext
	}

	return message.AttachmentTableContext, nil
}

// GetAttachmentCount returns the amount of rows in the attachment table context.
func (message *Message) GetAttachmentCount() (int, error) {
	attachmentTableContext, err := message.GetAttachmentTableContext()

	if eris.Is(err, ErrAttachmentsNotFound) {
		return 0, nil
	} else if err != nil {
		return 0, eris.Wrap(err, "failed to get attachment table context")
	}

	return len(attachmentTableContext.Properties), nil
}

// GetAttachment returns the specified attachment.
func (message *Message) GetAttachment(attachmentIndex int) (*Attachment, error) {
	attachmentsTableContext, err := message.GetAttachmentTableContext()

	if err != nil {
		return nil, eris.Wrap(err, "failed to get attachments table context")
	} else if attachmentIndex > len(attachmentsTableContext.Properties)-1 {
		return nil, ErrAttachmentIndexInvalid
	}

	var attachmentHNID Identifier

	for _, attachmentProperty := range attachmentsTableContext.Properties[attachmentIndex] {
		// We only get the attachment identifier property from GetAttachmentTableContext.
		propertyReader, err := attachmentsTableContext.GetPropertyReader(attachmentProperty)

		if err != nil {
			return nil, eris.Wrap(err, "failed to get attachments table context property reader")
		}

		identifier, err := propertyReader.GetInteger32()

		if err != nil {
			return nil, eris.Wrap(err, "failed to read identifier")
		}

		attachmentHNID = Identifier(identifier)
	}

	if attachmentHNID == 0 {
		return nil, eris.New("failed to get attachment HNID")
	}

	attachmentLocalDescriptor, err := FindLocalDescriptor(attachmentHNID, message.LocalDescriptors)

	if err != nil {
		return nil, eris.Wrap(err, "failed to find attachment local descriptor")
	}

	attachmentLocalDescriptors, err := message.File.GetLocalDescriptorsFromIdentifier(attachmentLocalDescriptor.LocalDescriptorsIdentifier)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get local descriptors from identifier")
	}

	attachmentHeapOnNode, err := message.File.GetHeapOnNodeFromLocalDescriptor(attachmentLocalDescriptor)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get attachment Heap-on-Node")
	}

	attachmentPropertyContext, err := message.File.GetPropertyContext(attachmentHeapOnNode)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get attachment property context")
	}

	attachment := &Attachment{
		PropertyContext:  attachmentPropertyContext,
		LocalDescriptors: attachmentLocalDescriptors,
	}

	if err := attachmentPropertyContext.Populate(attachment, attachmentLocalDescriptors); err != nil {
		return nil, eris.Wrap(err, "failed to populate attachment property context")
	}

	return attachment, nil
}

// GetAllAttachments returns the attachments of this message.
// See AttachmentIterator.
func (message *Message) GetAllAttachments() ([]*Attachment, error) {
	attachmentCount, err := message.GetAttachmentCount()

	if eris.Is(err, ErrAttachmentsNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, eris.Wrap(err, "failed to get attachment count")
	}

	attachments := make([]*Attachment, attachmentCount)

	for i := 0; i < attachmentCount; i++ {
		attachment, err := message.GetAttachment(i)

		if err != nil {
			return nil, eris.Wrap(err, "failed to get attachment")
		}

		attachments[i] = attachment
	}

	return attachments, nil
}

// AttachmentIterator implements an attachment iterator.
type AttachmentIterator struct {
	message *Message

	err               error
	currentIndex      int
	currentAttachment *Attachment
}

// Err return the error cause.
func (attachmentIterator *AttachmentIterator) Err() error {
	return attachmentIterator.err
}

// Next will ensure that Value returns the next item when executed.
// If the next value is not retrievable, Next will return false and Err() will return the error cause.
func (attachmentIterator *AttachmentIterator) Next() bool {
	hasNext := len(attachmentIterator.message.AttachmentTableContext.Properties) > attachmentIterator.currentIndex

	if !hasNext {
		return false
	}

	attachment, err := attachmentIterator.message.GetAttachment(attachmentIterator.currentIndex)

	if err != nil {
		attachmentIterator.err = eris.Wrap(err, "failed to get attachment")
		return false
	}

	attachmentIterator.currentIndex++
	attachmentIterator.currentAttachment = attachment

	return true
}

// Value returns the current value in the iterator.
func (attachmentIterator *AttachmentIterator) Value() *Attachment {
	return attachmentIterator.currentAttachment
}

// Size returns the amount of attachments in the message iterator.
func (attachmentIterator *AttachmentIterator) Size() int {
	return len(attachmentIterator.message.AttachmentTableContext.Properties)
}

func (attachmentIterator *AttachmentIterator) CurrentIndex() int {
	return attachmentIterator.currentIndex
}

// GetAttachmentIterator returns an iterator for attachments.
func (message *Message) GetAttachmentIterator() (AttachmentIterator, error) {
	attachmentCount, err := message.GetAttachmentCount()

	// TODO - Return an empty iterator instead of an error.
	if err != nil {
		return AttachmentIterator{}, eris.Wrap(err, "failed to get attachment count")
	} else if attachmentCount == 0 {
		return AttachmentIterator{}, ErrAttachmentsNotFound
	}

	return AttachmentIterator{
		message: message,
	}, nil
}

// WriteTo writes the attachment to the specified io.Writer.
func (attachment *Attachment) WriteTo(writer io.Writer) (int64, error) {
	attachmentReader, err := attachment.PropertyContext.GetPropertyReader(14081, attachment.LocalDescriptors)

	if eris.Is(err, ErrPropertyNoData) {
		return 0, nil
	} else if err != nil {
		return 0, eris.Wrap(err, "failed to get attachment property reader")
	}

	sectionReader := io.NewSectionReader(&attachmentReader, 0, attachmentReader.Size())

	written, err := io.CopyN(writer, sectionReader, sectionReader.Size())

	if err != nil {
		return written, eris.Wrap(err, "failed to write attachment")
	}

	return written, nil
}
