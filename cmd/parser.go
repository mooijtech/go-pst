// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package main

import (
	log "github.com/sirupsen/logrus"
	pst "pst/pkg"
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

	nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

	if err != nil {
		log.Errorf("Failed to get node b-tree offset: %s", err)
		return
	}

	log.Infof("Node b-tree offset: %d", nodeBTreeOffset)

	blockBTreeOffset, err := pstFile.GetBlockBTreeOffset(formatType)

	log.Infof("Block b-tree offset: %d", blockBTreeOffset)

	btreeEntryCount, err := pstFile.GetBTreeEntryCount(nodeBTreeOffset, formatType)

	if err != nil {
		log.Errorf("Failed to get b-tree entry count: %s", err)
		return
	}

	log.Infof("Node b-tree entry count: %d", btreeEntryCount)
}
