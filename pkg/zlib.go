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
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/readahead"
	"github.com/rotisserie/eris"
	"io"
)

// ZLibDecompressor represents the ZLib decompressor used for Unicode 4k.
// Used by HeapOnNodeReader.
type ZLibDecompressor struct {
	// reader Read-Ahead which uses a Goroutine to get the result.
	// References:
	// - https://github.com/klauspost/readahead
	// - https://blog.klauspost.com/an-async-read-ahead-package-for-go/
	reader readahead.ReadSeekCloser
	// zlibReader inflates (decompresses compressed) bytes using ZLib
	// References https://github.com/klauspost/compress
	zlibReader io.ReadCloser
}

// NewZLibDecompressor creates a new ZLibDecompressor.
func NewZLibDecompressor(reader io.ReadSeeker) (*ZLibDecompressor, error) {
	readAheadReader := readahead.NewReadSeeker(reader)
	zlibReader, err := zlib.NewReader(readAheadReader)

	if err != nil {
		return nil, eris.Wrap(err, "failed to create Z-Lib reader")
	}

	return &ZLibDecompressor{
		reader:     readAheadReader,
		zlibReader: zlibReader,
	}, nil
}

// Decompress ZLib.
func (zlibDecompressor *ZLibDecompressor) Decompress(compressed []byte, outputBuffer io.Writer) (int64, error) {
	// Decompress using ZLib.
	if _, err := zlibDecompressor.zlibReader.Read(compressed); err != nil {
		return 0, eris.Wrap(err, "failed to read compressed bytes for ZLib")
	}

	// Copy the decompressed bytes to the output buffer.
	return io.Copy(outputBuffer, zlibDecompressor.reader)
}

// TODO - Compress.
