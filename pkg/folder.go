// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	log "github.com/sirupsen/logrus"
)

// Folder represents a folder.
type Folder struct {
	Identifier      int
	PropertyContext []PropertyContextItem
	TableContext    []TableContextItem
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
		Identifier:      IdentifierTypeRootFolder,
		PropertyContext: propertyContextItems,
	}, nil
}

// GetSubFolderTableContext returns the table context for the sub-folders of this folder.
func (pstFile *File) GetSubFolderTableContext(folder Folder, formatType string) ([][]TableContextItem, error) {
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
		return nil, err
	}

	var subFolders []Folder

	for _, tableContextRows := range tableContext {
		hasSubFolders := false

		for _, tableContextColumn := range tableContextRows {
			if tableContextColumn.PropertyID == 12289 {
				log.Infof("Display name: %s", string(tableContextColumn.Data))
			} else if tableContextColumn.PropertyID == 13834 {
				hasSubFolders = tableContextColumn.ReferenceHNID == 1
			} else if tableContextColumn.PropertyID == 26610 {
				if !hasSubFolders {
					// This is supposed to be a sub-folder but
					// if there are actually no sub-folders this references a folder that doesn't exist.
					// This caused an issue where the table context was not found (because the folder doesn't exist).
					// References go-pst issue #1.
					// java-libpst doesn't perform this check so I assumed "26610" always indicated there is a sub-folder.
					// Special thanks to James McLeod (https://github.com/Jmcleodfoss/pstreader) for telling me to check if there are actually sub-folders.
					continue
				}

				identifierType := tableContextColumn.ReferenceHNID & 0x1F

				if identifierType == IdentifierTypeNormalFolder || identifierType == IdentifierTypeSearchFolder {
					// Find the data node from the reference HNID.
					nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

					if err != nil {
						return nil, err
					}

					node, err := pstFile.FindBTreeNode(nodeBTreeOffset, tableContextColumn.ReferenceHNID, formatType)

					if err != nil {
						return nil, err
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

					subFolders = append(subFolders, Folder{
						Identifier:      tableContextColumn.ReferenceHNID,
						PropertyContext: propertyContext,
					})
				} else {
					log.Infof("It's a message!")
				}
			}
		}
	}

	return subFolders, nil
}
