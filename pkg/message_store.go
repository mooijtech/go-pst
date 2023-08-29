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
	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
)

// MessageStore represents the MessageStore of a PST file.
// Each PST file has at most one MessageStore.
type MessageStore struct {
	// PropertyContext represents the PropertyContext of the MessageStore.
	PropertyContext *PropertyContext
}

// NewMessageStore creates a new MessageStore.
func NewMessageStore(propertyContext *PropertyContext) *MessageStore {
	return &MessageStore{
		PropertyContext: propertyContext,
	}
}

// GetMessageStore returns the MessageStore of the PST file.
func (file *File) GetMessageStore() (*MessageStore, error) {
	dataBTreeNode, err := file.GetDataBTreeNode(IdentifierMessageStore)

	if err != nil {
		return nil, eris.Wrapf(err, "failed to find data b-tree node")
	}

	heapOnNode, err := file.GetHeapOnNode(dataBTreeNode)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get Heap-on-Node")
	}

	propertyContext, err := file.GetPropertyContext(heapOnNode)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return NewMessageStore(propertyContext), nil
}
