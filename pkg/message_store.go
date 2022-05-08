// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

// MessageStore represents the message store.
type MessageStore struct {
	PropertyContext []PropertyContextItem
}

// GetMessageStore returns the message store of the PST file.
func (pstFile *File) GetMessageStore(formatType string, encryptionType string) (MessageStore, error) {
	nodeBTreeNode, err := pstFile.GetNodeBTreeNode(IdentifierTypeMessageStore)

	if err != nil {
		return MessageStore{}, err
	}

	blockBTreeNode, err := pstFile.GetBlockBTreeNode(nodeBTreeNode.DataIdentifier)

	if err != nil {
		return MessageStore{}, err
	}

	heapOnNode, err := pstFile.NewHeapOnNodeFromNode(blockBTreeNode, formatType, encryptionType)

	if err != nil {
		return MessageStore{}, err
	}

	propertyContext, err := pstFile.GetPropertyContext(heapOnNode, formatType, encryptionType)

	if err != nil {
		return MessageStore{}, err
	}

	return MessageStore{
		PropertyContext: propertyContext,
	}, nil
}
