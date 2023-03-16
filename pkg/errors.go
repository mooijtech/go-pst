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
	"errors"
)

// Constants defining go-pst errors.
// We use stack-traces (Wrap) via https://github.com/rotisserie/eris
var (
	ErrFileSignatureInvalid             = errors.New("go-pst: invalid file signature")
	ErrFormatTypeUnsupported            = errors.New("go-pst: unsupported format type")
	ErrEncryptionTypeUnsupported        = errors.New("go-pst: unsupported encryption type")
	ErrContentTypeUnsupported           = errors.New("go-pst: unsupported content type")
	ErrMessageIdentifierTypeInvalid     = errors.New("go-pst: invalid message identifier type")
	ErrPropertyNotFound                 = errors.New("go-pst: failed to find property by ID")
	ErrTableTypeInvalid                 = errors.New("go-pst: invalid table type")
	ErrBTreeNodeConflict                = errors.New("go-pst: conflicting b-tree node entry")
	ErrBTreeNodeNotFound                = errors.New("go-pst: failed to find b-tree node")
	ErrHeapOnNodeExternalNode           = errors.New("go-pst: external node, no local descriptors provided")
	ErrTableSignatureInvalid            = errors.New("go-pst: invalid table signature")
	ErrTableContextNoColumns            = errors.New("go-pst: there are no columns in this table context")
	ErrTableContextNoRows               = errors.New("go-pst: there are no rows in this table context")
	ErrBlockTypeInvalid                 = errors.New("go-pst: unsupported block type")
	ErrBlockSignatureInvalid            = errors.New("go-pst: invalid block signature")
	ErrAttachmentIndexInvalid           = errors.New("go-pst: invalid attachment index, there are no more attachments")
	ErrLocalDescriptorsSignatureInvalid = errors.New("go-pst: invalid local descriptors signature")
	ErrLocalDescriptorNotFound          = errors.New("go-pst: failed to find local descriptor")
	ErrLocalDescriptorBranchNode        = errors.New("go-pst: local descriptors level is not 0, please open an issue on GitHub for this to be implemented")
	ErrPropertyTypeMismatch             = errors.New("go-pst: property type is not the same as the value expected from the caller")
	ErrPropertyNoData                   = errors.New("go-pst: property has no data")
	ErrNameToIDMapKeyNotFound           = errors.New("go-pst: failed to find key in Name-To-ID Map")
	ErrMessagesNotFound                 = errors.New("go-pst: folder has no messages")
	ErrAttachmentsNotFound              = errors.New("go-pst: message has no attachments")
	ErrBlockIndexNotFound               = errors.New("go-pst: block index not found")
	ErrTotalBlocksSizeMismatch          = errors.New("go-pst: block total size mismatch")
)
