// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// Attachment represents a message attachment.
type Attachment struct {
	PropertyContext  []PropertyContextItem
	LocalDescriptors []LocalDescriptor
}

// HasAttachments returns true if this message has attachments.
func (message *Message) HasAttachments() (bool, error) {
	hasAttachments, err := message.GetInteger(3591)

	if err != nil {
		return false, err
	}

	return hasAttachments&0x10 != 0, nil
}

// GetAttachmentsTableContext returns the table context of the attachments of this message.
func (message *Message) GetAttachmentsTableContext(pstFile *File, formatType string, encryptionType string) ([][]TableContextItem, error) {
	if len(message.AttachmentsTableContext) == 0 {
		// Initialize the attachments table context.
		attachmentsLocalDescriptor, err := FindLocalDescriptor(message.LocalDescriptors, 1649, formatType)

		if err != nil {
			return nil, err
		}

		attachmentsHeapOnNode, err := pstFile.NewHeapOnNodeFromLocalDescriptor(attachmentsLocalDescriptor, formatType, encryptionType)

		if err != nil {
			return nil, err
		}

		attachmentsLocalDescriptorLocalDescriptorsIdentifier, err := attachmentsLocalDescriptor.GetLocalDescriptorsIdentifier(formatType)

		if err != nil {
			return nil, err
		}

		attachmentsLocalDescriptors, err := pstFile.GetLocalDescriptorsFromIdentifier(attachmentsLocalDescriptorLocalDescriptorsIdentifier, formatType)

		if err != nil {
			return nil, err
		}

		attachmentsTableContext, err := pstFile.GetTableContext(attachmentsHeapOnNode, attachmentsLocalDescriptors, formatType, encryptionType, -1, -1, -1)

		if err != nil {
			return nil, err
		}

		message.AttachmentsTableContext = attachmentsTableContext

		return attachmentsTableContext, nil
	}

	return message.AttachmentsTableContext, nil
}

// GetAttachmentsCount returns the amount of rows in the attachments table context.
func (message *Message) GetAttachmentsCount(pstFile *File, formatType string, encryptionType string) (int, error) {
	attachmentsTableContext, err := message.GetAttachmentsTableContext(pstFile, formatType, encryptionType)

	if err != nil {
		return -1, err
	}

	return len(attachmentsTableContext), nil
}

// GetAttachment returns the specified attachment.
func (message *Message) GetAttachment(attachmentNumber int, pstFile *File, formatType string, encryptionType string) (Attachment, error) {
	hasAttachments, err := message.HasAttachments()

	if err != nil {
		return Attachment{}, err
	}

	if !hasAttachments {
		return Attachment{}, nil
	}

	attachmentsTableContext, err := message.GetAttachmentsTableContext(pstFile, formatType, encryptionType)

	if err != nil {
		return Attachment{}, err
	}

	if attachmentNumber > len(attachmentsTableContext) {
		return Attachment{}, errors.New(fmt.Sprintf("invalid attachment number, there are only %d attachments", len(attachmentsTableContext)))
	}

	attachmentTableContextItem := attachmentsTableContext[attachmentNumber]

	var attachmentReferenceHNID int

	for _, attachmentTableContextItemColumn := range attachmentTableContextItem {
		if attachmentTableContextItemColumn.PropertyID == 26610 {
			attachmentReferenceHNID = attachmentTableContextItemColumn.ReferenceHNID
			break
		}
	}

	attachmentLocalDescriptor, err := FindLocalDescriptor(message.LocalDescriptors, attachmentReferenceHNID, formatType)

	if err != nil {
		return Attachment{}, err
	}

	attachmentLocalDescriptorLocalDescriptorsIdentifier, err := attachmentLocalDescriptor.GetLocalDescriptorsIdentifier(formatType)

	if err != nil {
		return Attachment{}, err
	}

	attachmentLocalDescriptorLocalDescriptors, err := pstFile.GetLocalDescriptorsFromIdentifier(attachmentLocalDescriptorLocalDescriptorsIdentifier, formatType)

	if err != nil {
		return Attachment{}, err
	}

	attachmentHeapOnNode, err := pstFile.NewHeapOnNodeFromLocalDescriptor(attachmentLocalDescriptor, formatType, encryptionType)

	if err != nil {
		return Attachment{}, err
	}

	attachmentPropertyContext, err := pstFile.GetPropertyContext(attachmentHeapOnNode, formatType, encryptionType)

	if err != nil {
		return Attachment{}, err
	}

	return Attachment{
		PropertyContext:  attachmentPropertyContext,
		LocalDescriptors: attachmentLocalDescriptorLocalDescriptors,
	}, nil
}

// GetAttachments returns the attachments of this message.
func (message *Message) GetAttachments(pstFile *File, formatType string, encryptionType string) ([]Attachment, error) {
	hasAttachments, err := message.HasAttachments()

	if err != nil {
		return nil, err
	}

	if !hasAttachments {
		return nil, nil
	}

	attachmentsTableContext, err := message.GetAttachmentsTableContext(pstFile, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	var attachments []Attachment

	for i := 0; i < len(attachmentsTableContext); i++ {
		attachment, err := message.GetAttachment(i, pstFile, formatType, encryptionType)

		if err != nil {
			return nil, err
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// GetString returns the string value of the property.
func (attachment *Attachment) GetString(propertyID int) (string, error) {
	propertyContextItem, err := FindPropertyContextItem(attachment.PropertyContext, propertyID)

	if err != nil {
		return "", err
	}

	return DecodeBytesToUTF16String(propertyContextItem.data)
}

// GetFilename returns the file name of this attachment.
func (attachment *Attachment) GetFilename() (string, error) {
	return attachment.GetString(14084)
}

// GetLongFilename returns the long file name of this attachment.
func (attachment *Attachment) GetLongFilename() (string, error) {
	return attachment.GetString(14087)
}

// GetInputStream returns the input stream of this attachment.
func (attachment *Attachment) GetInputStream(pstFile *File, formatType string, encryptionType string) (HeapOnNodeInputStream, error) {
	attachmentInputStreamPropertyContextItem, err := FindPropertyContextItem(attachment.PropertyContext, 14081)

	if err != nil {
		return HeapOnNodeInputStream{}, err
	}

	if attachmentInputStreamPropertyContextItem.IsExternalValueReference {
		attachmentInputStreamLocalDescriptor, err := FindLocalDescriptor(attachment.LocalDescriptors, attachmentInputStreamPropertyContextItem.ReferenceHNID, formatType)

		if err != nil {
			return HeapOnNodeInputStream{}, err
		}

		attachmentInputStreamHeapOnNode, err := pstFile.NewHeapOnNodeFromLocalDescriptor(attachmentInputStreamLocalDescriptor, formatType, encryptionType)

		if err != nil {
			return HeapOnNodeInputStream{}, err
		}

		return attachmentInputStreamHeapOnNode.InputStream, nil
	} else {
		return HeapOnNodeInputStream{
			UnencryptedInternalAttachmentData: attachmentInputStreamPropertyContextItem.data,
			Size:                              len(attachmentInputStreamPropertyContextItem.data),
		}, nil
	}
}

// WriteToFile writes the input stream of the attachment to the specified output path.
func (attachment *Attachment) WriteToFile(outputPath string, pstFile *File, formatType string, encryptionType string) error {
	attachmentInputStream, err := attachment.GetInputStream(pstFile, formatType, encryptionType)

	if err != nil {
		return err
	}

	outputBuffer, err := attachmentInputStream.ReadCompletely()

	if err != nil {
		return err
	}

	_, err = os.Create(outputPath)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outputPath, outputBuffer, 0755)

	if err != nil {
		return err
	}

	return nil
}
