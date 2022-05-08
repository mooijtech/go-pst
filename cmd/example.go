// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package main

import (
	"fmt"
	pst "github.com/mooijtech/go-pst/v4/pkg"
	"time"
)

func main() {
	startTime := time.Now()

	pstFile, err := pst.NewFromFile("data/enron.pst")

	if err != nil {
		fmt.Printf("Failed to create PST file: %s\n", err)
		return
	}

	defer func() {
		err := pstFile.Close()

		if err != nil {
			fmt.Printf("Failed to close PST file: %s", err)
		}
	}()

	fmt.Printf("Parsing file...")

	isValidSignature, err := pstFile.IsValidSignature()

	if err != nil {
		fmt.Printf("Failed to read signature: %s\n", err)
		return
	}

	if !isValidSignature {
		fmt.Printf("Invalid file signature.\n")
		return
	}

	contentType, err := pstFile.GetContentType()

	if err != nil {
		fmt.Printf("Failed to get content type: %s\n", err)
		return
	}

	fmt.Printf("Content type: %s\n", contentType)

	formatType, err := pstFile.GetFormatType()

	if err != nil {
		fmt.Printf("Failed to get format type: %s\n", err)
		return
	}

	fmt.Printf("Format type: %s\n", formatType)

	encryptionType, err := pstFile.GetEncryptionType(formatType)

	if err != nil {
		fmt.Printf("Failed to get encryption type: %s\n", err)
		return
	}

	fmt.Printf("Encryption type: %s\n", encryptionType)

	fmt.Printf("Initializing B-Trees...\n")

	err = pstFile.InitializeBTrees(formatType)

	if err != nil {
		fmt.Printf("Failed to initialize node and block b-tree.\n")
		return
	}

	rootFolder, err := pstFile.GetRootFolder(formatType, encryptionType)

	if err != nil {
		fmt.Printf("Failed to get root folder: %s\n", err)
		return
	}

	err = GetSubFolders(pstFile, rootFolder, formatType, encryptionType)

	if err != nil {
		fmt.Printf("Failed to get sub-folders: %s\n", err)
		return
	}

	fmt.Printf("Time: %s", time.Now().Sub(startTime).String())
}

// GetSubFolders is a recursive function which retrieves all sub-folders for the specified folder.
func GetSubFolders(pstFile pst.File, folder pst.Folder, formatType string, encryptionType string) error {
	subFolders, err := pstFile.GetSubFolders(folder, formatType, encryptionType)

	if err != nil {
		return err
	}

	for _, subFolder := range subFolders {
		fmt.Printf("Parsing sub-folder: %s\n", subFolder.DisplayName)

		messages, err := pstFile.GetMessages(subFolder, formatType, encryptionType)

		if err != nil {
			return err
		}

		if len(messages) > 0 {
			fmt.Printf("Found %d messages.\n", len(messages))
		}

		err = GetSubFolders(pstFile, subFolder, formatType, encryptionType)

		if err != nil {
			return err
		}
	}

	return nil
}
