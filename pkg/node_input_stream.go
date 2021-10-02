// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/binary"
	"errors"
)

// NodeInputStream represents a node input stream for a Heap-on-Node.
type NodeInputStream struct {
	File *File
	EncryptionType string
	FileOffset int
	Size int
}

// Read reads from the node input stream.
func (nodeInputStream *NodeInputStream) Read(outputBufferSize int, offset int) ([]byte, error) {
	outputBuffer, err := nodeInputStream.File.Read(outputBufferSize, nodeInputStream.FileOffset + offset)

	if err != nil {
		return nil, err
	}

	switch nodeInputStream.EncryptionType {
	case EncryptionTypePermute:
		return DecodeCompressibleEncryption(outputBuffer), nil
	case EncryptionTypeNone:
		return outputBuffer, nil
	default:
		return nil, errors.New("unsupported encryption type")
	}
}

// SeekAndReadUint16 seeks and reads an uint16.
func (nodeInputStream *NodeInputStream) SeekAndReadUint16(outputBufferSize int, offset int) (int, error) {
	if outputBufferSize > 2 || outputBufferSize < 1 {
		return -1, errors.New("invalid buffer size for uint16")
	}

	outputBuffer, err := nodeInputStream.Read(outputBufferSize, offset)

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
func (nodeInputStream *NodeInputStream) SeekAndReadUint32(outputBufferSize int, offset int) (int, error) {
	if outputBufferSize > 4 || outputBufferSize <= 1 {
		return -1, errors.New("invalid buffer size for uint32")
	}

	outputBuffer, err := nodeInputStream.Read(outputBufferSize, offset)

	if err != nil {
		return -1, err
	}

	return int(binary.LittleEndian.Uint32(outputBuffer)), nil
}

// SeekAndReadUint64 seeks and reads an uint64.
func (nodeInputStream *NodeInputStream) SeekAndReadUint64(outputBufferSize int, offset int) (int, error) {
	if outputBufferSize > 8 || outputBufferSize <= 1 {
		return -1, errors.New("invalid buffer size for uint32")
	}

	outputBuffer, err := nodeInputStream.Read(outputBufferSize, offset)

	if err != nil {
		return -1, err
	}

	return int(binary.LittleEndian.Uint64(outputBuffer)), nil
}