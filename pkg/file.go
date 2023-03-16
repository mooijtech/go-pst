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
	"context"
	"encoding/binary"
	_ "github.com/emersion/go-message/charset"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
	"io"
)

// File represents a PST file.
type File struct {
	Reader         Reader
	FormatType     FormatType
	EncryptionType EncryptionType
	NodeBTree      BTreeStore
	BlockBTree     BTreeStore
	NameToIDMap    *NameToIDMap
}

// Reader defines the file reader used by go-pst to support asynchronous I/O.
// Non-linux systems will fall back to DefaultReader.
// See AsyncReader.
type Reader interface {
	ReadAtAsync(outputBuffer []byte, offset uint64, callback func(err error)) (uint64, error)
	io.ReaderAt // Blocking call.
}

// DefaultReader implements Reader using io.ReaderAt.
type DefaultReader struct {
	reader io.ReaderAt
}

func NewDefaultReader(reader io.ReaderAt) *DefaultReader {
	return &DefaultReader{
		reader: reader,
	}
}

// New is a constructor for creating PST files.
// See also NewAsync.
func New(reader io.ReaderAt) (*File, error) {
	return NewFromReaderWithBTrees(NewDefaultReader(reader), NewBTreeStoreInMemory(), NewBTreeStoreInMemory())
}

// NewFromReaderWithBTrees is a constructor for creating PST files from a reader using the specified b-tree stores.
// Initialization of the b-tree stores will be skipped respectively if not empty.
func NewFromReaderWithBTrees(reader Reader, nodeBTree BTreeStore, blockBTree BTreeStore) (*File, error) {
	pstFile := &File{
		Reader: &DefaultReader{
			reader: reader,
		},
		NodeBTree:  nodeBTree,
		BlockBTree: blockBTree,
	}

	isValidSignature, err := pstFile.IsValidSignature()

	if err != nil {
		return nil, err
	} else if !isValidSignature {
		return nil, ErrFileSignatureInvalid
	}

	formatType, err := pstFile.GetFormatType()

	if err != nil {
		return nil, err
	}

	pstFile.FormatType = formatType

	if _, err := pstFile.GetContentType(); err != nil {
		return nil, err
	}

	encryptionType, err := pstFile.GetEncryptionType()

	if err != nil {
		return nil, err
	}

	pstFile.EncryptionType = encryptionType

	walkGroupContext, walkGroupCancel := context.WithCancelCause(context.Background())
	walkGroup, _ := errgroup.WithContext(walkGroupContext)

	walkGroup.SetLimit(8)

	if pstFile.NodeBTree.Len() == 0 {
		nodeBTreeOffset, err := pstFile.GetNodeBTreeOffset()

		if err != nil {
			return nil, err
		}

		pstFile.WalkAndCreateBTree(nodeBTreeOffset, BTreeTypeNode, pstFile.NodeBTree, walkGroup, walkGroupCancel)

		if err != nil {
			return nil, err
		}
	}

	if pstFile.BlockBTree.Len() == 0 {
		blockBTreeOffset, err := pstFile.GetBlockBTreeOffset()

		if err != nil {
			return nil, err
		}

		pstFile.WalkAndCreateBTree(blockBTreeOffset, BTreeTypeBlock, pstFile.BlockBTree, walkGroup, walkGroupCancel)

		if err != nil {
			return nil, err
		}
	}

	// Wait for walking of the b-trees to complete.
	if err := walkGroup.Wait(); err != nil {
		return nil, eris.Wrap(err, "failed to walk b-trees")
	} else if walkGroupContext.Err() != nil {
		return nil, eris.Wrap(context.Cause(walkGroupContext), "failed to walk and create b-trees")
	}

	nameToIDMap, err := pstFile.GetNameToIDMap()

	if err != nil {
		return nil, err
	}

	pstFile.NameToIDMap = nameToIDMap

	return pstFile, nil
}

// IsValidSignature returns true is the file matches the PFF format signature.
// References "File Header".
func (file *File) IsValidSignature() (bool, error) {
	signature := make([]byte, 4)

	if _, err := file.Reader.ReadAt(signature, 0); err != nil {
		return false, eris.Wrap(err, "failed to read signature")
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

	if _, err := file.Reader.ReadAt(contentType, 8); err != nil {
		return 0, eris.Wrap(err, "failed to get content type")
	}

	if bytes.Equal(contentType, []byte("SM")) {
		return ContentTypePST, nil
	} else if bytes.Equal(contentType, []byte("SO")) {
		return ContentTypeOST, nil
	} else if bytes.Equal(contentType, []byte("AB")) {
		return ContentTypePAB, nil
	} else {
		return 0, ErrContentTypeUnsupported
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

	if _, err := file.Reader.ReadAt(formatType, 10); err != nil {
		return 0, eris.Wrap(err, "failed to read format type")
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
		return 0, ErrFormatTypeUnsupported
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

	if _, err := file.Reader.ReadAt(outputBuffer, offset); err != nil {
		return 0, eris.Wrap(err, "failed to read encryption type")
	}

	switch outputBuffer[0] {
	case 0:
		return EncryptionTypeNone, nil
	case 1:
		return EncryptionTypePermute, nil
	default:
		return 0, ErrEncryptionTypeUnsupported
	}
}

// Cleanup clears the node and block b-trees.
func (file *File) Cleanup() {
	file.NodeBTree.Clear()
	file.BlockBTree.Clear()
}

// ReadAt calls the underlying io.ReaderAt.
func (defaultReader *DefaultReader) ReadAt(outputBuffer []byte, offset int64) (int, error) {
	return defaultReader.reader.ReadAt(outputBuffer, offset)
}

// ReadAtAsync is a fall-back which calls io.ReaderAt.
// See AsyncReader for Linux io_uring support.
func (defaultReader *DefaultReader) ReadAtAsync(outputBuffer []byte, offset uint64, callback func(err error)) (uint64, error) {
	_, err := defaultReader.reader.ReadAt(outputBuffer, int64(offset))

	callback(err)

	return 0, err
}
