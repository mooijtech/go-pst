// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import "errors"

// GetBlockSize returns the size of a block.
// References "Blocks".
func (pstFile *File) GetBlockSize(formatType string) (int, error) {
	switch formatType {
	case FormatTypeUnicode:
		return 8192, nil
	case FormatTypeUnicode4k:
		return 65536, nil
	case FormatTypeANSI:
		return 8192, nil
	default:
		return -1, errors.New("unsupported format type")
	}
}

// GetBlockTrailerSize returns the size of a block trailer.
// References "Blocks".
func (pstFile *File) GetBlockTrailerSize(formatType string) (int, error) {
	switch formatType {
	case FormatTypeUnicode:
		return 16, nil
	case FormatTypeUnicode4k:
		return 16, nil
	case FormatTypeANSI:
		return 12, nil
	default:
		return -1, errors.New("unsupported format type")
	}
}
