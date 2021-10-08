// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package main

import (
	pst "github.com/mooijtech/go-pst/pkg"
	log "github.com/sirupsen/logrus"
)

func main() {
	pstFile := pst.New("data/enron.pst")

	log.Infof("Parsing file: %s", pstFile.Filepath)

	isValidSignature, err := pstFile.IsValidSignature()

	if err != nil {
		log.Errorf("Failed to read signature: %s", err)
		return
	}

	if !isValidSignature {
		log.Errorf("Invalid file signature.")
		return
	}

	contentType, err := pstFile.GetContentType()

	if err != nil {
		log.Errorf("Failed to get content type: %s", err)
		return
	}

	log.Infof("Content type: %s", contentType)

	formatType, err := pstFile.GetFormatType()

	if err != nil {
		log.Errorf("Failed to get format type: %s", err)
		return
	}

	log.Infof("Format type: %s", formatType)

	encryptionType, err := pstFile.GetEncryptionType(formatType)

	if err != nil {
		log.Errorf("Failed to get encryption type: %s", err)
		return
	}

	log.Infof("Encryption type: %s", encryptionType)

	err = pstFile.InitializeBTrees(formatType)

	if err != nil {
		log.Errorf("Failed to initialize node and block b-tree.")
		return
	}

	rootFolder, err := pstFile.GetRootFolder(formatType, encryptionType)

	if err != nil {
		log.Errorf("Failed to get root folder: %s", err)
		return
	}

	err = GetSubFolders(pstFile, rootFolder, formatType, encryptionType)

	if err != nil {
		log.Errorf("Failed to get sub-folders: %s", err)
		return
	}
}

// GetSubFolders is a recursive function which retrieves all sub-folders for the specified folder.
func GetSubFolders(pstFile pst.File, folder pst.Folder, formatType string, encryptionType string) error {
	subFolders, err := pstFile.GetSubFolders(folder, formatType, encryptionType)

	if err != nil {
		return err
	}

	for _, subFolder := range subFolders {
		log.Infof("Parsing sub-folder: %s", subFolder.DisplayName)

		messages, err := pstFile.GetMessages(subFolder, formatType, encryptionType)

		if err != nil {
			return err
		}

		if len(messages) > 0 {
			log.Infof("Found %d messages.", len(messages))
		}

		err = GetSubFolders(pstFile, subFolder, formatType, encryptionType)

		if err != nil {
			return err
		}
	}

	return nil
}