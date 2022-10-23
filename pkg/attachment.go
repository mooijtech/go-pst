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
	"github.com/mooijtech/go-pst/v5/pkg/properties"
	"io"

	"github.com/pkg/errors"
)

// Attachment represents a message attachment.
type Attachment struct {
	PropertyContext  *PropertyContext
	LocalDescriptors []LocalDescriptor
	properties.Attachment
}

// HasAttachments returns true if this message has attachments.
func (message *Message) HasAttachments() (bool, error) {
	reader, err := message.PropertyContext.GetPropertyReader(3591, message.LocalDescriptors...)

	if err != nil {
		return false, errors.WithStack(err)
	}

	value, err := reader.GetInteger32()

	if err != nil {
		return false, errors.WithStack(err)
	}

	return value&0x10 != 0, nil
}

// GetAttachmentTableContext returns the table context of the attachments of this message.
// Note we only return the attachment identifier property.
func (message *Message) GetAttachmentTableContext() (*TableContext, error) {
	hasAttachments, err := message.HasAttachments()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if !hasAttachments {
		return nil, errors.WithStack(ErrAttachmentsNotFound)
	}

	if message.AttachmentTableContext == nil {
		// Initialize the attachments table context.
		attachmentLocalDescriptor, err := FindLocalDescriptor(1649, message.LocalDescriptors)

		if err != nil {
			return nil, errors.WithStack(err)
		}

		attachmentHeapOnNode, err := message.File.GetHeapOnNodeFromLocalDescriptor(attachmentLocalDescriptor)

		if err != nil {
			return nil, errors.WithStack(err)
		}

		attachmentLocalDescriptors, err := message.File.GetLocalDescriptorsFromIdentifier(attachmentLocalDescriptor.LocalDescriptorsIdentifier)

		if err != nil {
			return nil, errors.WithStack(err)
		}

		attachmentTableContext, err := message.File.GetTableContext(attachmentHeapOnNode, attachmentLocalDescriptors, 26610)

		if err != nil {
			return nil, errors.WithStack(err)
		}

		message.AttachmentTableContext = &attachmentTableContext
	}

	return message.AttachmentTableContext, nil
}

// GetAttachmentCount returns the amount of rows in the attachment table context.
func (message *Message) GetAttachmentCount() (int, error) {
	attachmentTableContext, err := message.GetAttachmentTableContext()

	if errors.Is(err, ErrAttachmentsNotFound) {
		return 0, nil
	} else if err != nil {
		return 0, errors.WithStack(err)
	}

	return len(attachmentTableContext.Properties), nil
}

// GetAttachment returns the specified attachment.
func (message *Message) GetAttachment(attachmentIndex int) (*Attachment, error) {
	attachmentsTableContext, err := message.GetAttachmentTableContext()

	if err != nil {
		return nil, errors.WithStack(err)
	} else if attachmentIndex > len(attachmentsTableContext.Properties)-1 {
		return nil, errors.WithStack(ErrAttachmentIndexInvalid)
	}

	var attachmentHNID Identifier

	for _, attachmentProperty := range attachmentsTableContext.Properties[attachmentIndex] {
		// We only get the attachment identifier property from GetAttachmentTableContext.
		propertyReader, err := attachmentsTableContext.GetPropertyReader(attachmentProperty)

		if err != nil {
			return nil, errors.WithStack(err)
		}

		identifier, err := propertyReader.GetInteger32()

		if err != nil {
			return nil, errors.WithStack(err)
		}

		attachmentHNID = Identifier(identifier)
	}

	if attachmentHNID == 0 {
		return nil, errors.WithStack(errors.New("go-pst: failed to get attachment HNID"))
	}

	attachmentLocalDescriptor, err := FindLocalDescriptor(attachmentHNID, message.LocalDescriptors)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	attachmentLocalDescriptors, err := message.File.GetLocalDescriptorsFromIdentifier(attachmentLocalDescriptor.LocalDescriptorsIdentifier)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	attachmentHeapOnNode, err := message.File.GetHeapOnNodeFromLocalDescriptor(attachmentLocalDescriptor)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	attachmentPropertyContext, err := message.File.GetPropertyContext(attachmentHeapOnNode)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	attachment := &Attachment{
		PropertyContext:  attachmentPropertyContext,
		LocalDescriptors: attachmentLocalDescriptors,
	}

	if err := attachmentPropertyContext.LoadProperties(attachment, attachmentLocalDescriptors...); err != nil {
		return nil, errors.WithStack(err)
	}

	return attachment, nil
}

// GetAllAttachments returns the attachments of this message.
// See AttachmentIterator.
func (message *Message) GetAllAttachments() ([]*Attachment, error) {
	attachmentCount, err := message.GetAttachmentCount()

	if errors.Is(err, ErrAttachmentsNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	attachments := make([]*Attachment, attachmentCount)

	for i := 0; i < attachmentCount; i++ {
		attachment, err := message.GetAttachment(i)

		if err != nil {
			return nil, errors.WithStack(err)
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
	hasNext := len(attachmentIterator.message.AttachmentTableContext.Properties)-1 > attachmentIterator.currentIndex

	if !hasNext {
		return false
	}

	attachment, err := attachmentIterator.message.GetAttachment(attachmentIterator.currentIndex)

	if err != nil {
		attachmentIterator.err = errors.WithStack(err)
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

	if err != nil {
		return AttachmentIterator{}, errors.WithStack(err)
	} else if attachmentCount == 0 {
		return AttachmentIterator{}, ErrAttachmentsNotFound
	}

	return AttachmentIterator{
		message: message,
	}, nil
}

// WriteTo writes the attachment to the specified io.Writer.
func (attachment *Attachment) WriteTo(writer io.Writer) (int64, error) {
	attachmentReader, err := attachment.PropertyContext.GetPropertyReader(14081, attachment.LocalDescriptors...)

	if errors.Is(err, ErrPropertyNoData) {
		return 0, nil
	} else if err != nil {
		return 0, errors.WithStack(err)
	}

	sectionReader := io.NewSectionReader(&attachmentReader, 0, attachmentReader.Size())

	written, err := io.CopyN(writer, sectionReader, sectionReader.Size())

	if err != nil {
		return written, errors.WithStack(err)
	}

	return written, nil
}
