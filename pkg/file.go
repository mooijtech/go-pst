// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// File represents a PST file.
type File struct {
	Reader io.ReadSeekCloser

	// Variables which need to be initialized.
	NodeBTree  []BTreeNodeEntry
	BlockBTree []BTreeNodeEntry
	NameToIDMap NameToIDMap
}

// NewFromFile is a constructor for creating PST files from a file path.
func NewFromFile(filePath string) (File, error) {
	inputFile, err := os.Open(filePath)

	if err != nil {
		return File{}, err
	}

	return File{
		Reader: inputFile,
	}, nil
}

// NewFromReader is a constructor for creating PST files from a reader.
func NewFromReader(reader io.ReadSeekCloser) File {
	return File{
		Reader: reader,
	}
}

// Close closes the PST file reader.
func (pstFile *File) Close() error {
	return pstFile.Reader.Close()
}

// Read reads the PST file from the given output buffer size and offset to bytes.
func (pstFile *File) Read(outputBufferSize int, offset int) ([]byte, error) {
	_, err := pstFile.Reader.Seek(int64(offset), 0)

	if err != nil {
		return nil, err
	}

	outputBuffer := make([]byte, outputBufferSize)

	_, err = pstFile.Reader.Read(outputBuffer)

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
	FormatTypeANSI      = "ANSI"
	FormatTypeUnicode   = "Unicode"
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
	EncryptionTypeNone    = "None"
	EncryptionTypePermute = "Permute"
	EncryptionTypeCyclic  = "Cyclic"
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
