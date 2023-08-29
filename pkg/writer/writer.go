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
	"encoding/binary"
	"github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
	"hash/crc32"
	"io"
)

// Writer writes PST files.
type Writer struct {
	// Writer represents where the PST file will be written to.
	Writer io.WriterAt
	// WriteOptions defines options used while writing.
	WriteOptions WriteOptions
	// Folders to write.
	Folders []*FolderWriter
}

// NewWriter returns a writer for PST files.
func NewWriter(writer io.WriterAt, writeOptions WriteOptions) *Writer {
	return &Writer{Writer: writer, WriteOptions: writeOptions}
}

// AddFolder adds a pst.Folder to write.
func (writer *Writer) AddFolder(folder *FolderWriter) {
	writer.Folders = append(writer.Folders, folder)
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

// Write writes the PST file.
func (writer *Writer) Write() error {
	if _, err := writer.WriteHeader(); err != nil {
		return eris.Wrap(err, "failed to write header")
	}

	return nil
}

// WriteHeader writes the PST header.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#header-1
func (writer *Writer) WriteHeader() (int, error) {
	var headerSize int

	switch writer.WriteOptions.FormatType {
	case pst.FormatTypeUnicode:
		// 4+4+2+2+2+1+1+4+4+8+8+4+128+8+ROOT+4+128+128+1+1+2+8+4+3+1+32
		// Header + header root
		headerSize = 492 + 72
	case pst.FormatTypeANSI:
		// 4+4+2+2+2+1+1+4+4+4+4+4+128+ROOT+128+128+1+1+2+8+4+3+1+32
		// Header + header root
		headerSize = 472 + 40
	default:
		return 0, pst.ErrFormatTypeUnsupported
	}

	header := make([]byte, headerSize)

	WriteBuffer([]byte("!BDN"), header) // Magic bytes
	//WriteBuffer() // Partial CRC is updated at the end when we have all values.
	WriteBuffer([]byte{0x53, 0x4D}, header) // Magic client

	// File format version
	switch writer.WriteOptions.FormatType {
	case pst.FormatTypeUnicode:
		// MUST be greater than 23 if the file is a Unicode PST file.
		WriteBuffer([]byte{30}, header)
	case pst.FormatTypeANSI:
		// This value MUST be 14 or 15 if the file is an ANSI PST file.
		WriteBuffer([]byte{15}, header)
	default:
		return 0, pst.ErrFormatTypeUnsupported
	}

	WriteBuffer([]byte{19}, header)      // Client file format version.
	WriteBuffer([]byte{1}, header)       // Platform Create
	WriteBuffer([]byte{1}, header)       // Platform Access
	WriteBuffer(make([]byte, 4), header) // Reserved1
	WriteBuffer(make([]byte, 4), header) // Reserved2

	if writer.WriteOptions.FormatType == pst.FormatTypeUnicode {
		// Padding (bidUnused) for Unicode.
		WriteBuffer(make([]byte, 8), header)
	}
	if writer.WriteOptions.FormatType == pst.FormatTypeANSI {
		// Next BID (bidNextB) for ANSI only.
		// go-pst does not read this.
		WriteBuffer(make([]byte, 4), header)
	}

	// Next page BID (bidNextP)
	// go-pst does not read this.
	if writer.WriteOptions.FormatType == pst.FormatTypeUnicode {
		WriteBuffer(make([]byte, 8), header)
	}
	if writer.WriteOptions.FormatType == pst.FormatTypeANSI {
		WriteBuffer(make([]byte, 4), header)
	}

	// This is a monotonically-increasing value that is modified every time the PST file's HEADER structure is modified.
	// The function of this value is to provide a unique value, and to ensure that the HEADER CRCs are different after each header modification.
	WriteBuffer([]byte{1, 3, 3, 7}, header)

	// rgnid
	// go-pst does not read this.
	WriteBuffer(make([]byte, 128), header)

	if writer.WriteOptions.FormatType == pst.FormatTypeUnicode {
		// Unused space; MUST be set to zero. Unicode PST file format only.
		// (qwUnused)
		WriteBuffer(make([]byte, 8), header)
	}

	// Header root
	if err := writer.WriteHeaderRoot(header); err != nil {
		return 0, eris.Wrap(err, "failed to write header root")
	}

	// Unused alignment bytes; MUST be set to zero.
	// Unicode PST file format only.
	if writer.WriteOptions.FormatType == pst.FormatTypeUnicode {
		WriteBuffer(make([]byte, 4), header)
	}

	WriteBuffer(make([]byte, 128), header) // Deprecated FMap (rgbFM).
	WriteBuffer(make([]byte, 128), header) // Deprecated FMap (rgbFP).

	for i := 0; i < 128+128; i++ {
		// Fill FMap.
		copy(header[len(header):len(header)+i], []byte{255})
	}

	WriteBuffer([]byte{128}, header)                                      // bSentinel
	WriteBuffer([]byte{byte(writer.WriteOptions.EncryptionType)}, header) // Encryption. Indicates how the data within the PST file is encoded. (bCryptMethod)
	WriteBuffer(make([]byte, 2), header)                                  // rgbReserved

	if writer.WriteOptions.FormatType == pst.FormatTypeUnicode {
		// Next BID. go-pst does not read this value (bidNextB)
		WriteBuffer(make([]byte, 8), header)

		// The 32-bit CRC value of the 516 bytes of data starting from wMagicClient to bidNextB, inclusive.
		// Unicode PST file format only. (dwCRCFull)
		WriteBuffer([]byte{byte(crc32.ChecksumIEEE(header[4 : 4+516]))}, header)
	}

	if writer.WriteOptions.FormatType == pst.FormatTypeANSI {
		WriteBuffer(make([]byte, 8), header) // ullReserved
		WriteBuffer(make([]byte, 4), header) // dwReserved
	}

	WriteBuffer(make([]byte, 3), header)  // rgbReserved2
	WriteBuffer(make([]byte, 1), header)  // bReserved
	WriteBuffer(make([]byte, 32), header) // rgbReserved3

	// Update first partial CRC
	copy(header[4:4+4], []byte{byte(crc32.ChecksumIEEE(header[10 : 10+471]))})

	if len(header) != headerSize {
		return 0, eris.New("header size mismatch")
	}

	return writer.Writer.WriteAt(header, 0)
}

// WriteHeaderRoot writes the header root.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#root
func (writer *Writer) WriteHeaderRoot(outputBuffer []byte) error {
	var headerSize int

	switch writer.WriteOptions.FormatType {
	case pst.FormatTypeUnicode:
		// 4+8+8+8+8+16+16+1+1+2
		headerSize = 72
	case pst.FormatTypeANSI:
		// 4+4+4+4+4+8+8+1+1+2
		headerSize = 40
	default:
		return pst.ErrFormatTypeUnsupported
	}

	header := make([]byte, headerSize)

	WriteBuffer(make([]byte, 4), header) // dwReserved

	switch writer.WriteOptions.FormatType {
	case pst.FormatTypeUnicode:
		// The size of the PST file, in bytes. (ibFileEof) TODO
		WriteBuffer(make([]byte, 8), header)
		// An IB structure (section 2.2.2.3) that contains the absolute file offset to the last AMap page of the PST file.
		WriteBuffer(make([]byte, 8), header)
		// The total free space in all AMaps, combined.
		WriteBuffer(make([]byte, 8), header)
		// The total free space in all PMaps, combined. Because the PMap is deprecated, this value SHOULD be zero.
		WriteBuffer(make([]byte, 8), header)
		// A BREF structure (section 2.2.2.4) that references the root page of the Node BTree (NBT).
		WriteBuffer(append(binary.LittleEndian.AppendUint64([]byte{}, uint64(writer.RootNodeBTree)), make([]byte, 8)...), header)
		// A BREF structure that references the root page of the Block BTree (BBT).
		WriteBuffer(append(binary.LittleEndian.AppendUint64([]byte{}, uint64(writer.RootBlockBTree)), make([]byte, 8)...), header)
		// Indicates whether the AMaps in this PST file are valid (0 = INVALID_AMAP).
		WriteBuffer([]byte{0}, header)
	case pst.FormatTypeANSI:
		// The size of the PST file, in bytes. (ibFileEof)
		WriteBuffer(make([]byte, 4), header)
		// An IB structure (section 2.2.2.3) that contains the absolute file offset to the last AMap page of the PST file.
		WriteBuffer(make([]byte, 4), header)
		// The total free space in all AMaps, combined.
		WriteBuffer(make([]byte, 4), header)
		// The total free space in all PMaps, combined. Because the PMap is deprecated, this value SHOULD be zero.
		WriteBuffer(make([]byte, 4), header)
		// A BREF structure (section 2.2.2.4) that references the root page of the Node BTree (NBT).
		binary.LittleEndian.PutUint64(header[len(header):len(header)+8], uint64(writer.RootNodeBTree))
		// A BREF structure that references the root page of the Block BTree (BBT).
		binary.LittleEndian.PutUint64(header[len(header):len(header)+8], uint64(writer.RootBlockBTree))
		// Indicates whether the AMaps in this PST file are valid (0 = INVALID_AMAP).
		WriteBuffer([]byte{0}, header)
	default:
		return pst.ErrFormatTypeUnsupported
	}

	WriteBuffer(make([]byte, 1), header) // bReserved
	WriteBuffer(make([]byte, 2), header) // wReserved

	if len(header) != headerSize {
		return eris.New("header root size mismatch")
	}

	copy(outputBuffer, header)

	return nil
}
