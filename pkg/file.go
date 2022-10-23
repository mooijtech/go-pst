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
	"bytes"
	"encoding/binary"
	"io"
	"os"

	"github.com/pkg/errors"
)

// File represents a PST file.
type File struct {
	Reader         io.ReadSeekCloser
	FormatType     FormatType
	EncryptionType EncryptionType
	NodeBTree      BTreeStore
	BlockBTree     BTreeStore
	NameToIDMap    *NameToIDMap
}

// NewFromFile is a constructor for creating PST files from a file path.
func NewFromFile(name string) (*File, error) {
	inputFile, err := os.Open(name)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return NewFromReader(inputFile)
}

// NewFromReader is a constructor for creating PST files from an io.ReadSeekCloser.
func NewFromReader(reader io.ReadSeekCloser) (*File, error) {
	return NewFromReaderWithBTrees(reader, NewBTreeStoreInMemory(), NewBTreeStoreInMemory())
}

// NewFromFileWithBTrees is a constructor for creating PST files from a file path using the b-tree stores.
// Initialization of the b-tree stores will be skipped respectively if not empty.
func NewFromFileWithBTrees(filePath string, nodeBTree BTreeStore, blockBTree BTreeStore) (*File, error) {
	inputFile, err := os.Open(filePath)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return NewFromReaderWithBTrees(inputFile, nodeBTree, blockBTree)
}

// NewFromReaderWithBTrees is a constructor for creating PST files from a reader using the specified b-tree stores.
// Initialization of the b-tree stores will be skipped respectively if not empty.
func NewFromReaderWithBTrees(reader io.ReadSeekCloser, nodeBTree BTreeStore, blockBTree BTreeStore) (*File, error) {
	pstFile := &File{
		Reader:     reader,
		NodeBTree:  nodeBTree,
		BlockBTree: blockBTree,
	}

	isValidSignature, err := pstFile.IsValidSignature()

	if err != nil {
		return nil, errors.WithStack(err)
	} else if !isValidSignature {
		return nil, errors.WithStack(ErrFileSignatureInvalid)
	}

	formatType, err := pstFile.GetFormatType()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	pstFile.FormatType = formatType

	if _, err := pstFile.GetContentType(); err != nil {
		return nil, errors.WithStack(err)
	}

	encryptionType, err := pstFile.GetEncryptionType()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	pstFile.EncryptionType = encryptionType

	if pstFile.NodeBTree.Len() == 0 {
		nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset()

		if err != nil {
			return nil, errors.WithStack(err)
		}

		err = pstFile.WalkAndCreateBTree(nodeBTreeOffset, BTreeTypeNode, pstFile.NodeBTree)

		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if pstFile.BlockBTree.Len() == 0 {
		blockBTreeOffset, err := pstFile.GetBlockBTreeOffset()

		if err != nil {
			return nil, errors.WithStack(err)
		}

		err = pstFile.WalkAndCreateBTree(blockBTreeOffset, BTreeTypeBlock, pstFile.BlockBTree)

		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	nameToIDMap, err := pstFile.GetNameToIDMap()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	pstFile.NameToIDMap = nameToIDMap

	return pstFile, nil
}

// IsValidSignature returns true is the file matches the PFF format signature.
// References "File Header".
func (file *File) IsValidSignature() (bool, error) {
	signature := make([]byte, 4)

	if _, err := file.ReadAt(signature, 0); err != nil {
		return false, errors.WithStack(err)
	}

	return bytes.Equal(signature, []byte("!BDN")), nil
}

// ContentType represents a PST, OST or PAB file.
type ContentType uint8

// Constants defining the content types.
// References "Content Types".
const (
	ContentTypePST ContentType = iota
	ContentTypeOST
	ContentTypePAB
)

// GetContentType returns if the file is a PST, OST or PAB file.
// References "File Header", "Content Types".
func (file *File) GetContentType() (ContentType, error) {
	contentType := make([]byte, 2)

	if _, err := file.ReadAt(contentType, 8); err != nil {
		return 0, errors.WithStack(err)
	}

	if bytes.Equal(contentType, []byte("SM")) {
		return ContentTypePST, nil
	} else if bytes.Equal(contentType, []byte("SO")) {
		return ContentTypeOST, nil
	} else if bytes.Equal(contentType, []byte("AB")) {
		return ContentTypePAB, nil
	} else {
		return 0, errors.WithStack(ErrContentTypeUnsupported)
	}
}

// FormatType represents a Unicode or ANSI format type.
type FormatType uint8

// Constants defining the format types.
// References "Format Types".
const (
	FormatTypeANSI FormatType = iota
	FormatTypeUnicode
	FormatTypeUnicode4k
)

// GetFormatType returns the format type.
// References "File Header", "Format Types".
func (file *File) GetFormatType() (FormatType, error) {
	formatType := make([]byte, 2)

	if _, err := file.ReadAt(formatType, 10); err != nil {
		return 0, errors.WithStack(err)
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
		return 0, errors.WithStack(ErrFormatTypeUnsupported)
	}
}

type EncryptionType uint8

// Constants defining the encryption types.
// References "Encryption Types".
const (
	EncryptionTypeNone    EncryptionType = 0
	EncryptionTypePermute EncryptionType = 1
	//EncryptionTypeCyclic  EncryptionType = 2 // Not implemented currently.
)

// GetEncryptionType returns the encryption type.
// References "The 64-bit header data", "The 32-bit header data", "Encryption Types".
func (file *File) GetEncryptionType() (EncryptionType, error) {
	outputBuffer := make([]byte, 1)
	var offset int64

	switch file.FormatType {
	case FormatTypeANSI:
		offset = 461
	default:
		offset = 513
	}

	if _, err := file.ReadAt(outputBuffer, offset); err != nil {
		return 0, errors.WithStack(err)
	}

	switch outputBuffer[0] {
	case 0:
		return EncryptionTypeNone, nil
	case 1:
		return EncryptionTypePermute, nil
	default:
		return 0, errors.WithStack(ErrEncryptionTypeUnsupported)
	}
}

// Close closes the PST file.
func (file *File) Close() error {
	file.NodeBTree.Clear()
	file.BlockBTree.Clear()

	return file.Reader.Close()
}

// ReadAt implements io.ReadAt for the PST file.
func (file *File) ReadAt(outputBuffer []byte, offset int64) (int, error) {
	if _, err := file.Reader.Seek(offset, 0); err != nil {
		return 0, errors.WithStack(err)
	}

	written, err := file.Reader.Read(outputBuffer)

	if err != nil {
		return written, errors.WithStack(err)
	}

	return written, nil
}
