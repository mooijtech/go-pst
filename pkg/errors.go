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
	"github.com/rotisserie/eris"
)

// Constants defining go-pst eris.
// We use stack-traces via https://github.com/rotisserie/eris
var (
	ErrFileSignatureInvalid             = eris.New("go-pst: invalid file signature")
	ErrFormatTypeUnsupported            = eris.New("go-pst: unsupported format type")
	ErrEncryptionTypeUnsupported        = eris.New("go-pst: unsupported encryption type")
	ErrContentTypeUnsupported           = eris.New("go-pst: unsupported content type")
	ErrMessageIdentifierTypeInvalid     = eris.New("go-pst: invalid message identifier type")
	ErrPropertyNotFound                 = eris.New("go-pst: failed to find property by ID")
	ErrTableTypeInvalid                 = eris.New("go-pst: invalid table type")
	ErrBTreeNodeConflict                = eris.New("go-pst: conflicting b-tree node entry")
	ErrBTreeNodeNotFound                = eris.New("go-pst: failed to find b-tree node")
	ErrHeapOnNodeExternalNode           = eris.New("go-pst: external node, no local descriptors provided")
	ErrTableSignatureInvalid            = eris.New("go-pst: invalid table signature")
	ErrTableContextNoColumns            = eris.New("go-pst: there are no columns in this table context")
	ErrTableContextNoRows               = eris.New("go-pst: there are no rows in this table context")
	ErrBlockTypeInvalid                 = eris.New("go-pst: unsupported block type")
	ErrBlockSignatureInvalid            = eris.New("go-pst: invalid block signature")
	ErrAttachmentIndexInvalid           = eris.New("go-pst: invalid attachment index, there are no more attachments")
	ErrLocalDescriptorsSignatureInvalid = eris.New("go-pst: invalid local descriptors signature")
	ErrLocalDescriptorNotFound          = eris.New("go-pst: failed to find local descriptor")
	ErrLocalDescriptorBranchNode        = eris.New("go-pst: local descriptors level is not 0, please open an issue on GitHub for this to be implemented")
	ErrPropertyTypeMismatch             = eris.New("go-pst: property type is not the same as the value expected from the caller")
	ErrPropertyNoData                   = eris.New("go-pst: property has no data")
	ErrNameToIDMapKeyNotFound           = eris.New("go-pst: failed to find key in Name-To-ID Map")
	ErrMessagesNotFound                 = eris.New("go-pst: folder has no messages")
	ErrAttachmentsNotFound              = eris.New("go-pst: message has no attachments")
	ErrBlockIndexNotFound               = eris.New("go-pst: block index not found")
	ErrTotalBlocksSizeMismatch          = eris.New("go-pst: block total size mismatch")
)
