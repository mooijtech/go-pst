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

	btreeNodeEntryCount, err := pstFile.GetBTreeNodeEntryCount(nodeBTreeOffset, formatType)

	if err != nil {
		log.Errorf("Failed to get b-tree entry count: %s", err)
		return
	}

	log.Infof("Node b-tree entry count: %d", btreeNodeEntryCount)

	btreeNodeEntrySize, err := pstFile.GetBTreeNodeEntrySize(nodeBTreeOffset, formatType)

	if err != nil {
		log.Infof("Failed to get node b-tree entry size: %s", err)
		return
	}

	log.Infof("Node b-tree entry size: %d", btreeNodeEntrySize)

	btreeNodeLevel, err := pstFile.GetBTreeNodeLevel(nodeBTreeOffset, formatType)

	if err != nil {
		log.Errorf("Failed to get node b-tree level: %s", btreeNodeLevel)
	}

	log.Infof("Node b-tree level: %d", btreeNodeLevel)

	btreeNodeEntries, err := pstFile.GetBTreeNodeEntries(nodeBTreeOffset, formatType)

	if err != nil {
		log.Errorf("Failed to get node b-tree entries: %s", err)
		return
	}

	log.Infof("Node b-tree entries: %d", len(btreeNodeEntries))

	rootFolderNode, err := pstFile.FindBTreeNode(nodeBTreeOffset, 290, formatType)

	if err != nil {
		log.Infof("Failed to find root folder node: %s", err)
		return
	}

	log.Infof("Root folder node: %b", rootFolderNode.Data)

	rootFolderNodeDataIdentifier, err := rootFolderNode.GetDataIdentifier(formatType)

	if err != nil {
		log.Errorf("Failed to get root folder node data identifier: %s", err)
		return
	}

	log.Infof("Root folder node data identifier: %d", rootFolderNodeDataIdentifier)
}
