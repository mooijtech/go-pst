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

package writer

import (
	"io"
)

// MessageStoreWriter represents a writer for Message Stores.
type MessageStoreWriter struct {
	// PropertyContextWriter represents the pst.PropertyContext writer.
	PropertyContextWriter *PropertyContextWriter
}

// TODO - properties.MessageStore

// NewMessageStoreWriter creates a new MessageStoreWriter.
func NewMessageStoreWriter(propertyContextWriter *PropertyContextWriter) *MessageStoreWriter {
	return &MessageStoreWriter{
		PropertyContextWriter: propertyContextWriter,
	}
}

// WriteTo writes the Message Store.
func (messageStoreWriter *MessageStoreWriter) WriteTo(writer io.Writer) (int64, error) {
	return messageStoreWriter.PropertyContextWriter.WriteTo(writer)
}
