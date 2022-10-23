// go-pst is a library for reading Personal Storage Table (.pst) files (written in Go/Golang).
//
// Copyright (C) 2022  Marten Mooij
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
	return btree.NewBTreeG(BTreeNodeLessFunc)
}
