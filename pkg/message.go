package pst

import log "github.com/sirupsen/logrus"

// GetMessageTableContext returns the message table context of this folder.
func (pstFile *File) GetMessageTableContext(folder Folder, formatType string) ([][]TableContextItem, error) {
	emailsIdentifier := folder.Identifier + 12

	nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

	if err != nil {
		return nil, err
	}

	emailsNode, err := pstFile.FindBTreeNode(nodeBTreeOffset, emailsIdentifier, formatType)

	if err != nil {
		return nil, err
	}

	emailsDataIdentifier, err := emailsNode.GetDataIdentifier(formatType)

	if err != nil {
		return nil, err
	}

	localDescriptors, err := pstFile.GetLocalDescriptors(emailsNode, formatType)

	if err != nil {
		return nil, err
	}

	blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

	if err != nil {
		return nil, err
	}

	emailsDataNode, err := pstFile.FindBTreeNode(blockBTreeOffset, emailsDataIdentifier, formatType)

	if err != nil {
		return nil, err
	}

	emailsHeapOnNode, err := pstFile.GetHeapOnNode(emailsDataNode, formatType)

	if err != nil {
		return nil, err
	}

	tableContext, err := pstFile.GetTableContext(emailsHeapOnNode, localDescriptors, formatType, 0, -1, 26610)

	if err != nil {
		return nil, err
	}

	return tableContext, nil
}

// GetMessages returns an array of messages from the message table context.
func (pstFile *File) GetMessages(folder Folder, formatType string) error {
	if folder.MessageCount == 0 {
		return nil
	}

	identifierType := folder.Identifier & 0x1F

	if identifierType == IdentifierTypeSearchFolder {
		return nil
	}

	messageTableContext, err := pstFile.GetMessageTableContext(folder, formatType)

	if err != nil {
		return err
	}

	for _, messageTableContextRow := range messageTableContext {
		for _, messageTableContextColumn := range messageTableContextRow {
			log.Infof("Property ID: %d", messageTableContextColumn.PropertyID)
		}
	}

	return nil
}