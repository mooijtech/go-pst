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
	"encoding/binary"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/ianaindex"
)

// [MS-OXRTFCP]: Rich Text Format (RTF) Compression Algorithm:
// https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-oxrtfcp/65dfe2df-1b69-43fc-8ebd-21819a7463fb?redirectedfrom=MSDN

type RTFDecoder struct{}

func NewRTFDecoder() *RTFDecoder {
	return &RTFDecoder{}
}

const LZFUHeader = "{\\rtf1\\ansi\\mac\\deff0\\deftab720{\\fonttbl;}{\\f0\\fnil \\froman \\fswiss \\fmodern \\fscript \\fdecor MS Sans SerifSymbolArialTimes New RomanCourier{\\colortbl\\red0\\green0\\blue0\n\r\\par \\pard\\plain\\f0\\fs20\\b\\i\\u\\tab\\tx"

var (
	CompressionTypeCompressed   = []byte("LZFu")
	CompressionTypeUncompressed = []byte("MELA")
)

func (rtfDecoder *RTFDecoder) Decode(data []byte) (string, error) {
	_ = binary.LittleEndian.Uint32(data[:4]) // Compressed size
	uncompressedSize := int(binary.LittleEndian.Uint32(data[4:8]))
	compressionSignature := data[8:12]
	_ = binary.LittleEndian.Uint32(data[12:16]) // CRC

	if bytes.Equal(compressionSignature, CompressionTypeCompressed) {
		// Compressed
		outputBuffer := make([]byte, uncompressedSize)
		outputPosition := 0
		currentPosition := 16
		lzBuffer := make([]byte, 4096)

		asciiEncoding, err := ianaindex.MIME.Encoding("us-ascii")

		if err != nil {
			return "", errors.WithStack(err)
		}

		lzfuHeaderEncoded, err := asciiEncoding.NewEncoder().Bytes([]byte(LZFUHeader))

		if err != nil {
			return "", errors.WithStack(err)
		}

		copy(lzBuffer, lzfuHeaderEncoded)

		bufferPosition := len(lzfuHeaderEncoded)

		for currentPosition < len(data)-2 && outputPosition < len(outputBuffer) {
			flags := data[currentPosition] & 0xFF
			currentPosition++

			for i := 0; i < 8 && outputPosition < len(outputBuffer); i++ {
				isRef := (flags & 1) == 1
				flags >>= 1

				if isRef {
					refOffsetOrig := int(data[currentPosition] & 0xFF)
					currentPosition++
					refSizeOrig := int(data[currentPosition] & 0xFF)
					currentPosition++

					refOffset := (refOffsetOrig << 4) | (refSizeOrig >> 4)
					refSize := (refSizeOrig & 0xF) + 2

					for y := 0; y < refSize && outputPosition < len(outputBuffer); y++ {
						outputBuffer[outputPosition] = lzBuffer[refOffset]
						outputPosition++
						lzBuffer[bufferPosition] = lzBuffer[refOffset]

						bufferPosition++
						bufferPosition %= 4096
						refOffset++
						refOffset %= 4096
					}
				} else {
					lzBuffer[bufferPosition] = data[currentPosition]

					bufferPosition++
					bufferPosition %= 4096

					outputBuffer[outputPosition] = data[currentPosition]
					outputPosition++
					currentPosition++
				}
			}
		}

		if outputPosition != uncompressedSize {
			return "", errors.New("output position does not match uncompressed size")
		}

		return string(outputBuffer), nil
	} else if bytes.Equal(compressionSignature, CompressionTypeUncompressed) {
		// Uncompressed
		return string(data[16 : len(data)-16]), nil
	} else {
		return "", errors.New("unknown compression signature")
	}
}
