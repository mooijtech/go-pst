package pst

import (
	"errors"
	log "github.com/sirupsen/logrus"
)

// Message represents a message.
type Message struct {
	PropertyContext []PropertyContextItem
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

	tableContext, err := pstFile.GetTableContext(emailsHeapOnNode, localDescriptors, formatType, encryptionType, 0, -1, 26610)

	if err != nil {
		return nil, err
	}

	return tableContext, nil
}

// GetMessages returns an array of messages from the message table context.
func (pstFile *File) GetMessages(folder Folder, formatType string, encryptionType string) error {
	if folder.MessageCount == 0 {
		return nil
	}

	identifierType := folder.Identifier & 0x1F

	if identifierType == IdentifierTypeSearchFolder {
		return nil
	}

	messageTableContext, err := pstFile.GetMessageTableContext(folder, formatType, encryptionType)

	if err != nil {
		return err
	}

	for _, messageTableContextRow := range messageTableContext {
		for _, messageTableContextColumn := range messageTableContextRow {
			if messageTableContextColumn.PropertyID == 26610 {
				log.Infof("Processing message: %d", messageTableContextColumn.ReferenceHNID)

				err := pstFile.GetMessage(messageTableContextColumn.ReferenceHNID, formatType, encryptionType)

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// GetMessage returns the message of the identifier.
func (pstFile *File) GetMessage(identifier int, formatType string, encryptionType string) error {
	identifierType := identifier & 0x1F

	if identifierType != IdentifierTypeNormalMessage {
		return errors.New("invalid identifier type")
	}

	messageNode, err := pstFile.GetNodeBTreeNode(identifier, formatType)

	if err != nil {
		return err
	}

	messageNodeDataIdentifier, err := messageNode.GetDataIdentifier(formatType)

	if err != nil {
		return err
	}

	messageDataNode, err := pstFile.GetBlockBTreeNode(messageNodeDataIdentifier, formatType)

	if err != nil {
		return err
	}

	messageHeapOnNode, err := pstFile.NewHeapOnNodeFromNode(messageDataNode, formatType, encryptionType)

	if err != nil {
		return err
	}

	localDescriptors, err := pstFile.GetLocalDescriptors(messageNode, formatType)

	if err != nil {
		return err
	}

	tableContext, err := pstFile.GetTableContext(messageHeapOnNode, localDescriptors, formatType, encryptionType, 0, 1, 26)

	if err != nil {
		return err
	}

	log.Infof("Table context: %s", tableContext)

	return nil
}