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

// Package writer implements writing PST files.
package writer

import (
	"bytes"
	"github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
	"hash/crc32"
	"io"
)

// Writer writes PST files.
type Writer struct {
	// WriteOptions defines options used while writing.
	WriteOptions WriteOptions
	// Folders to write.
	Folders []*FolderWriter
}

// NewWriter returns a writer for PST files.
func NewWriter(writeOptions WriteOptions) *Writer {
	return &Writer{WriteOptions: writeOptions}
}

// AddFolder adds a pst.Folder to write (used by WriteTo).
func (pstWriter *Writer) AddFolder(folder *FolderWriter) {
	pstWriter.Folders = append(pstWriter.Folders, folder)
}

// WriteOptions defines the options used during writing.
type WriteOptions struct {
	// FormatType represents ANSI or Unicode.
	FormatType pst.FormatType
	// EncryptionType represents the encryption type.
	EncryptionType pst.EncryptionType
}

// NewWriteOptions creates a new WriteOptions used during writing PST files.
func NewWriteOptions(formatType pst.FormatType, encryptionType pst.EncryptionType) WriteOptions {
	return WriteOptions{
		FormatType:     formatType,
		EncryptionType: encryptionType,
	}
}

// WriteTo writes the PST file.
func (pstWriter *Writer) WriteTo(writer io.Writer) (int64, error) {
	var totalSize int64

	// Write folders.
	for _, folderWriter := range pstWriter.Folders {
		written, err := folderWriter.WriteTo(writer)

		if err != nil {
			return 0, eris.Wrap(err, "failed to write folder")
		}

		totalSize += written
	}

	// Write PST header.
	// TODO - Root B-Trees
	headerWrittenSize, err := pstWriter.WriteHeader(writer, totalSize, 0, 0)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write header")
	}

	return totalSize + headerWrittenSize, nil
}

// WriteHeader writes the PST header.
// totalSize is the total size of the PST file excluding the header.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#header-1
func (pstWriter *Writer) WriteHeader(writer io.Writer, totalSize int64, rootNodeBTree pst.Identifier, rootBlockBTree pst.Identifier) (int64, error) {
	var headerSize int

	switch pstWriter.WriteOptions.FormatType {
	case pst.FormatTypeUnicode:
		// 4+4+2+2+2+1+1+4+4+8+8+4+128+8+ROOT+4+128+128+1+1+2+8+4+3+1+32
		// Header + header root
		headerSize = 492 + 72
	case pst.FormatTypeANSI:
		// 4+4+2+2+2+1+1+4+4+4+4+4+128+ROOT+128+128+1+1+2+8+4+3+1+32
		// Header + header root
		headerSize = 472 + 40
	default:
		panic(pst.ErrFormatTypeUnsupported)
	}

	header := bytes.NewBuffer(make([]byte, headerSize))

	header.Write([]byte("!BDN")) // Magic bytes
	//header.Write() // Partial CRC is updated at the end when we have all values.
	header.Write([]byte{0x53, 0x4D}) // Magic client

	// File format version
	switch pstWriter.WriteOptions.FormatType {
	case pst.FormatTypeUnicode:
		// MUST be greater than 23 if the file is a Unicode PST file.
		header.Write([]byte{30})
	case pst.FormatTypeANSI:
		// This value MUST be 14 or 15 if the file is an ANSI PST file.
		header.Write([]byte{15})
	default:
		panic(pst.ErrFormatTypeUnsupported)
	}

	header.Write([]byte{19})      // Client file format version.
	header.Write([]byte{1})       // Platform Create
	header.Write([]byte{1})       // Platform Access
	header.Write(make([]byte, 4)) // Reserved1
	header.Write(make([]byte, 4)) // Reserved2

	if pstWriter.WriteOptions.FormatType == pst.FormatTypeUnicode {
		// Padding (bidUnused) for Unicode.
		header.Write(make([]byte, 8))
	}
	if pstWriter.WriteOptions.FormatType == pst.FormatTypeANSI {
		// Next BID (bidNextB) for ANSI only.
		// go-pst does not read this.
		header.Write(make([]byte, 4))
	}

	// Next page BID (bidNextP)
	// go-pst does not read this.
	if pstWriter.WriteOptions.FormatType == pst.FormatTypeUnicode {
		header.Write(make([]byte, 8))
	}
	if pstWriter.WriteOptions.FormatType == pst.FormatTypeANSI {
		header.Write(make([]byte, 4))
	}

	// This is a monotonically-increasing value that is modified every time the PST file's HEADER structure is modified.
	// The function of this value is to provide a unique value, and to ensure that the HEADER CRCs are different after each header modification.
	header.Write([]byte{1, 3, 3, 7})

	// rgnid
	// go-pst does not read this.
	header.Write(make([]byte, 128))

	if pstWriter.WriteOptions.FormatType == pst.FormatTypeUnicode {
		// Unused space; MUST be set to zero. Unicode PST file format only.
		// (qwUnused)
		header.Write(make([]byte, 8))
	}

	// Header root
	if _, err := pstWriter.WriteHeaderRoot(header, totalSize, rootNodeBTree, rootBlockBTree); err != nil {
		return 0, eris.Wrap(err, "failed to write header root")
	}

	// Unused alignment bytes; MUST be set to zero.
	// Unicode PST file format only.
	if pstWriter.WriteOptions.FormatType == pst.FormatTypeUnicode {
		header.Write(make([]byte, 4))
	}

	for i := 0; i < 128+128; i++ {
		// Fill both FMap (deprecated).
		header.Write([]byte{255})
	}

	// bSentinel
	header.Write([]byte{128})
	// Encryption. Indicates how the data within the PST file is encoded. (bCryptMethod)
	header.Write([]byte{byte(pstWriter.WriteOptions.EncryptionType)})
	// rgbReserved
	header.Write(make([]byte, 2))

	if pstWriter.WriteOptions.FormatType == pst.FormatTypeUnicode {
		// Next BID. go-pst does not read this value (bidNextB)
		header.Write(make([]byte, 8))

		// The 32-bit CRC value of the 516 bytes of data starting from wMagicClient to bidNextB, inclusive.
		// Unicode PST file format only. (dwCRCFull)
		header.Write([]byte{byte(crc32.ChecksumIEEE(header.Bytes()[4 : 4+516]))})
	}

	if pstWriter.WriteOptions.FormatType == pst.FormatTypeANSI {
		header.Write(make([]byte, 8)) // ullReserved
		header.Write(make([]byte, 4)) // dwReserved
	}

	header.Write(make([]byte, 3))  // rgbReserved2
	header.Write(make([]byte, 1))  // bReserved
	header.Write(make([]byte, 32)) // rgbReserved3

	// Update first partial CRC
	copy(header.Bytes()[4:4+4], []byte{byte(crc32.ChecksumIEEE(header.Bytes()[10 : 10+471]))})

	if header.Len() != headerSize {
		return 0, eris.New("header size mismatch")
	}

	return header.WriteTo(writer)
}

// WriteHeaderRoot writes the header root.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#root
func (pstWriter *Writer) WriteHeaderRoot(writer io.Writer, totalSize int64, rootNodeBTree pst.Identifier, rootBlockBTree pst.Identifier) (int64, error) {
	var headerSize int

	switch pstWriter.WriteOptions.FormatType {
	case pst.FormatTypeUnicode:
		// 4+8+8+8+8+16+16+1+1+2
		headerSize = 72
	case pst.FormatTypeANSI:
		// 4+4+4+4+4+8+8+1+1+2
		headerSize = 40
	default:
		panic(pst.ErrFormatTypeUnsupported)
	}

	header := bytes.NewBuffer(make([]byte, headerSize))

	header.Write(make([]byte, 4)) // dwReserved

	switch pstWriter.WriteOptions.FormatType {
	case pst.FormatTypeUnicode:
		// The size of the PST file, in bytes. (ibFileEof)
		header.Write(GetUint64(uint64(totalSize)))
		// An IB structure (section 2.2.2.3) that contains the absolute file offset to the last AMap page of the PST file.
		header.Write(make([]byte, 8))
		// The total free space in all AMaps, combined.
		header.Write(make([]byte, 8))
		// The total free space in all PMaps, combined. Because the PMap is deprecated, this value SHOULD be zero.
		header.Write(make([]byte, 8))
		// A BREF structure (section 2.2.2.4) that references the root page of the Node BTree (NBT).
		header.Write(append(GetUint64(uint64(rootNodeBTree)), make([]byte, 8)...))
		// A BREF structure that references the root page of the Block BTree (BBT).
		header.Write(append(GetUint64(uint64(rootBlockBTree)), make([]byte, 8)...))
		// Indicates whether the AMaps in this PST file are valid (0 = INVALID_AMAP).
		header.Write([]byte{0})
	case pst.FormatTypeANSI:
		// The size of the PST file, in bytes. (ibFileEof)
		header.Write(GetUint32(uint32(totalSize)))
		// An IB structure (section 2.2.2.3) that contains the absolute file offset to the last AMap page of the PST file.
		header.Write(make([]byte, 4))
		// The total free space in all AMaps, combined.
		header.Write(make([]byte, 4))
		// The total free space in all PMaps, combined. Because the PMap is deprecated, this value SHOULD be zero.
		header.Write(make([]byte, 4))
		// A BREF structure (section 2.2.2.4) that references the root page of the Node BTree (NBT).
		header.Write(GetUint64(uint64(rootNodeBTree)))
		// A BREF structure that references the root page of the Block BTree (BBT).
		header.Write(GetUint64(uint64(rootBlockBTree)))
		// Indicates whether the AMaps in this PST file are valid (0 = INVALID_AMAP).
		header.Write([]byte{0})
	default:
		panic(pst.ErrFormatTypeUnsupported)
	}

	header.Write(make([]byte, 1)) // bReserved
	header.Write(make([]byte, 2)) // wReserved

	if header.Len() != headerSize {
		panic("header root size mismatch")
	}

	return header.WriteTo(writer)
}
