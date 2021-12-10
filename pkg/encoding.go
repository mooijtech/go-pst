// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"log"
	"strconv"
	"strings"
)

// DecodeBytesToUTF16String decodes the bytes to UTF-16.
// The libpff documentation states:
// "Unicode strings are stored in UTF-16 little-endian without the byte order mark (BOM)."
func DecodeBytesToUTF16String(input []byte) (string, error) {
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()

	utf16String, err := decoder.String(string(input))

	if err != nil {
		return "", err
	}

	return utf16String, nil
}

// DecodeBytesToString decodes the message property context item to string using the message encoding.
func DecodeBytesToString(encoding Encoding, data []byte) (string, error) {
	if encoding.Name == "us-ascii" {
		// enron.pst shows replacement characters after each character with us-ascii but works with UTF-16.
		return DecodeBytesToUTF16String(data)
	}

	mimeEncoding, err := ianaindex.MIME.Encoding(encoding.Name)

	if err != nil {
		return "", err
	}

	if mimeEncoding == nil {
		if encoding.Name == "gb2312" {
			// Encoding gb2312 gives "invalid memory address or nil pointer dereference".
			// References https://github.com/golang/go/issues/24636
			decodedBytes, err := simplifiedchinese.HZGB2312.NewDecoder().Bytes(data)

			if err != nil {
				return "", err
			}

			return string(decodedBytes), nil
		} else {
			fmt.Printf("Failed to get encoding \"%s\", defaulting to UTF-16. Please open an issue on GitHub to add this encoding.\n", encoding.Name)

			utf16String, err := DecodeBytesToUTF16String(data)

			if err != nil {
				return "", err
			}

			log.Fatalf("Got it: %s", utf16String)

			return utf16String, nil
		}
	}

	decodedBytes, err := mimeEncoding.NewDecoder().Bytes(data)

	if err != nil {
		return "", err
	}

	return string(decodedBytes), nil
}

// Encoding represents an IANA index encoding.
type Encoding struct {
	// References https://docs.microsoft.com/en-us/windows/win32/intl/code-page-identifiers
	Identifier int
	Name string
}

//go:embed encodings.csv
var encodings string

// GetEncodings returns all available encodings.
func GetEncodings() ([]Encoding, error) {
	csvReader := csv.NewReader(strings.NewReader(encodings))

	csvEncodings, err := csvReader.ReadAll()

	if err != nil {
		return nil, err
	}

	var encodings []Encoding

	for _, encodingRow := range csvEncodings {
		identifier, err := strconv.ParseInt(encodingRow[0], 10, 64)

		if err != nil {
			continue
		}

		encodings = append(encodings, Encoding {
			Identifier: int(identifier),
			Name: encodingRow[1],
		})
	}

	return encodings, nil
}

// FindEncoding returns the encoding by the specified identifier.
func FindEncoding(identifier int) (Encoding, error) {
	encodings, err := GetEncodings()

	if err != nil {
		return Encoding{}, err
	}

	for _, encoding := range encodings {
		if encoding.Identifier == identifier {
			return encoding, nil
		}
	}

	return Encoding{}, errors.New(fmt.Sprintf("failed to find encoding: %d", identifier))
}

// String returns the encoding name.
func (encoding *Encoding) String() string {
	return encoding.Name
}

// GetEncoding returns the encoding of the message.
func (message *Message) GetEncoding() (Encoding, error) {
	encoding, err := message.GetInteger(16381) // PidTagMessageCodepage

	if err != nil {
		encoding, err = message.GetInteger(26307) // PidTagCodepage

		if err != nil {
			encoding, err = message.GetInteger(16350) // PidTagInternetCodepage

			if err != nil {
				// Encoding is set to -1
			}
		}
	}

	if encoding != -1 {
		// Found the encoding identifier.
		foundEncoding, err := FindEncoding(encoding)

		if err != nil {
			fmt.Printf("Unsupported encoding (%d), please open an issue on GitHub to support this encoding. Defaulting to UTF-8.\n", encoding)

			return Encoding {
				Identifier: 65001,
				Name: "utf-8",
			}, nil
		}

		return foundEncoding, nil
	} else {
		// TODO - Lookup the global encoding in the Message Store.
		fmt.Printf("Failed to find message encoding, defaulting to UTF-8.\n")

		return Encoding {
			Identifier: 65001,
			Name: "utf-8",
		}, nil
	}
}