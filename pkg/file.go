// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
)

// File represents a PST file.
type File struct {
	Filepath string
}

// New is a constructor for creating PST files.
func New(filePath string) File {
	return File {
		Filepath: filePath,
	}
}

// Read reads the PST file from the given output buffer size and offset to bytes.
func (pstFile *File) Read(outputBufferSize int, offset int) ([]byte, error) {
	inputFile, err := os.Open(pstFile.Filepath)

	if err != nil {
		return nil, err
	}

	_, err = inputFile.Seek(int64(offset), 0)

	if err != nil {
		return nil, err
	}

	inputReader := bufio.NewReader(inputFile)

	outputBuffer := make([]byte, outputBufferSize)

	_, err = inputReader.Read(outputBuffer)

	if err != nil {
		return nil, err
	}

	return outputBuffer, nil
}

// IsValidSignature returns true is the file matches the PFF format signature.
// References "File Header".
func (pstFile *File) IsValidSignature() (bool, error) {
	signature, err := pstFile.Read(4, 0)

	if err != nil {
		return false, err
	}

	return bytes.Equal(signature, []byte("!BDN")), nil
}

// Constants defining the content types.
// References "Content Types".
var (
	ContentTypePST = []byte("SM")
	ContentTypeOST = []byte("SO")
	ContentTypePAB = []byte("AB")
)

// GetContentType returns if the file is a PST, OST or PAB file.
// References "File Header", "Content Types".
func (pstFile *File) GetContentType() ([]byte, error) {
	contentType, err := pstFile.Read(2, 8)

	if err != nil {
		return nil, err
	}

	if bytes.Equal(contentType, ContentTypePST) {
		return ContentTypePST, nil
	} else if bytes.Equal(contentType, ContentTypeOST) {
		return ContentTypeOST, nil
	} else if bytes.Equal(contentType, ContentTypePAB) {
		return ContentTypePAB, nil
	} else {
		return nil, errors.New("unsupported content type")
	}
}

// Constants defining the format types.
// References "Format Types".
const (
	FormatTypeANSI = "ANSI"
	FormatTypeUnicode = "Unicode"
	FormatTypeUnicode4k = "Unicode4k"
)

// GetFormatType returns the format type.
// References "File Header", "Format Types".
func (pstFile *File) GetFormatType() (string, error) {
	formatType, err := pstFile.Read(2, 10)

	if err != nil {
		return "", err
	}

	switch binary.LittleEndian.Uint16(formatType) {
	case 14:
		return FormatTypeANSI, nil
	case 15:
		return FormatTypeANSI, nil
	case 21:
		return FormatTypeUnicode, nil
	case 23:
		return FormatTypeUnicode, nil
	case 36:
		return FormatTypeUnicode4k, nil
	default:
		return "", errors.New("unsupported format type")
	}
}

// Constants defining the encryption types.
// References "Encryption Types".
const (
	EncryptionTypeNone = "None"
	EncryptionTypePermute = "Permute"
	EncryptionTypeCyclic = "Cyclic"
)

// GetEncryptionType returns the encryption type.
// References "The 64-bit header data", "The 32-bit header data", "Encryption Types".
func (pstFile *File) GetEncryptionType(formatType string) (string, error) {
	var encryptionTypeOffset int

	switch formatType {
	case FormatTypeUnicode:
		encryptionTypeOffset = 513
		break
	case FormatTypeUnicode4k:
		encryptionTypeOffset = 513
		break
	case FormatTypeANSI:
		encryptionTypeOffset = 461
		break
	default:
		return "", errors.New("unsupported format type")
	}

	encryptionType, err := pstFile.Read(1, encryptionTypeOffset)

	if err != nil {
		return "", err
	}

	switch binary.LittleEndian.Uint16([]byte{encryptionType[0], 0}) {
	case 0:
		return EncryptionTypeNone, nil
	case 1:
		return EncryptionTypePermute, nil
	case 2:
		return EncryptionTypeCyclic, nil
	default:
		return "", errors.New("unsupported encryption type")
	}
}

// GetRootFolder returns the root folder of the PST file.
func (pstFile *File) GetRootFolder(formatType string) error {
	rootFolderIdentifier := 290

	nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset(formatType)

	if err != nil {
		return err
	}

	rootFolderNode, err := pstFile.FindBTreeNode(nodeBTreeOffset, rootFolderIdentifier, formatType)

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

	rootFolderNodeDataNode, err := pstFile.FindBTreeNode(blockBTreeOffset, rootFolderNodeDataIdentifier, formatType)

	if err != nil {
		return err
	}

	log.Infof("Root folder node data node: %b", rootFolderNodeDataNode.Data)

	err = pstFile.ReadHeapOnNode(rootFolderNodeDataNode, formatType)

	if err != nil {
		return err
	}

	return nil
}