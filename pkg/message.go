package pst

import (
	"errors"
	"fmt"
)

// Message represents a message.
type Message struct {
	PropertyContext  []PropertyContextItem
	LocalDescriptors []LocalDescriptor
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

	return propertyContextItem.GetString()
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
