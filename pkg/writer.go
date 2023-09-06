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

package pst

import (
	"bytes"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
	"hash/crc32"
	"io"
)

// Writer writes PST files.
type Writer struct {
	// Writer represents the writer.
	Writer io.WriteSeeker
	// WriteGroup represents the writers running in Goroutines.
	WriteGroup *errgroup.Group
	// WriteOptions represents options used while writing.
	WriteOptions WriteOptions
	// FolderWriter represents the writer for folders.
	FolderWriter *FolderWriter
	// FolderWriteCallback represents the callback which is called after a folder is written.
	FolderWriteCallback chan int64
}

// NewWriter returns a writer for PST files.
func NewWriter(writer io.WriteSeeker, writeGroup *errgroup.Group, writeOptions WriteOptions) (*Writer, error) {
	folderWriteCallback := make(chan int64)
	folderWriter, err := NewFolderWriter(writer, writeGroup, writeOptions.FormatType, folderWriteCallback)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create folder writer")
	}

	pstWriter := &Writer{
		Writer:              writer,
		WriteGroup:          writeGroup,
		WriteOptions:        writeOptions,
		FolderWriter:        folderWriter,
		FolderWriteCallback: folderWriteCallback,
	}

	return pstWriter, nil
}

// WriteOptions defines the options used during writing.
type WriteOptions struct {
	// FormatType represents ANSI or Unicode.
	FormatType FormatType
	// EncryptionType represents the encryption type.
	EncryptionType EncryptionType
}

// NewWriteOptions creates a new WriteOptions used during writing PST files.
func NewWriteOptions(formatType FormatType, encryptionType EncryptionType) WriteOptions {
	return WriteOptions{
		FormatType:     formatType,
		EncryptionType: encryptionType,
	}
}

// AddFolders adds the folders to the FolderWriter write queue.
// See AddRootFolder.
func (pstWriter *Writer) AddFolders(folders ...*Folder) {
	pstWriter.FolderWriter.AddFolders(folders...)
}

// WriteTo writes the PST file.
// WriteTo follows the path to root folder (fixed pst.Identifier, pst.IdentifierRootFolder) then to the pst.TableContext of the root folder.
// Once there, we can get the child folders ([]pst.Identifier, see FolderWriter), each folder can contain messages (see MessageWriter).
// Each message uses the pst.BTreeOnHeapHeader to construct a pst.HeapOnNode (this is where the data is).
//
// Extending the pst.HeapOnNode (where the data is) we can also use Local Descriptors (extend where this data is):
// pst.LocalDescriptor (see LocalDescriptorsWriter) are B-Tree nodes pointing to other B-Tree nodes.
// These local descriptors also have the pst.HeapOnNode structure which can be built upon (explained below).
// Local Descriptors are used to store more data in the pst.HeapOnNode structure (B-Tree with the nodes containing the data).
// XBlocks and XXBlocks include an array of []pst.Identifier pointing to B-Tree nodes, it is the format used to store data (see BlockWriter).
// This structure is used by the Local Descriptors.
//
// Each pst.HeapOnNode can contain either a pst.TableContext or pst.PropertyContext:
// pst.TableContext (see TableContextWriter):
// The pst.TableContext contains a Row Matrix structure to store data, used by folders (to find data such as the folder identifiers ([]pst.Identifier)).
// The pst.TableContext is column structured with data exceeding 8 bytes moving to different B-Tree nodes:
// pst.HeapOnNode which is <= 3580 bytes.
// pst.LocalDescriptor which is > 3580 bytes.
// pst.PropertyContext (see PropertyContextWriter):
// The pst.PropertyContext contains a list of properties ([]pst.Property) of the message, we can write this with PropertyWriter.
//
// Combining these structures we make up a PST file to write.
// Each writer communicates with Go channels and Goroutines.
func (pstWriter *Writer) WriteTo(writer io.Writer) (int64, error) {
	var totalSize int64

	// Wait for all folders to be written.
	// "Share memory by communicating; don't communicate by sharing memory."
	for folderWrittenSize := range pstWriter.FolderWriter.FolderWriteCallback {
		totalSize += folderWrittenSize
	}

	// Wait for channels to finish.
	if err := pstWriter.WriteGroup.Wait(); err != nil {
		return 0, eris.Wrap(err, "writer failed")
	}

	// Write PST header.
	headerWrittenSize, err := pstWriter.WriteHeader(writer, totalSize, IdentifierRootFolder, IdentifierRootFolder)

	if err != nil {
		return 0, eris.Wrap(err, "failed to write header")
	}

	return headerWrittenSize + totalSize, nil
}

// WriteHeader writes the PST header.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#header-1
func (pstWriter *Writer) WriteHeader(writer io.Writer, totalSize int64, rootNodeBTree Identifier, rootBlockBTree Identifier) (int64, error) {
	var headerSize int

	switch pstWriter.WriteOptions.FormatType {
	case FormatTypeUnicode4k, FormatTypeUnicode:
		// TODO - Where is the documentation for Unicode4k?
		// 4+4+2+2+2+1+1+4+4+8+8+4+128+8+ROOT+4+128+128+1+1+2+8+4+3+1+32
		// Header + header root
		headerSize = 492 + 72
	case FormatTypeANSI:
		// 4+4+2+2+2+1+1+4+4+4+4+4+128+ROOT+128+128+1+1+2+8+4+3+1+32
		// Header + header root
		headerSize = 472 + 40
	default:
		return 0, ErrFormatTypeUnsupported
	}

	header := bytes.NewBuffer(make([]byte, headerSize))

	header.Write([]byte("!BDN")) // Magic bytes
	//header.Write() // Partial CRC is updated at the end when we have all values.
	header.Write([]byte{0x53, 0x4D}) // Magic client

	// File format version
	switch pstWriter.WriteOptions.FormatType {
	case FormatTypeUnicode4k, FormatTypeUnicode:
		// TODO - Where is the documentation for Unicode4k?
		// MUST be greater than 23 if the file is a Unicode PST file.
		header.Write([]byte{30})
	case FormatTypeANSI:
		// This value MUST be 14 or 15 if the file is an ANSI PST file.
		header.Write([]byte{15})
	default:
		return 0, ErrFormatTypeUnsupported
	}

	// Client file format version.
	header.Write([]byte{19})
	// Platform Create
	header.Write([]byte{1})
	// Platform Access
	header.Write([]byte{1})
	// Reserved1
	header.Write(make([]byte, 4))
	// Reserved2
	header.Write(make([]byte, 4))

	// Padding (bidUnused) for Unicode.
	if pstWriter.WriteOptions.FormatType == FormatTypeUnicode || pstWriter.WriteOptions.FormatType == FormatTypeUnicode4k {
		header.Write(make([]byte, 8))
	}

	// Next BID (bidNextB) for ANSI only.
	// go-pst does not read this.
	if pstWriter.WriteOptions.FormatType == FormatTypeANSI {
		header.Write(make([]byte, 4))
	}

	// Next identifier (bidNextP).
	// This can be used to increment a value for generating identifiers used by B-Tree nodes.
	// I assume it's faster to generate an identifier with crypto/rand instead of having to read and update this value.
	// go-pst does not read this.
	if pstWriter.WriteOptions.FormatType == FormatTypeUnicode4k || pstWriter.WriteOptions.FormatType == FormatTypeUnicode {
		header.Write(make([]byte, 8))
	} else if pstWriter.WriteOptions.FormatType == FormatTypeANSI {
		header.Write(make([]byte, 4))
	}

	// This is a monotonically-increasing value that is modified every time the PST header structure is modified.
	// The function of this value is to provide a unique value which ensures that the header CRCs are different after each header modification.
	header.Write([]byte{1, 3, 3, 7})

	// Special Internal B-Tree nodes.
	// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#special-internal-nids
	if _, err := pstWriter.WriteInternalBTreeNodes(header); err != nil {
		return 0, eris.Wrap(err, "failed to write internal B-Tree nodes")
	}

	// Unused space; MUST be set to zero. Unicode PST file format only.
	// (qwUnused)
	if pstWriter.WriteOptions.FormatType == FormatTypeUnicode4k || pstWriter.WriteOptions.FormatType == FormatTypeUnicode {
		header.Write(make([]byte, 8))
	}

	// Header root
	if _, err := pstWriter.WriteHeaderRoot(header, totalSize, rootNodeBTree, rootBlockBTree); err != nil {
		return 0, eris.Wrap(err, "failed to write header root")
	}

	// Unused alignment bytes; MUST be set to zero.
	// Unicode PST file format only.
	if pstWriter.WriteOptions.FormatType == FormatTypeUnicode4k || pstWriter.WriteOptions.FormatType == FormatTypeUnicode {
		header.Write(make([]byte, 4))
	}

	// Fill both FMaps (deprecated).
	for i := 0; i < 128+128; i++ {
		header.Write([]byte{255})
	}

	// bSentinel
	header.Write([]byte{128})
	// Encryption. Indicates how the data within the PST file is encoded. (bCryptMethod)
	header.WriteByte(byte(pstWriter.WriteOptions.EncryptionType))
	// rgbReserved
	header.Write(make([]byte, 2))

	if pstWriter.WriteOptions.FormatType == FormatTypeUnicode4k || pstWriter.WriteOptions.FormatType == FormatTypeUnicode {
		// Next BID.
		// go-pst does not read this value (bidNextB)
		header.Write(make([]byte, 8))

		// The 32-bit CRC value of the 516 bytes of data starting from wMagicClient to bidNextB, inclusive.
		// Unicode PST file format only. (dwCRCFull)
		// TODO - Check matches the CRC algorithm used by Microsoft.
		header.Write([]byte{byte(crc32.ChecksumIEEE(header.Bytes()[4 : 4+516]))})
	}

	if pstWriter.WriteOptions.FormatType == FormatTypeANSI {
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

// WriteInternalBTreeNodes writes the special internal B-Tree nodes.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#special-internal-nids
func (pstWriter *Writer) WriteInternalBTreeNodes(writer io.Writer) (int64, error) {
	internalIdentifiersBuffer := bytes.NewBuffer(make([]byte, 128))
	formatType := pstWriter.WriteOptions.FormatType

	internalIdentifiersBuffer.Write(IdentifierMessageStore.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierNameToIDMap.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierNormalFolderTemplate.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierSearchFolderTemplate.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierRootFolder.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierSearchManagementQueue.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierSearchActivityList.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierReserved1.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierSearchDomainObject.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierSearchGathererQueue.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierSearchGathererDescriptor.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierReserved2.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierReserved3.Bytes(formatType))
	internalIdentifiersBuffer.Write(IdentifierSearchGathererFolderQueue.Bytes(formatType))

	return internalIdentifiersBuffer.WriteTo(writer)
}

// WriteHeaderRoot writes the header root.
// References https://github.com/mooijtech/go-pst/blob/main/docs/README.md#root
func (pstWriter *Writer) WriteHeaderRoot(writer io.Writer, totalSize int64, rootNodeBTree Identifier, rootBlockBTree Identifier) (int64, error) {
	var headerSize int

	switch pstWriter.WriteOptions.FormatType {
	case FormatTypeUnicode:
		// 4+8+8+8+8+16+16+1+1+2
		headerSize = 72
	case FormatTypeANSI:
		// 4+4+4+4+4+8+8+1+1+2
		headerSize = 40
	default:
		return 0, ErrFormatTypeUnsupported
	}

	header := bytes.NewBuffer(make([]byte, headerSize))

	header.Write(make([]byte, 4)) // dwReserved

	switch pstWriter.WriteOptions.FormatType {
	case FormatTypeUnicode:
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
	case FormatTypeANSI:
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
		return 0, ErrFormatTypeUnsupported
	}

	header.Write(make([]byte, 1)) // bReserved
	header.Write(make([]byte, 2)) // wReserved

	if header.Len() != headerSize {
		return 0, eris.New("header root size mismatch")
	}

	return header.WriteTo(writer)
}
