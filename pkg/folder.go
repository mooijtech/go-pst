// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

// Folder represents a folder.
type Folder struct {
	Identifier      int
	DisplayName     string
	HasSubFolders   bool
	MessageCount    int
	PropertyContext []PropertyContextItem
}

// GetRootFolder returns the root folder of the PST file.
func (pstFile *File) GetRootFolder(formatType string, encryptionType string) (Folder, error) {
	rootFolderDataNode, err := pstFile.GetDataBTreeNode(IdentifierTypeRootFolder, formatType)

	if err != nil {
		return Folder{}, err
	}

	rootFolderNodeDataNodeHeapOnNode, err := pstFile.NewHeapOnNodeFromNode(rootFolderDataNode, formatType, encryptionType)

	if err != nil {
		return Folder{}, err
	}

	propertyContextItems, err := pstFile.GetPropertyContext(rootFolderNodeDataNodeHeapOnNode, formatType, encryptionType)

	if err != nil {
		return Folder{}, err
	}

	return Folder{
		Identifier:      IdentifierTypeRootFolder,
		DisplayName:     "ROOT_FOLDER",
		HasSubFolders:   true,
		PropertyContext: propertyContextItems,
	}, nil
}

// GetSubFolderTableContext returns the table context for the sub-folders of this folder.
func (pstFile *File) GetSubFolderTableContext(folder Folder, formatType string, encryptionType string) ([][]TableContextItem, error) {
	subFoldersIdentifier := folder.Identifier + 11 // +11 returns the identifier of the sub-folders.

	subFoldersDataNode, err := pstFile.GetDataBTreeNode(subFoldersIdentifier, formatType)

	if err != nil {
		return nil, err
	}

	subFoldersDataNodeHeapOnNode, err := pstFile.NewHeapOnNodeFromNode(subFoldersDataNode, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	tableContext, err := pstFile.GetTableContext(subFoldersDataNodeHeapOnNode, []LocalDescriptor{}, formatType, encryptionType, -1, -1, -1)

	if err != nil {
		return nil, err
	}

	return tableContext, nil
}

// GetSubFolders returns the sub folders of this folder.
func (pstFile *File) GetSubFolders(folder Folder, formatType string, encryptionType string) ([]Folder, error) {
	if !folder.HasSubFolders {
		// This is supposed to be a sub-folder but
		// if there are actually no sub-folders this references a folder that doesn't exist.
		// This caused an issue where the table context was not found (because the folder doesn't exist).
		// References go-pst issue #1.
		// java-libpst doesn't perform this check so I assumed "26610" always indicated there is a sub-folder.
		// Special thanks to James McLeod (https://github.com/Jmcleodfoss/pstreader) for telling me to check if there are actually sub-folders.
		return []Folder{}, nil
	}

	tableContext, err := pstFile.GetSubFolderTableContext(folder, formatType, encryptionType)

	if err != nil {
		return nil, err
	}

	var subFolders []Folder

	for _, tableContextRow := range tableContext {
		var subFolder Folder

		for _, tableContextColumn := range tableContextRow {
			if tableContextColumn.PropertyID == 12289 {
				displayName, err := DecodeBytesToUTF16String(tableContextColumn.Data)

				if err != nil {
					return nil, err
				}

				subFolder.DisplayName = displayName
			} else if tableContextColumn.PropertyID == 13834 {
				subFolder.HasSubFolders = tableContextColumn.ReferenceHNID == 1
			} else if tableContextColumn.PropertyID == 26610 {
				subFolder.Identifier = tableContextColumn.ReferenceHNID

				subFolders = append(subFolders, subFolder)
			} else if tableContextColumn.PropertyID == 13826 {
				subFolder.MessageCount = tableContextColumn.ReferenceHNID
			}
		}
	}

	return subFolders, nil
}
