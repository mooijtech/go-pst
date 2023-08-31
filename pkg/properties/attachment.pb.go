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

//go:generate msgp -tests=false

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.2
// source: attachment.proto

package properties

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Attachment struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Contains the Content-Type of the Mac attachment.
	AttachmentMacContentType *string `protobuf:"bytes,1,opt,name=attachment_mac_content_type,json=attachmentMacContentType,proto3,oneof" json:"attachment_mac_content_type,omitempty"`
	// Contains the original permission type data associated with a web reference attachment.
	AttachmentOriginalPermissionType *int32 `protobuf:"varint,3,opt,name=attachment_original_permission_type,json=attachmentOriginalPermissionType,proto3,oneof" json:"attachment_original_permission_type,omitempty"`
	// Contains the permission type data associated with a web reference attachment.
	AttachmentPermissionType *int32 `protobuf:"varint,4,opt,name=attachment_permission_type,json=attachmentPermissionType,proto3,oneof" json:"attachment_permission_type,omitempty"`
	// Contains the provider type data associated with a web reference attachment.
	AttachmentProviderType *string `protobuf:"bytes,5,opt,name=attachment_provider_type,json=attachmentProviderType,proto3,oneof" json:"attachment_provider_type,omitempty"`
	// Contains the base of a relative URI.
	AttachContentBase *string `protobuf:"bytes,7,opt,name=attach_content_base,json=attachContentBase,proto3,oneof" json:"attach_content_base,omitempty" msg:"14097,omitempty" type:"31,omitempty"`  
	// Contains a content identifier unique to the Message object that matches a corresponding "cid:" URI schema reference in the HTML body of the Message object.
	AttachContentId *string `protobuf:"bytes,8,opt,name=attach_content_id,json=attachContentId,proto3,oneof" json:"attach_content_id,omitempty" msg:"14098,omitempty" type:"31,omitempty"`  
	// Contains a relative or full URI that matches a corresponding reference in the HTML body of a Message object.
	AttachContentLocation *string `protobuf:"bytes,9,opt,name=attach_content_location,json=attachContentLocation,proto3,oneof" json:"attach_content_location,omitempty" msg:"14099,omitempty" type:"31,omitempty"`  
	// Contains a file name extension that indicates the document type of an attachment.
	AttachExtension *string `protobuf:"bytes,13,opt,name=attach_extension,json=attachExtension,proto3,oneof" json:"attach_extension,omitempty" msg:"14083,omitempty" type:"31,omitempty"`  
	// Contains the 8.3 name of the PidTagAttachLongFilename property (section 2.595).
	AttachFilename *string `protobuf:"bytes,14,opt,name=attach_filename,json=attachFilename,proto3,oneof" json:"attach_filename,omitempty" msg:"14084,omitempty" type:"31,omitempty"`  
	// Indicates which body formats might reference this attachment when rendering data.
	AttachFlags *int32 `protobuf:"varint,15,opt,name=attach_flags,json=attachFlags,proto3,oneof" json:"attach_flags,omitempty" msg:"14100,omitempty" type:"3,omitempty"`  
	// Contains the full filename and extension of the Attachment object.
	AttachLongFilename *string `protobuf:"bytes,16,opt,name=attach_long_filename,json=attachLongFilename,proto3,oneof" json:"attach_long_filename,omitempty" msg:"14087,omitempty" type:"31,omitempty"`  
	// Contains the fully-qualified path and file name with extension.
	AttachLongPathname *string `protobuf:"bytes,17,opt,name=attach_long_pathname,json=attachLongPathname,proto3,oneof" json:"attach_long_pathname,omitempty" msg:"14093,omitempty" type:"31,omitempty"`  
	// Indicates that a contact photo attachment is attached to a Contact object.
	AttachmentContactPhoto *bool `protobuf:"varint,18,opt,name=attachment_contact_photo,json=attachmentContactPhoto,proto3,oneof" json:"attachment_contact_photo,omitempty" msg:"32767,omitempty" type:"11,omitempty"`  
	// Indicates special handling for an Attachment object.
	AttachmentFlags *int32 `protobuf:"varint,19,opt,name=attachment_flags,json=attachmentFlags,proto3,oneof" json:"attachment_flags,omitempty" msg:"32765,omitempty" type:"3,omitempty"`  
	// Indicates whether an Attachment object is hidden from the end user.
	AttachmentHidden *bool `protobuf:"varint,20,opt,name=attachment_hidden,json=attachmentHidden,proto3,oneof" json:"attachment_hidden,omitempty" msg:"32766,omitempty" type:"11,omitempty"`  
	// Contains the type of Message object to which an attachment is linked.
	AttachmentLinkId *int32 `protobuf:"varint,21,opt,name=attachment_link_id,json=attachmentLinkId,proto3,oneof" json:"attachment_link_id,omitempty" msg:"32762,omitempty" type:"3,omitempty"`  
	// Represents the way the contents of an attachment are accessed.
	AttachMethod *int32 `protobuf:"varint,22,opt,name=attach_method,json=attachMethod,proto3,oneof" json:"attach_method,omitempty" msg:"14085,omitempty" type:"3,omitempty"`  
	// Contains a content-type MIME header.
	AttachMimeTag *string `protobuf:"bytes,23,opt,name=attach_mime_tag,json=attachMimeTag,proto3,oneof" json:"attach_mime_tag,omitempty" msg:"14094,omitempty" type:"31,omitempty"`  
	// Identifies the Attachment object within its Message object.
	AttachNumber *int32 `protobuf:"varint,24,opt,name=attach_number,json=attachNumber,proto3,oneof" json:"attach_number,omitempty" msg:"3617,omitempty" type:"3,omitempty"`  
	// Contains the 8.3 name of the PidTagAttachLongPathname property (section 2.596).
	AttachPathname *string `protobuf:"bytes,25,opt,name=attach_pathname,json=attachPathname,proto3,oneof" json:"attach_pathname,omitempty" msg:"14088,omitempty" type:"31,omitempty"`  
	// Contains the size, in bytes, consumed by the Attachment object on the server.
	AttachSize *int32 `protobuf:"varint,27,opt,name=attach_size,json=attachSize,proto3,oneof" json:"attach_size,omitempty" msg:"3616,omitempty" type:"3,omitempty"`  
	// Contains the name of an attachment file, modified so that it can be correlated with TNEF messages.
	AttachTransportName *string `protobuf:"bytes,29,opt,name=attach_transport_name,json=attachTransportName,proto3,oneof" json:"attach_transport_name,omitempty" msg:"14092,omitempty" type:"31,omitempty"`  
	// Specifies the character set of an attachment received via MIME with the content-type of text.
	TextAttachmentCharset *string `protobuf:"bytes,31,opt,name=text_attachment_charset,json=textAttachmentCharset,proto3,oneof" json:"text_attachment_charset,omitempty" msg:"14107,omitempty" type:"31,omitempty"`  
}

func (x *Attachment) Reset() {
	*x = Attachment{}
	if protoimpl.UnsafeEnabled {
		mi := &file_attachment_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Attachment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Attachment) ProtoMessage() {}

func (x *Attachment) ProtoReflect() protoreflect.Message {
	mi := &file_attachment_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Attachment.ProtoReflect.Descriptor instead.
func (*Attachment) Descriptor() ([]byte, []int) {
	return file_attachment_proto_rawDescGZIP(), []int{0}
}

func (x *Attachment) GetAttachmentMacContentType() string {
	if x != nil && x.AttachmentMacContentType != nil {
		return *x.AttachmentMacContentType
	}
	return ""
}

func (x *Attachment) GetAttachmentOriginalPermissionType() int32 {
	if x != nil && x.AttachmentOriginalPermissionType != nil {
		return *x.AttachmentOriginalPermissionType
	}
	return 0
}

func (x *Attachment) GetAttachmentPermissionType() int32 {
	if x != nil && x.AttachmentPermissionType != nil {
		return *x.AttachmentPermissionType
	}
	return 0
}

func (x *Attachment) GetAttachmentProviderType() string {
	if x != nil && x.AttachmentProviderType != nil {
		return *x.AttachmentProviderType
	}
	return ""
}

func (x *Attachment) GetAttachContentBase() string {
	if x != nil && x.AttachContentBase != nil {
		return *x.AttachContentBase
	}
	return ""
}

func (x *Attachment) GetAttachContentId() string {
	if x != nil && x.AttachContentId != nil {
		return *x.AttachContentId
	}
	return ""
}

func (x *Attachment) GetAttachContentLocation() string {
	if x != nil && x.AttachContentLocation != nil {
		return *x.AttachContentLocation
	}
	return ""
}

func (x *Attachment) GetAttachExtension() string {
	if x != nil && x.AttachExtension != nil {
		return *x.AttachExtension
	}
	return ""
}

func (x *Attachment) GetAttachFilename() string {
	if x != nil && x.AttachFilename != nil {
		return *x.AttachFilename
	}
	return ""
}

func (x *Attachment) GetAttachFlags() int32 {
	if x != nil && x.AttachFlags != nil {
		return *x.AttachFlags
	}
	return 0
}

func (x *Attachment) GetAttachLongFilename() string {
	if x != nil && x.AttachLongFilename != nil {
		return *x.AttachLongFilename
	}
	return ""
}

func (x *Attachment) GetAttachLongPathname() string {
	if x != nil && x.AttachLongPathname != nil {
		return *x.AttachLongPathname
	}
	return ""
}

func (x *Attachment) GetAttachmentContactPhoto() bool {
	if x != nil && x.AttachmentContactPhoto != nil {
		return *x.AttachmentContactPhoto
	}
	return false
}

func (x *Attachment) GetAttachmentFlags() int32 {
	if x != nil && x.AttachmentFlags != nil {
		return *x.AttachmentFlags
	}
	return 0
}

func (x *Attachment) GetAttachmentHidden() bool {
	if x != nil && x.AttachmentHidden != nil {
		return *x.AttachmentHidden
	}
	return false
}

func (x *Attachment) GetAttachmentLinkId() int32 {
	if x != nil && x.AttachmentLinkId != nil {
		return *x.AttachmentLinkId
	}
	return 0
}

func (x *Attachment) GetAttachMethod() int32 {
	if x != nil && x.AttachMethod != nil {
		return *x.AttachMethod
	}
	return 0
}

func (x *Attachment) GetAttachMimeTag() string {
	if x != nil && x.AttachMimeTag != nil {
		return *x.AttachMimeTag
	}
	return ""
}

func (x *Attachment) GetAttachNumber() int32 {
	if x != nil && x.AttachNumber != nil {
		return *x.AttachNumber
	}
	return 0
}

func (x *Attachment) GetAttachPathname() string {
	if x != nil && x.AttachPathname != nil {
		return *x.AttachPathname
	}
	return ""
}

func (x *Attachment) GetAttachSize() int32 {
	if x != nil && x.AttachSize != nil {
		return *x.AttachSize
	}
	return 0
}

func (x *Attachment) GetAttachTransportName() string {
	if x != nil && x.AttachTransportName != nil {
		return *x.AttachTransportName
	}
	return ""
}

func (x *Attachment) GetTextAttachmentCharset() string {
	if x != nil && x.TextAttachmentCharset != nil {
		return *x.TextAttachmentCharset
	}
	return ""
}

var File_attachment_proto protoreflect.FileDescriptor

var file_attachment_proto_rawDesc = []byte{
	0x0a, 0x10, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x87, 0x0e, 0x0a, 0x0a, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e,
	0x74, 0x12, 0x42, 0x0a, 0x1b, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f,
	0x6d, 0x61, 0x63, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x18, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68,
	0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x61, 0x63, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79,
	0x70, 0x65, 0x88, 0x01, 0x01, 0x12, 0x52, 0x0a, 0x23, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d,
	0x65, 0x6e, 0x74, 0x5f, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x5f, 0x70, 0x65, 0x72,
	0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x48, 0x01, 0x52, 0x20, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74,
	0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x88, 0x01, 0x01, 0x12, 0x41, 0x0a, 0x1a, 0x61, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x48, 0x02, 0x52,
	0x18, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x50, 0x65, 0x72, 0x6d, 0x69,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x88, 0x01, 0x01, 0x12, 0x3d, 0x0a, 0x18,
	0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x64, 0x65, 0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x48, 0x03,
	0x52, 0x16, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x76,
	0x69, 0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x88, 0x01, 0x01, 0x12, 0x33, 0x0a, 0x13, 0x61,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x62, 0x61,
	0x73, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x48, 0x04, 0x52, 0x11, 0x61, 0x74, 0x74, 0x61,
	0x63, 0x68, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x42, 0x61, 0x73, 0x65, 0x88, 0x01, 0x01,
	0x12, 0x2f, 0x0a, 0x11, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x48, 0x05, 0x52, 0x0f, 0x61,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x88, 0x01,
	0x01, 0x12, 0x3b, 0x0a, 0x17, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x5f, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x09, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x06, 0x52, 0x15, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x12, 0x2e,
	0x0a, 0x10, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x48, 0x07, 0x52, 0x0f, 0x61, 0x74, 0x74, 0x61,
	0x63, 0x68, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x12, 0x2c,
	0x0a, 0x0f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x48, 0x08, 0x52, 0x0e, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x46, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x12, 0x26, 0x0a, 0x0c,
	0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x18, 0x0f, 0x20, 0x01,
	0x28, 0x05, 0x48, 0x09, 0x52, 0x0b, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x46, 0x6c, 0x61, 0x67,
	0x73, 0x88, 0x01, 0x01, 0x12, 0x35, 0x0a, 0x14, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x6c,
	0x6f, 0x6e, 0x67, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x10, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x0a, 0x52, 0x12, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x4c, 0x6f, 0x6e, 0x67,
	0x46, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x12, 0x35, 0x0a, 0x14, 0x61,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x09, 0x48, 0x0b, 0x52, 0x12, 0x61, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x4c, 0x6f, 0x6e, 0x67, 0x50, 0x61, 0x74, 0x68, 0x6e, 0x61, 0x6d, 0x65, 0x88,
	0x01, 0x01, 0x12, 0x3d, 0x0a, 0x18, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74,
	0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x5f, 0x70, 0x68, 0x6f, 0x74, 0x6f, 0x18, 0x12,
	0x20, 0x01, 0x28, 0x08, 0x48, 0x0c, 0x52, 0x16, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65,
	0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x50, 0x68, 0x6f, 0x74, 0x6f, 0x88, 0x01,
	0x01, 0x12, 0x2e, 0x0a, 0x10, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f,
	0x66, 0x6c, 0x61, 0x67, 0x73, 0x18, 0x13, 0x20, 0x01, 0x28, 0x05, 0x48, 0x0d, 0x52, 0x0f, 0x61,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x46, 0x6c, 0x61, 0x67, 0x73, 0x88, 0x01,
	0x01, 0x12, 0x30, 0x0a, 0x11, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f,
	0x68, 0x69, 0x64, 0x64, 0x65, 0x6e, 0x18, 0x14, 0x20, 0x01, 0x28, 0x08, 0x48, 0x0e, 0x52, 0x10,
	0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x48, 0x69, 0x64, 0x64, 0x65, 0x6e,
	0x88, 0x01, 0x01, 0x12, 0x31, 0x0a, 0x12, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e,
	0x74, 0x5f, 0x6c, 0x69, 0x6e, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x15, 0x20, 0x01, 0x28, 0x05, 0x48,
	0x0f, 0x52, 0x10, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x4c, 0x69, 0x6e,
	0x6b, 0x49, 0x64, 0x88, 0x01, 0x01, 0x12, 0x28, 0x0a, 0x0d, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68,
	0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x16, 0x20, 0x01, 0x28, 0x05, 0x48, 0x10, 0x52,
	0x0c, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x88, 0x01, 0x01,
	0x12, 0x2b, 0x0a, 0x0f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x6d, 0x69, 0x6d, 0x65, 0x5f,
	0x74, 0x61, 0x67, 0x18, 0x17, 0x20, 0x01, 0x28, 0x09, 0x48, 0x11, 0x52, 0x0d, 0x61, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x4d, 0x69, 0x6d, 0x65, 0x54, 0x61, 0x67, 0x88, 0x01, 0x01, 0x12, 0x28, 0x0a,
	0x0d, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x18,
	0x20, 0x01, 0x28, 0x05, 0x48, 0x12, 0x52, 0x0c, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x4e, 0x75,
	0x6d, 0x62, 0x65, 0x72, 0x88, 0x01, 0x01, 0x12, 0x2c, 0x0a, 0x0f, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x19, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x13, 0x52, 0x0e, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x50, 0x61, 0x74, 0x68, 0x6e, 0x61,
	0x6d, 0x65, 0x88, 0x01, 0x01, 0x12, 0x24, 0x0a, 0x0b, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f,
	0x73, 0x69, 0x7a, 0x65, 0x18, 0x1b, 0x20, 0x01, 0x28, 0x05, 0x48, 0x14, 0x52, 0x0a, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x53, 0x69, 0x7a, 0x65, 0x88, 0x01, 0x01, 0x12, 0x37, 0x0a, 0x15, 0x61,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x1d, 0x20, 0x01, 0x28, 0x09, 0x48, 0x15, 0x52, 0x13, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x4e, 0x61, 0x6d,
	0x65, 0x88, 0x01, 0x01, 0x12, 0x3b, 0x0a, 0x17, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x61, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x72, 0x73, 0x65, 0x74, 0x18,
	0x1f, 0x20, 0x01, 0x28, 0x09, 0x48, 0x16, 0x52, 0x15, 0x74, 0x65, 0x78, 0x74, 0x41, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x68, 0x61, 0x72, 0x73, 0x65, 0x74, 0x88, 0x01,
	0x01, 0x42, 0x1e, 0x0a, 0x1c, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74,
	0x5f, 0x6d, 0x61, 0x63, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x42, 0x26, 0x0a, 0x24, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74,
	0x5f, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x5f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x42, 0x1d, 0x0a, 0x1b, 0x5f, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x42, 0x1b, 0x0a, 0x19, 0x5f, 0x61, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x42, 0x16, 0x0a, 0x14, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68,
	0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x62, 0x61, 0x73, 0x65, 0x42, 0x14, 0x0a,
	0x12, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x5f, 0x69, 0x64, 0x42, 0x1a, 0x0a, 0x18, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42,
	0x13, 0x0a, 0x11, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e,
	0x73, 0x69, 0x6f, 0x6e, 0x42, 0x12, 0x0a, 0x10, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f,
	0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x0f, 0x0a, 0x0d, 0x5f, 0x61, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x5f, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x42, 0x17, 0x0a, 0x15, 0x5f, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61,
	0x6d, 0x65, 0x42, 0x17, 0x0a, 0x15, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x6c, 0x6f,
	0x6e, 0x67, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x1b, 0x0a, 0x19, 0x5f,
	0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61,
	0x63, 0x74, 0x5f, 0x70, 0x68, 0x6f, 0x74, 0x6f, 0x42, 0x13, 0x0a, 0x11, 0x5f, 0x61, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x42, 0x14, 0x0a,
	0x12, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x68, 0x69, 0x64,
	0x64, 0x65, 0x6e, 0x42, 0x15, 0x0a, 0x13, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65,
	0x6e, 0x74, 0x5f, 0x6c, 0x69, 0x6e, 0x6b, 0x5f, 0x69, 0x64, 0x42, 0x10, 0x0a, 0x0e, 0x5f, 0x61,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x42, 0x12, 0x0a, 0x10,
	0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x6d, 0x69, 0x6d, 0x65, 0x5f, 0x74, 0x61, 0x67,
	0x42, 0x10, 0x0a, 0x0e, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x6e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x42, 0x12, 0x0a, 0x10, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x70, 0x61,
	0x74, 0x68, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x42, 0x18, 0x0a, 0x16, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x42, 0x1a, 0x0a, 0x18, 0x5f, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68,
	0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x72, 0x73, 0x65, 0x74, 0x42, 0x28, 0x5a, 0x26,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6f, 0x6f, 0x69, 0x6a,
	0x74, 0x65, 0x63, 0x68, 0x2f, 0x67, 0x6f, 0x2d, 0x70, 0x73, 0x74, 0x3b, 0x70, 0x72, 0x6f, 0x70,
	0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_attachment_proto_rawDescOnce sync.Once
	file_attachment_proto_rawDescData = file_attachment_proto_rawDesc
)

func file_attachment_proto_rawDescGZIP() []byte {
	file_attachment_proto_rawDescOnce.Do(func() {
		file_attachment_proto_rawDescData = protoimpl.X.CompressGZIP(file_attachment_proto_rawDescData)
	})
	return file_attachment_proto_rawDescData
}

var file_attachment_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_attachment_proto_goTypes = []interface{}{
	(*Attachment)(nil), // 0: Attachment
}
var file_attachment_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_attachment_proto_init() }
func file_attachment_proto_init() {
	if File_attachment_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_attachment_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Attachment); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_attachment_proto_msgTypes[0].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_attachment_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_attachment_proto_goTypes,
		DependencyIndexes: file_attachment_proto_depIdxs,
		MessageInfos:      file_attachment_proto_msgTypes,
	}.Build()
	File_attachment_proto = out.File
	file_attachment_proto_rawDesc = nil
	file_attachment_proto_goTypes = nil
	file_attachment_proto_depIdxs = nil
}
