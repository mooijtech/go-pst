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
	"github.com/tidwall/btree"
)

// BTreeStore is an abstraction used to store the node and block b-tree.
// This is useful if you want to persist the b-tree to disk for example.
// An initialized b-tree store can be passed to pst.NewFromReaderWithBTrees.
//
// This interface defines the functions we use from tidwall/btree.
type BTreeStore interface {
	// Load adds the b-tree nodes to the b-tree store.
	Load(node BTreeNode) (BTreeNode, bool)
	Get(key BTreeNode) (BTreeNode, bool)
	Len() int
	Clear()
}

// NewBTreeStoreInMemory creates a new b-tree store using google/btree.
func NewBTreeStoreInMemory() *btree.BTreeG[BTreeNode] {
	return btree.NewBTreeGOptions(BTreeNodeLessFunc, btree.Options{
		Degree: 255, // PST files use one byte to represent the node level which is uint8 (255).
	})
}
