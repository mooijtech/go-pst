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

import "github.com/rotisserie/eris"

// GetMessageStore returns the message store of the PST file.
func (file *File) GetMessageStore() (*PropertyContext, error) {
	dataBTreeNode, err := file.GetDataBTreeNode(IdentifierMessageStore)

	if err != nil {
		return nil, eris.Wrapf(err, "failed to find data b-tree node")
	}

	heapOnNode, err := file.GetHeapOnNode(dataBTreeNode)

	if err != nil {
		return nil, eris.Wrap(err, "failed to get Heap-on-Node")
	}

	return file.GetPropertyContext(heapOnNode)
}
