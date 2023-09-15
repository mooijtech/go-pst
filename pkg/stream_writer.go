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
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
	"io"
)

// StreamWriter uses a write channel and callback channel (used by all writers).
type StreamWriter struct {
	// writer used by the WriteChannel.
	writer io.WriteSeeker
	// writeGroup is used to start the WriteChannel.
	writeGroup *errgroup.Group
	// writeChannel is the channel for writing.
	writeChannel chan io.WriterTo
	// callbackChannel is used to calculate the total written size.
	callbackChannel chan int64
}

// NewStreamWriter creates a new StreamWriter.
// The caller must start the WriteChannel of the returned StreamWriter with StartWriteChannel.
func NewStreamWriter(writer io.WriteSeeker, writeGroup *errgroup.Group) *StreamWriter {
	writeChannel := make(chan io.WriterTo)
	callbackChannel := make(chan int64)

	return &StreamWriter{
		writer:          writer,
		writeGroup:      writeGroup,
		writeChannel:    writeChannel,
		callbackChannel: callbackChannel,
	}
}

// Send the io.WriterTo to the WriteChannel.
func (streamWriter *StreamWriter) Send(writeRequests ...io.WriterTo) {
	for _, writeRequest := range writeRequests {
		streamWriter.writeChannel <- writeRequest
	}
}

// StartWriteChannel starts the WriteChannel.
func (streamWriter *StreamWriter) StartWriteChannel() {
	streamWriter.writeGroup.Go(func() error {
		for streamRequest := range streamWriter.writeChannel {
			streamWriter.writeGroup.Go(func() error {
				written, err := streamRequest.WriteTo(streamWriter.writer)

				if err != nil {
					return eris.Wrap(err, "failed to write to writer")
				}

				streamWriter.callbackChannel <- written

				return nil
			})
		}

		return nil
	})
}

// RegisterCallback registers a callback function.
func (streamWriter *StreamWriter) RegisterCallback(callbackChannel chan int64) {
	streamWriter.writeGroup.Go(func() error {
		for streamResponse := range streamWriter.callbackChannel {
			// Send to callback function.
			callbackChannel <- streamResponse
		}

		return nil
	})
}

// Close the stream writer.
func (streamWriter *StreamWriter) Close() {
	close(streamWriter.writeChannel)
	close(streamWriter.callbackChannel)
}
