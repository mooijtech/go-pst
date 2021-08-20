// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import log "github.com/sirupsen/logrus"

// GetRootFolder returns the root folder of the PST file.
func (pstFile *File) GetRootFolder(formatType string) error {
	nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

	if err != nil {
		return err
	}

	rootFolderNode, err := pstFile.FindBTreeNode(nodeBTreeOffset, IdentifierTypeRootFolder, formatType)

	if err != nil {
		return err
	}

	rootFolderNodeDataIdentifier, err := rootFolderNode.GetDataIdentifier(formatType)

	if err != nil {
		return err
	}

	blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

	if err != nil {
		return err
	}

	rootFolderDataNode, err := pstFile.FindBTreeNode(blockBTreeOffset, rootFolderNodeDataIdentifier, formatType)

	if err != nil {
		return err
	}

	rootFolderNodeDataNodeHeapOnNode, err := pstFile.GetHeapOnNode(rootFolderDataNode, formatType)

	if err != nil {
		return err
	}

	propertyContextItems, err := pstFile.GetPropertyContext(rootFolderNodeDataNodeHeapOnNode, formatType)

	if err != nil {
		return err
	}

	log.Infof("Property context items: %s", propertyContextItems)

	// Get sub-folders.
	subFoldersIdentifier := IdentifierTypeRootFolder + 11

	log.Infof("IT'S: %d", subFoldersIdentifier)

	subFoldersNode, err := pstFile.FindBTreeNode(nodeBTreeOffset, subFoldersIdentifier, formatType)

	if err != nil {
		return err
	}

	subFoldersDataIdentifier, err := subFoldersNode.GetDataIdentifier(formatType)

	log.Infof("DATA IDENTIFIER!: %d", subFoldersDataIdentifier)

	if err != nil {
		return err
	}

	subFoldersDataNode, err := pstFile.FindBTreeNode(blockBTreeOffset, subFoldersDataIdentifier, formatType)

	if err != nil {
		return err
	}

	log.Infof("GOT IT: %b", subFoldersDataNode.Data)

	return nil
}
