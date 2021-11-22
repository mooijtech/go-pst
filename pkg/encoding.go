// Package pst
// This file is part of go-pst (https://github.com/mooijtech/go-pst)
// Copyright (C) 2021 Marten Mooij (https://www.mooijtech.com/)
package pst

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Encoding represents an IANA index encoding.
type Encoding struct {
	// References https://docs.microsoft.com/en-us/windows/win32/intl/code-page-identifiers
	Identifier int
	Name string
}

// GetEncodings returns all available encodings.
func GetEncodings() ([]Encoding, error) {
	csvFile, err := os.Open("data/encoding.csv")

	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(csvFile)

	csvEncodings, err := csvReader.ReadAll()

	if err != nil {
		return nil, err
	}

	err = csvFile.Close()

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
	encoding := message.GetInteger(16381) // PidTagMessageCodepage

	if encoding == -1 {
		encoding = message.GetInteger(26307) // PidTagCodepage

		if encoding == -1 {
			encoding = message.GetInteger(16350) // PidTagInternetCodepage
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