// go-pst is a library for reading Personal Storage Table (.pst) files (written in Go/Golang).
//
// Copyright 2023 Marten Mooij
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pst

import (
	"github.com/mooijtech/go-pst/v6/pkg/properties"
	"github.com/rotisserie/eris"
)

// Folder represents a folder.
// The folder is writable using a FolderWriter.
type Folder struct {
	Identifier Identifier
	// Properties are populate by the PropertyContext.
	// See GetPropertyContext.
	Properties *properties.Folder

	// TODO - Remove pointer so this can be used by the writer separately?
	File *File
}

// NewFolder creates a new Folder with the properties for the FolderWriter.
//func NewFolder(properties *properties.Folder) *Folder {
//	return &Folder{
//		// Identifier is set by the FolderWriter, TODO move here?
//		//Identifier:
//		Properties: properties,
//	}
//}

// NewFolderWithIdentifier creates a Folder with the specified identifier for the FolderWriter.
//func NewFolderWithIdentifier(identifier Identifier, properties *properties.Folder) *Folder {
//	return &Folder{
//		Identifier: identifier,
//		Properties: properties,
//	}
//}

// GetPropertyContext returns the PropertyContext of the Folder.
func (folder *Folder) GetPropertyContext() (*PropertyContext, error) {
	rootFolderDataNode, err := folder.File.GetDataBTreeNode(IdentifierRootFolder)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get data b-tree node")
	}

	rootFolderHeapOnNode, err := folder.File.GetHeapOnNode(rootFolderDataNode)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get Heap-on-Node")
	}

	propertyContext, err := folder.File.GetPropertyContext(rootFolderHeapOnNode)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get property context")
	}

	return propertyContext, nil
}

// GetRootFolder returns the root folder of the PST file.
func (file *File) GetRootFolder() (*Folder, error) {
	return &Folder{
		Identifier: IdentifierRootFolder,
		Properties: &properties.Folder{
			Name: "IdentifierRootFolder",
			// TODO - Extend
		},
	}, nil
}

// GetSubFoldersTableContext returns the TableContext for the sub-folders of this folder.
// Note this limits the returned properties to the ones we use in the Folder struct.
func (folder *Folder) GetSubFoldersTableContext() (TableContext, error) {
	nodeBTreeNode, err := folder.File.GetNodeBTreeNode(folder.Identifier + 11) // +11 is the identifier of the sub-folders.

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get node b-tree node")
	}

	localDescriptors, err := folder.File.GetLocalDescriptors(nodeBTreeNode)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get local descriptors")
	}

	blockBTreeNode, err := folder.File.GetBlockBTreeNode(nodeBTreeNode.DataIdentifier)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get block b-tree node")
	}

	subFoldersDataNodeHeapOnNode, err := folder.File.GetHeapOnNode(blockBTreeNode)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get Heap-on-Node")
	}

	tableContext, err := folder.File.GetTableContext(subFoldersDataNodeHeapOnNode, localDescriptors, 12289, 13834, 26610, 13826)

	if err != nil {
		return TableContext{}, eris.Wrap(err, "failed to get table context")
	}

	return tableContext, nil
}

// GetSubFolders returns the sub-folders of this folder.
func (folder *Folder) GetSubFolders() ([]Folder, error) {
	if !folder.Properties.HasSubFolders { // TODO - Update from new properties.Folder
		// If there are actually no sub-folders this references a folder that doesn't exist.
		// java-libpst doesn't perform this check, so I assumed property ID 26610 always indicated there is a sub-folder.
		// Special thanks to James McLeod (https://github.com/Jmcleodfoss/pstreader) for telling me to check if there are actually sub-folders.
		return nil, nil
	}

	tableContext, err := folder.GetSubFoldersTableContext()

	if err != nil {
		return nil, eris.Wrap(err, "failed to get sub folders table context")
	}

	subFolders := make([]Folder, len(tableContext.Properties))

	for i, tableContextRow := range tableContext.Properties {
		subFolder := Folder{File: folder.File}

		for _, property := range tableContextRow {
			// We only return certain properties when calling GetSubFoldersTableContext,
			// so we don't need to check if we want to read this property.
			propertyReader, err := tableContext.GetPropertyReader(property)

			if err != nil {
				return nil, eris.Wrap(err, "failed to get property reader")
			}

			switch {
			case property.Identifier == 12289:
				// TODO - Check if this is String8 on ANSI FormatType.
				folderName, err := propertyReader.GetString()

				if err != nil {
					return nil, eris.Wrap(err, "failed to get folder name")
				}

				subFolder.Properties.Name = folderName
			case property.Identifier == 26610:
				identifier, err := propertyReader.GetInteger32()

				if err != nil {
					return nil, eris.Wrap(err, "failed to get identifier")
				}

				subFolder.Identifier = Identifier(identifier)
			case property.Identifier == 13826:
				messageCount, err := propertyReader.GetInteger32()

				if err != nil {
					return nil, eris.Wrap(err, "failed to get message count")
				}

				// TODO - Extend properties.Folder
				subFolder.Properties.MessageCount = messageCount
			case property.Identifier == 13834:
				hasSubFolders, err := propertyReader.GetBoolean()

				if err != nil {
					return nil, eris.Wrap(err, "failed to get has sub folders")
				}

				// TODO - Extend properties.Folder
				subFolder.Properties.HasSubFolders = hasSubFolders
			}
		}

		subFolders[i] = subFolder
	}

	return subFolders, nil
}

// WalkFolderFunc is the type of the function called by WalkFolders when visiting each folder.
type WalkFolderFunc = func(folder *Folder) error

// WalkFolders walks all folders recursively.
func (file *File) WalkFolders(walkFolderFunc WalkFolderFunc) error {
	rootFolder, err := file.GetRootFolder()

	if err != nil {
		return eris.Wrap(err, "failed to get root folder")
	}

	return rootFolder.WalkFolders(walkFolderFunc)
}

// WalkFolders recursively walks the sub-folders of this folder.
func (folder *Folder) WalkFolders(walkFolderFunc WalkFolderFunc) error {
	if err := walkFolderFunc(folder); err != nil {
		return eris.Wrap(err, "failed during calling walk folder function")
	}

	subFolders, err := folder.GetSubFolders()

	if err != nil {
		return eris.Wrap(err, "failed to get sub folders")
	}

	for _, subFolder := range subFolders {
		if err := subFolder.WalkFolders(walkFolderFunc); err != nil {
			return eris.Wrap(err, "failed to walk folders")
		}
	}

	return nil
}
