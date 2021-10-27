// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"bufio"
	"bytes"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"strings"
)

// BytesToString converts bytes to string and deals with encoding.
// References https://stackoverflow.com/a/55632545
func BytesToString(input []byte) string {
	inputScanner := bufio.NewScanner(transform.NewReader(bytes.NewReader(input), unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))

	outputStringBuilder := strings.Builder{}

	for inputScanner.Scan() {
		outputStringBuilder.WriteString(inputScanner.Text())
	}

	return outputStringBuilder.String()
}