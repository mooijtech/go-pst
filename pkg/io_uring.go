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

//go:build linux

package pst

import (
	"context"
	"github.com/alitto/pond"
	"github.com/godzie44/go-uring/reactor"
	"github.com/godzie44/go-uring/uring"
	"os"
	"os/signal"
)

// NewAsync uses io_uring on Linux.
func NewAsync(name string) (*File, error) {
	asyncReader, err := NewReaderAsync(name)

	if err != nil {
		return nil, err
	}

	return New(asyncReader)
}

type AsyncReader struct {
	ring             *uring.Ring
	eventLoop        *reactor.Reactor
	eventLoopContext context.Context
	eventLoopCancel  context.CancelFunc
	reader           *os.File
	pool             *pond.WorkerPool
}

// NewReaderAsync uses io_uring on Linux for asynchronous I/O operations.
func NewReaderAsync(name string) (*AsyncReader, error) {
	reader, err := os.Open(name)

	if err != nil {
		return nil, err
	}

	ring, err := uring.New(8)

	if err != nil {
		return nil, err
	}

	eventLoop, err := reactor.New([]*uring.Ring{ring})

	if err != nil {
		return nil, err
	}

	eventLoopContext, eventLoopCancel := signal.NotifyContext(context.Background(), os.Interrupt)

	go func() {
		eventLoop.Run(eventLoopContext)
	}()

	return &AsyncReader{
		ring:             ring,
		eventLoop:        eventLoop,
		eventLoopContext: eventLoopContext,
		eventLoopCancel:  eventLoopCancel,
		reader:           reader,
		pool:             pond.New(8, 100),
	}, nil
}

// ReadAtAsync io_uring operation.
// The callback function is called when a CQE is received.
// Returns uint64 which can be used as the SQE identifier.
func (asyncReader *AsyncReader) ReadAtAsync(outputBuffer []byte, offset uint64, callback func(err error)) (uint64, error) {
	readOperation := uring.Read(asyncReader.reader.Fd(), outputBuffer, offset)

	return asyncReader.eventLoop.Queue(readOperation, func(event uring.CQEvent) {
		callback(event.Error())
	})
}

// ReadAt is a fall-back for where we haven't added async disk I/O support yet.
func (asyncReader *AsyncReader) ReadAt(outputBuffer []byte, offset int64) (int, error) {
	return asyncReader.reader.ReadAt(outputBuffer, offset)
}

// Close closes the event loop.
func (asyncReader *AsyncReader) Close() {
	<-asyncReader.eventLoopContext.Done()
	asyncReader.eventLoopCancel()
}
