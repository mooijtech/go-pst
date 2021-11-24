// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"golang.org/x/text/encoding/unicode"
)

// The libpff documentation states:
// "Unicode strings are stored in UTF-16 little-endian without the byte order mark (BOM)."
func BytesToString(input []byte) string {
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

	utf16String, err := decoder.String(string(input))

	if err != nil {
		return err.Error()
	}

	return utf16String
}