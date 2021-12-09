// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
)

// HeapOnNodeInputStream represents a node input stream for a Heap-on-Node.
type HeapOnNodeInputStream struct {
	File           *File
	FormatType     string
	EncryptionType string
	FileOffset     int
	StartOffset    int
	Size           int
	Blocks         []BTreeNodeEntry
}

// NewHeapOnNodeInputStream creates a node input stream from the Heap-on-Node.
func (pstFile *File) NewHeapOnNodeInputStream(btreeNodeEntry BTreeNodeEntry, formatType string, encryptionType string) (HeapOnNodeInputStream, error) {
	nodeEntryHeapOnNodeOffset, err := btreeNodeEntry.GetFileOffset(false, formatType)

	if err != nil {
		return HeapOnNodeInputStream{}, err
	}

	nodeEntryHeapOnNodeSize, err := btreeNodeEntry.GetSize(formatType)

	if err != nil {
		return HeapOnNodeInputStream{}, err
	}

	nodeEntryIdentifier, err := btreeNodeEntry.GetIdentifier(formatType)

	if err != nil {
		return HeapOnNodeInputStream{}, err
	}

	// Internal identifiers have blocks (XBlock or XXBlock).
	// This is a list of block identifiers that point to block b-tree entries (where the data is).
	isInternal := nodeEntryIdentifier&0x02 != 0

	if isInternal {
		blocks, err := pstFile.GetBlocks(nodeEntryHeapOnNodeOffset, formatType)

		if err != nil {
			return HeapOnNodeInputStream{}, err
		}

		blocksTotalSize, err := pstFile.GetBlocksTotalSize(nodeEntryHeapOnNodeOffset)

		return HeapOnNodeInputStream{
			File:           pstFile,
			FormatType:     formatType,
			EncryptionType: encryptionType,
			FileOffset:     nodeEntryHeapOnNodeOffset,
			Size:           blocksTotalSize,
			Blocks:         blocks,
		}, nil
	}

	return HeapOnNodeInputStream{
		File:           pstFile,
		FormatType:     formatType,
		EncryptionType: encryptionType,
		FileOffset:     nodeEntryHeapOnNodeOffset,
		Size:           nodeEntryHeapOnNodeSize,
	}, nil
}

// NewHeapOnNodeInputStreamFromHNID returns the offsets from the allocation table of the given HID.
func (pstFile *File) NewHeapOnNodeInputStreamFromHNID(hnid int, heapOnNode HeapOnNode, localDescriptors []LocalDescriptor, formatType string, encryptionType string) (HeapOnNodeInputStream, error) {
	if len(localDescriptors) > 0 {
		localDescriptor, err := FindLocalDescriptor(localDescriptors, hnid, formatType)

		if err == nil {
			// Found the local descriptor for this identifier.
			localDescriptorHeapOnNode, err := pstFile.NewHeapOnNodeFromLocalDescriptor(localDescriptor, formatType, encryptionType)

			if err != nil {
				return HeapOnNodeInputStream{}, err
			}

			return localDescriptorHeapOnNode.InputStream, nil
		}
	}

	if (hnid & 0x1F) != 0 {
		return HeapOnNodeInputStream{}, errors.New("external node")
	}

	hidBlockIndex := hnid >> 16
	blockOffset := 0

	if hidBlockIndex > 0 {
		if hidBlockIndex-1 > len(heapOnNode.InputStream.Blocks) {
			return HeapOnNodeInputStream{}, errors.New("block doesn't exist")
		}

		for i := 0; i < len(heapOnNode.InputStream.Blocks); i++ {
			block := heapOnNode.InputStream.Blocks[i]

			blockSize, err := block.GetSize(formatType)

			if err != nil {
				return HeapOnNodeInputStream{}, err
			}

			blockOffset = blockOffset + blockSize

			if i == hidBlockIndex-1 {
				break
			}
		}
	}

	hidIndex := (hnid & 0xFFFF) >> 5

	heapOnNodePageMap, err := heapOnNode.GetPageMap(blockOffset)

	if err != nil {
		return HeapOnNodeInputStream{}, err
	}

	heapOnNodePageMap += blockOffset

	// The allocation table starts at byte offset 4 from the page map.
	// Every 2 bytes in the allocation table is the offset of the item.
	heapOnNodePageMap += (2 * hidIndex) + 2

	startOffset, err := heapOnNode.InputStream.SeekAndReadUint16(2, heapOnNodePageMap)

	if err != nil {
		return HeapOnNodeInputStream{}, err
	}

	startOffset += blockOffset

	endOffset, err := heapOnNode.InputStream.SeekAndReadUint16(2, heapOnNodePageMap+2)

	endOffset += blockOffset

	return HeapOnNodeInputStream{
		File:           pstFile,
		FormatType:     formatType,
		EncryptionType: encryptionType,
		FileOffset:     heapOnNode.InputStream.FileOffset,
		StartOffset:    startOffset,
		Size:           endOffset - startOffset,
		Blocks:         heapOnNode.InputStream.Blocks,
	}, nil
}

// Read reads from the node input stream.
func (heapOnNodeInputStream *HeapOnNodeInputStream) Read(outputBufferSize int, offset int) ([]byte, error) {
	if len(heapOnNodeInputStream.Blocks) == 0 {
		outputBuffer, err := heapOnNodeInputStream.File.Read(outputBufferSize, heapOnNodeInputStream.FileOffset+heapOnNodeInputStream.StartOffset+offset)

		if err != nil {
			return nil, err
		}

		switch heapOnNodeInputStream.EncryptionType {
		case EncryptionTypePermute:
			return DecodeCompressibleEncryption(outputBuffer), nil
		case EncryptionTypeNone:
			return outputBuffer, nil
		default:
			return nil, errors.New("unsupported encryption type")
		}
	} else {
		// The following code is harder to implement than it looks.
		// Before we used to read ALL blocks into memory and then get the bytes but this takes up way too much memory.
		// Now we only reads what's required.

		// Get the start offsets of each block.
		blockStartOffsets := make([]int, len(heapOnNodeInputStream.Blocks))
		currentBlockStartOffset := 0

		for i := 0; i < len(heapOnNodeInputStream.Blocks); i++ {
			block := heapOnNodeInputStream.Blocks[i]

			blockSize, err := block.GetSize(heapOnNodeInputStream.FormatType)

			if err != nil {
				return nil, err
			}

			blockStartOffsets[i] = currentBlockStartOffset
			currentBlockStartOffset += blockSize
		}

		// Get the current block based on the offset.
		currentBlock := 0
		currentBlockEndOffset := blockStartOffsets[currentBlock+1]

		for heapOnNodeInputStream.StartOffset+offset >= currentBlockEndOffset {
			currentBlock += 1

			if currentBlock == len(heapOnNodeInputStream.Blocks)-1 {
				break
			} else {
				currentBlockEndOffset = blockStartOffsets[currentBlock+1]
			}
		}

		block := heapOnNodeInputStream.Blocks[currentBlock]

		// This was the hardest part, figure out what the offset is into the current block.
		var currentOffsetInBlock int

		if currentBlock == 0 {
			currentOffsetInBlock = heapOnNodeInputStream.StartOffset + offset
		} else {
			if heapOnNodeInputStream.StartOffset >= blockStartOffsets[currentBlock] || offset >= blockStartOffsets[currentBlock] {
				currentOffsetInBlock = (heapOnNodeInputStream.StartOffset + offset) - blockStartOffsets[currentBlock]
			} else {
				currentOffsetInBlock = blockStartOffsets[currentBlock] - (heapOnNodeInputStream.StartOffset - offset)
			}
		}

		blockFileOffset, err := block.GetFileOffset(false, heapOnNodeInputStream.FormatType)

		if err != nil {
			return nil, err
		}

		blockSize, err := block.GetSize(heapOnNodeInputStream.FormatType)

		if err != nil {
			return nil, err
		}

		blockEndOffset := blockStartOffsets[currentBlock] + blockSize

		if outputBufferSize > blockSize {
			return nil, errors.New("output buffer size is larger than the block size")
		}

		// The end offset of this block should not be larger than the requested offset and output buffer size.
		if heapOnNodeInputStream.StartOffset+offset+outputBufferSize > blockEndOffset {
			return nil, errors.New("requested offset is larger than the maximum block size, please open an issue on GitHub")
		}

		blockData, err := heapOnNodeInputStream.File.Read(outputBufferSize, blockFileOffset+currentOffsetInBlock)

		return DecodeCompressibleEncryption(blockData), nil
	}
}

// ReadCompletely reads all the data.
func (heapOnNodeInputStream *HeapOnNodeInputStream) ReadCompletely(formatType string) ([]byte, error) {
	var outputBuffer []byte

	if len(heapOnNodeInputStream.Blocks) > 0 {
		currentOffset := 0

		for _, block := range heapOnNodeInputStream.Blocks {
			blockSize, err := block.GetSize(formatType)

			if err != nil {
				return nil, err
			}

			data, err := heapOnNodeInputStream.Read(blockSize, currentOffset)

			if err != nil {
				return nil, err
			}

			currentOffset += blockSize
			outputBuffer = append(outputBuffer, data...)
		}
	} else {
		data, err := heapOnNodeInputStream.Read(heapOnNodeInputStream.Size, 0)

		if err != nil {
			return nil, err
		}

		outputBuffer = data
	}

	return outputBuffer, nil
}

// SeekAndReadUint16 seeks and reads an uint16.
func (heapOnNodeInputStream *HeapOnNodeInputStream) SeekAndReadUint16(outputBufferSize int, offset int) (int, error) {
	if outputBufferSize > 2 || outputBufferSize < 1 {
		return -1, errors.New("invalid buffer size for uint16")
	}

	outputBuffer, err := heapOnNodeInputStream.Read(outputBufferSize, offset)

	if err != nil {
		return -1, err
	}

	switch outputBufferSize {
	case 1:
		return int(binary.LittleEndian.Uint16([]byte{outputBuffer[0], 0})), nil
	case 2:
		return int(binary.LittleEndian.Uint16(outputBuffer)), nil
	default:
		return -1, errors.New("invalid buffer size for uint16")
	}
}

// SeekAndReadUint32 seeks and reads an uint32.
func (heapOnNodeInputStream *HeapOnNodeInputStream) SeekAndReadUint32(outputBufferSize int, offset int) (int, error) {
	if outputBufferSize > 4 || outputBufferSize <= 1 {
		return -1, errors.New("invalid buffer size for uint32")
	}

	outputBuffer, err := heapOnNodeInputStream.Read(outputBufferSize, offset)

	if err != nil {
		return -1, err
	}

	return int(binary.LittleEndian.Uint32(outputBuffer)), nil
}
