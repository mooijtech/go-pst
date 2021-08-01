// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"bufio"
	"bytes"
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