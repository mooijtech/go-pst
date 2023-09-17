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
	"fmt"
	"github.com/rotisserie/eris"
	"io"
	"sort"
)

// HeapOnNodeReader implements io.SectionReader.
type HeapOnNodeReader struct {
	zLibDecompressor ZLibDecompressor
	blocks           []io.SectionReader
	blockOffsets     []int64
	totalBlockSize   int64
	options          Options
}

// NewHeapOnNodeReader creates a new Heap-on-Node reader.
func NewHeapOnNodeReader(options Options, blocks ...io.SectionReader) *HeapOnNodeReader {
	blockOffsets := make([]int64, len(blocks))
	blockOffset := int64(0)

	// Get the block offsets.
	for i, block := range blocks {
		blockOffsets[i] = blockOffset
		blockOffset += block.Size()
	}

	return &HeapOnNodeReader{
		blocks:         blocks,
		blockOffsets:   blockOffsets,
		totalBlockSize: blockOffset,
		options:        options,
	}
}

// Size is the total byte size.
func (heapOnNodeReader *HeapOnNodeReader) Size() int64 {
	return heapOnNodeReader.totalBlockSize
}

// ReadAt is adapted from Brad Fitz (http://talks.golang.org/2013/oscon-dl/sizereaderat.go).
func (heapOnNodeReader *HeapOnNodeReader) ReadAt(p []byte, requestedOffset int64) (n int, err error) {
	wantSize := len(p)

	// Skip past the requested offset.
	skipParts := sort.Search(len(heapOnNodeReader.blocks), func(i int) bool {
		// This function returns whether blocks[i] will
		// contribute any bytes to our output.
		part := heapOnNodeReader.blocks[i]
		return heapOnNodeReader.blockOffsets[i]+part.Size() > requestedOffset
	})
	blocks := heapOnNodeReader.blocks[skipParts:]

	// How far to skip in the first part.
	blockStartOffset := requestedOffset
	if len(blocks) > 0 {
		blockStartOffset -= heapOnNodeReader.blockOffsets[skipParts]
	}

	for len(blocks) > 0 && len(p) > 0 {
		readP := p
		partSize := blocks[0].Size()

		if int64(len(readP)) > partSize-blockStartOffset {
			readP = readP[:partSize-blockStartOffset]
		}

		pn, err := blocks[0].ReadAt(readP, blockStartOffset)

		if err != nil {
			return n, err
		}

		// Detect ZLib used by BTreeNode in the Unicode 4k format (OST).
		if heapOnNodeReader.options.formatType == FormatTypeUnicode4k {
			zlibDecompressor, err := NewZLibDecompressor(&blocks[0])

			if err != nil {
				return 0, eris.Wrap(err, "failed to ZLib decompress, possibly not compressed")
			}

			decompressedZLib := &bytes.Buffer{}

			// TODO - Which order in combination with Encryption?
			if _, err := zlibDecompressor.Decompress(readP, decompressedZLib); err != nil {
				return 0, eris.Wrap(err, "failed to decompress ZLib")
			}

			fmt.Printf("Got decompressed ZLib: %s\n", decompressedZLib)
		}

		// See DecodeCompressibleEncryption.
		switch heapOnNodeReader.options.encryptionType {
		case EncryptionTypeNone:
		case EncryptionTypePermute:
			copy(readP, heapOnNodeReader.DecodeCompressibleEncryption(readP))
		default:
			return n, ErrEncryptionTypeUnsupported
		}

		n += pn
		p = p[pn:]

		if int64(pn)+blockStartOffset == partSize {
			blocks = blocks[1:]
		}

		blockStartOffset = 0
	}

	if n != wantSize {
		return n, io.ErrUnexpectedEOF
	}
	return n, nil
}

// DecodeCompressibleEncryption decodes the Heap-on-Node using compressible encryption.
// References "Compressible encryption".
func (heapOnNodeReader *HeapOnNodeReader) DecodeCompressibleEncryption(data []byte) []byte {
	compressibleEncryption := []int{
		0x47, 0xf1, 0xb4, 0xe6, 0x0b, 0x6a, 0x72, 0x48, 0x85, 0x4e, 0x9e, 0xeb, 0xe2, 0xf8, 0x94, 0x53, 0xe0,
		0xbb, 0xa0, 0x02, 0xe8, 0x5a, 0x09, 0xab, 0xdb, 0xe3, 0xba, 0xc6, 0x7c, 0xc3, 0x10, 0xdd, 0x39, 0x05,
		0x96, 0x30, 0xf5, 0x37, 0x60, 0x82, 0x8c, 0xc9, 0x13, 0x4a, 0x6b, 0x1d, 0xf3, 0xfb, 0x8f, 0x26, 0x97,
		0xca, 0x91, 0x17, 0x01, 0xc4, 0x32, 0x2d, 0x6e, 0x31, 0x95, 0xff, 0xd9, 0x23, 0xd1, 0x00, 0x5e, 0x79,
		0xdc, 0x44, 0x3b, 0x1a, 0x28, 0xc5, 0x61, 0x57, 0x20, 0x90, 0x3d, 0x83, 0xb9, 0x43, 0xbe, 0x67, 0xd2,
		0x46, 0x42, 0x76, 0xc0, 0x6d, 0x5b, 0x7e, 0xb2, 0x0f, 0x16, 0x29, 0x3c, 0xa9, 0x03, 0x54, 0x0d, 0xda,
		0x5d, 0xdf, 0xf6, 0xb7, 0xc7, 0x62, 0xcd, 0x8d, 0x06, 0xd3, 0x69, 0x5c, 0x86, 0xd6, 0x14, 0xf7, 0xa5,
		0x66, 0x75, 0xac, 0xb1, 0xe9, 0x45, 0x21, 0x70, 0x0c, 0x87, 0x9f, 0x74, 0xa4, 0x22, 0x4c, 0x6f, 0xbf,
		0x1f, 0x56, 0xaa, 0x2e, 0xb3, 0x78, 0x33, 0x50, 0xb0, 0xa3, 0x92, 0xbc, 0xcf, 0x19, 0x1c, 0xa7, 0x63,
		0xcb, 0x1e, 0x4d, 0x3e, 0x4b, 0x1b, 0x9b, 0x4f, 0xe7, 0xf0, 0xee, 0xad, 0x3a, 0xb5, 0x59, 0x04, 0xea,
		0x40, 0x55, 0x25, 0x51, 0xe5, 0x7a, 0x89, 0x38, 0x68, 0x52, 0x7b, 0xfc, 0x27, 0xae, 0xd7, 0xbd, 0xfa,
		0x07, 0xf4, 0xcc, 0x8e, 0x5f, 0xef, 0x35, 0x9c, 0x84, 0x2b, 0x15, 0xd5, 0x77, 0x34, 0x49, 0xb6, 0x12,
		0x0a, 0x7f, 0x71, 0x88, 0xfd, 0x9d, 0x18, 0x41, 0x7d, 0x93, 0xd8, 0x58, 0x2c, 0xce, 0xfe, 0x24, 0xaf,
		0xde, 0xb8, 0x36, 0xc8, 0xa1, 0x80, 0xa6, 0x99, 0x98, 0xa8, 0x2f, 0x0e, 0x81, 0x65, 0x73, 0xe4, 0xc2,
		0xa2, 0x8a, 0xd4, 0xe1, 0x11, 0xd0, 0x08, 0x8b, 0x2a, 0xf2, 0xed, 0x9a, 0x64, 0x3f, 0xc1, 0x6c, 0xf9, 0xec,
	}

	for i := 0; i < len(data); i++ {
		temp := data[i] & 0xff
		data[i] = byte(compressibleEncryption[temp])
	}

	return data
}

// TODO - EncodeCompressibleEncryption
