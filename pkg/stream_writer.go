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
	"sync"
)

// StreamWriter uses a write channel and callback channel.
// StreamWriter is used by all writers and is generic.
type StreamWriter struct {
	// Writer used by the WriteChannel.
	Writer io.WriteSeeker
	// WriteGroup is used to start the WriteChannel.
	WriteGroup *errgroup.Group
	// WriteChannel is the channel for writing.
	WriteChannel chan StreamRequest
	// CallbackChannel is where a StreamResponse is sent.
	// Transform may be used to return a value other than the default, which is bytes written (int64).
	CallbackChannel chan StreamResponse
	// TransformFunc transforms the request to a response.
	// For example, I can transform a byte representation (request) to a struct (response).
	TransformFunc TransformFunc
}

// TransformFunc is used by Transform to convert a StreamRequest to a StreamResponse.
// See WithTransform of a StreamWriter.
type TransformFunc func(streamRequest StreamRequest) (StreamResponse, error)

// NewStreamWriter creates a new StreamWriter.
// The caller must start the WriteChannel of the returned StreamWriter with StartWriteChannel.
// See WithTransform.
func NewStreamWriter(writer io.WriteSeeker, writeGroup *errgroup.Group) *StreamWriter {
	writeChannel := make(chan StreamRequest)
	callbackChannel := make(chan StreamResponse)

	return &StreamWriter{
		Writer:          writer,
		WriteGroup:      writeGroup,
		WriteChannel:    writeChannel,
		CallbackChannel: callbackChannel,
	}
}

// WithTransform allows transforming the StreamRequest to a StreamResponse for the callback.
func (streamWriter *StreamWriter) WithTransform(transformFunc TransformFunc) *StreamWriter {
	streamWriter.TransformFunc = transformFunc

	return streamWriter
}

// StreamRequest represents something to write (see StreamResponse).
type StreamRequest interface {
	io.WriterTo
}

// StreamResponse represents the write response which is sent to a callback.
type StreamResponse struct {
	// Value represents the result of the write operation.
	// It can be transformed with Transform (you will never be a real woman).
	// By default, this is int64.
	Value any
}

// NewStreamResponse creates a new StreamResponse.
func NewStreamResponse(value any) StreamResponse {
	return StreamResponse{Value: value}
}

// Send the StreamRequest to the WriteChannel.
func (streamWriter *StreamWriter) Send(streamRequests ...StreamRequest) {
	for _, streamRequest := range streamRequests {
		streamWriter.WriteChannel <- streamRequest
	}
}

// StartWriteChannel starts the WriteChannel.
func (streamWriter *StreamWriter) StartWriteChannel() {
	streamWriter.WriteGroup.Go(func() error {
		// Wait for writes to complete.
		waitGroup := sync.WaitGroup{}

		// Launch writes all at once.
		for streamRequest := range streamWriter.WriteChannel {
			waitGroup.Add(1)

			streamWriter.WriteGroup.Go(func() error {
				written, err := streamRequest.WriteTo(streamWriter.Writer)

				if err != nil {
					return eris.Wrap(err, "failed to write to writer")
				}

				// Callback, by default sends the amount of bytes written as int64.
				// Optionally transforms (via TransformFunc) to a different callback value.
				if streamWriter.TransformFunc != nil {
					transformedRequest, err := streamWriter.TransformFunc(streamRequest)

					if err != nil {
						return eris.Wrap(err, "failed during transform of request for callback")
					}

					streamWriter.CallbackChannel <- transformedRequest
				} else {
					streamWriter.CallbackChannel <- NewStreamResponse(written)
				}

				return nil
			})
		}

		// Collect results.
		waitGroup.Wait()

		return nil
	})
}
