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
	"github.com/pkg/errors"
)

// Folder represents a folder.
type Folder struct {
	Identifier      Identifier
	Name            string
	HasSubFolders   bool
	MessageCount    int32
	PropertyContext *PropertyContext
	File            *File
}

// GetRootFolder returns the root folder of the PST file.
func (file *File) GetRootFolder() (Folder, error) {
	rootFolderDataNode, err := file.GetDataBTreeNode(IdentifierRootFolder)

	if err != nil {
		return Folder{}, errors.WithStack(err)
	}

	rootFolderHeapOnNode, err := file.GetHeapOnNode(rootFolderDataNode)

	if err != nil {
		return Folder{}, errors.WithStack(err)
	}

	propertyContext, err := file.GetPropertyContext(rootFolderHeapOnNode)

	if err != nil {
		return Folder{}, errors.WithStack(err)
	}

	return Folder{
		Identifier:      IdentifierRootFolder,
		Name:            "ROOT_FOLDER",
		HasSubFolders:   true,
		MessageCount:    0,
		PropertyContext: propertyContext,
		File:            file,
	}, nil
}

// GetSubFoldersTableContext returns the TableContext for the sub-folders of this folder.
// Note this limits the returned properties to the ones we use in the Folder struct.
func (folder *Folder) GetSubFoldersTableContext() (TableContext, error) {
	subFoldersDataNode, err := folder.File.GetDataBTreeNode(folder.Identifier + 11) // +11 returns the identifier of the sub-folders.

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	subFoldersDataNodeHeapOnNode, err := folder.File.GetHeapOnNode(subFoldersDataNode)

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	tableContext, err := folder.File.GetTableContext(subFoldersDataNodeHeapOnNode, []LocalDescriptor{}, 12289, 13834, 26610, 13826)

	if err != nil {
		return TableContext{}, errors.WithStack(err)
	}

	return tableContext, nil
}

// GetSubFolders returns the sub-folders of this folder.
func (folder *Folder) GetSubFolders() ([]Folder, error) {
	if !folder.HasSubFolders {
		// If there are actually no sub-folders this references a folder that doesn't exist.
		// java-libpst doesn't perform this check so I assumed property ID 26610 always indicated there is a sub-folder.
		// Special thanks to James McLeod (https://github.com/Jmcleodfoss/pstreader) for telling me to check if there are actually sub-folders.
		return nil, nil
	}

	tableContext, err := folder.GetSubFoldersTableContext()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	subFolders := make([]Folder, len(tableContext.Properties))

	for i, tableContextRow := range tableContext.Properties {
		subFolder := Folder{File: folder.File}

		for _, property := range tableContextRow {
			// We only return certain properties when calling GetSubFoldersTableContext,
			// so we don't need to check if we want to read this property.
			propertyReader, err := tableContext.GetPropertyReader(property)

			if err != nil {
				return nil, errors.WithStack(err)
			}

			switch {
			case property.ID == 12289:
				// TODO - Check if this is String8 on ANSI FormatType.
				folderName, err := propertyReader.GetString()

				if err != nil {
					return nil, errors.WithStack(err)
				}

				subFolder.Name = folderName
			case property.ID == 26610:
				identifier, err := propertyReader.GetInteger32()

				if err != nil {
					return nil, errors.WithStack(err)
				}

				subFolder.Identifier = Identifier(identifier)
			case property.ID == 13826:
				messageCount, err := propertyReader.GetInteger32()

				if err != nil {
					return nil, errors.WithStack(err)
				}

				subFolder.MessageCount = messageCount
			case property.ID == 13834:
				hasSubFolders, err := propertyReader.GetBoolean()

				if err != nil {
					return nil, errors.WithStack(err)
				}

				subFolder.HasSubFolders = hasSubFolders
			}
		}

		subFolders[i] = subFolder
	}

	return subFolders, nil
}

// WalkFolderFunc is the type of the function called by WalkFolders when visiting each folder.
type WalkFolderFunc = func(folder Folder) error

// WalkFolders walks all folders recursively.
func (file *File) WalkFolders(walkFolderFunc WalkFolderFunc) error {
	rootFolder, err := file.GetRootFolder()

	if err != nil {
		return errors.WithStack(err)
	}

	return rootFolder.WalkFolders(walkFolderFunc)
}

// WalkFolders recursively walks the sub-folders of this folder.
func (folder *Folder) WalkFolders(walkFolderFunc WalkFolderFunc) error {
	if err := walkFolderFunc(*folder); err != nil {
		return errors.WithStack(err)
	}

	subFolders, err := folder.GetSubFolders()

	if err != nil {
		return errors.WithStack(err)
	}

	for _, subFolder := range subFolders {
		if err := subFolder.WalkFolders(walkFolderFunc); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
