// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

// Folder represents a folder.
type Folder struct {
	Identifier int
	PropertyContext []PropertyContextItem
	TableContext []TableContextItem
}

// GetRootFolder returns the root folder of the PST file.
func (pstFile *File) GetRootFolder(formatType string) (Folder, error) {
	nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

	if err != nil {
		return Folder{}, err
	}

	rootFolderNode, err := pstFile.FindBTreeNode(nodeBTreeOffset, IdentifierTypeRootFolder, formatType)

	if err != nil {
		return Folder{}, err
	}

	rootFolderNodeDataIdentifier, err := rootFolderNode.GetDataIdentifier(formatType)

	if err != nil {
		return Folder{}, err
	}

	blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

	if err != nil {
		return Folder{}, err
	}

	rootFolderDataNode, err := pstFile.FindBTreeNode(blockBTreeOffset, rootFolderNodeDataIdentifier, formatType)

	if err != nil {
		return Folder{}, err
	}

	rootFolderNodeDataNodeHeapOnNode, err := pstFile.GetHeapOnNode(rootFolderDataNode, formatType)

	if err != nil {
		return Folder{}, err
	}

	propertyContextItems, err := pstFile.GetPropertyContext(rootFolderNodeDataNodeHeapOnNode, formatType)

	if err != nil {
		return Folder{}, err
	}

	return Folder{
		Identifier: IdentifierTypeRootFolder,
		PropertyContext: propertyContextItems,
	}, nil
}

func (pstFile *File) GetSubFolderTableContext(folder Folder, formatType string) ([]TableContextItem, error) {
	nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

	if err != nil {
		return nil, err
	}

	subFoldersIdentifier := folder.Identifier + 11 // +11 returns the identifier of the sub-folders.

	subFoldersNode, err := pstFile.FindBTreeNode(nodeBTreeOffset, subFoldersIdentifier, formatType)

	if err != nil {
		return nil, err
	}

	subFoldersDataIdentifier, err := subFoldersNode.GetDataIdentifier(formatType)

	if err != nil {
		return nil, err
	}

	blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

	if err != nil {
		return nil, err
	}

	subFoldersDataNode, err := pstFile.FindBTreeNode(blockBTreeOffset, subFoldersDataIdentifier, formatType)

	if err != nil {
		return nil, err
	}

	subFoldersDataNodeHeapOnNode, err := pstFile.GetHeapOnNode(subFoldersDataNode, formatType)

	if err != nil {
		return nil, err
	}

	tableContext, err := pstFile.GetTableContext(subFoldersDataNodeHeapOnNode, formatType, -1, -1)

	if err != nil {
		return nil, err
	}

	return tableContext, nil
}

// GetSubFolders returns the sub folders of this folder.
func (pstFile *File) GetSubFolders(folder Folder, formatType string) ([]Folder, error) {
	tableContext, err := pstFile.GetSubFolderTableContext(folder, formatType)

	if err != nil {
		// TODO - What are we doing wrong? It's a property context instead of table context.
		log.Warnf("Failed to get table context of folder (%d): %s", folder.Identifier, err)
		return nil, err
	}

	var subFolders []Folder

	for _, tableContextItem := range tableContext {

		if fmt.Sprintf("%x", tableContextItem.PropertyID) == "3001" {
			log.Infof("Display name: %s", string(tableContextItem.Data))
		}

		if tableContextItem.PropertyID == 26610 {
			// References btreeNodeEntryHeapOnNode.GetIdentifierType
			identifierType := tableContextItem.ReferenceHNID & 0x1F

			if identifierType == IdentifierTypeNormalFolder || identifierType == IdentifierTypeSearchFolder {
				// Find the data node from the reference HNID.
				nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

				if err != nil {
					return nil, err
				}

				node, err := pstFile.FindBTreeNode(nodeBTreeOffset, tableContextItem.ReferenceHNID, formatType)

				if err != nil {
					continue
				}

				dataIdentifier, err := node.GetDataIdentifier(formatType)

				if err != nil {
					return nil, err
				}

				blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

				if err != nil {
					return nil, err
				}

				dataNode, err := pstFile.FindBTreeNode(blockBTreeOffset, dataIdentifier, formatType)

				if err != nil {
					return nil, err
				}

				heapOnNode, err := pstFile.GetHeapOnNode(dataNode, formatType)

				if err != nil {
					return nil, err
				}

				propertyContext, err := pstFile.GetPropertyContext(heapOnNode, formatType)

				if err != nil {
					return nil, err
				}

				subFolders = append(subFolders, Folder {
					Identifier: tableContextItem.ReferenceHNID,
					PropertyContext: propertyContext,
				})
			}
		}
	}

	return subFolders, nil
}