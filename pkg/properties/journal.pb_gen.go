package properties

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Journal) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "3457711":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentPosted")
					return
				}
				z.LogDocumentPosted = nil
			} else {
				if z.LogDocumentPosted == nil {
					z.LogDocumentPosted = new(bool)
				}
				*z.LogDocumentPosted, err = dc.ReadBool()
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentPosted")
					return
				}
			}
		case "3457411":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentPrinted")
					return
				}
				z.LogDocumentPrinted = nil
			} else {
				if z.LogDocumentPrinted == nil {
					z.LogDocumentPrinted = new(bool)
				}
				*z.LogDocumentPrinted, err = dc.ReadBool()
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentPrinted")
					return
				}
			}
		case "3457611":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentRouted")
					return
				}
				z.LogDocumentRouted = nil
			} else {
				if z.LogDocumentRouted == nil {
					z.LogDocumentRouted = new(bool)
				}
				*z.LogDocumentRouted, err = dc.ReadBool()
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentRouted")
					return
				}
			}
		case "3457511":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentSaved")
					return
				}
				z.LogDocumentSaved = nil
			} else {
				if z.LogDocumentSaved == nil {
					z.LogDocumentSaved = new(bool)
				}
				*z.LogDocumentSaved, err = dc.ReadBool()
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentSaved")
					return
				}
			}
		case "345673":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogDuration")
					return
				}
				z.LogDuration = nil
			} else {
				if z.LogDuration == nil {
					z.LogDuration = new(int32)
				}
				*z.LogDuration, err = dc.ReadInt32()
				if err != nil {
					err = msgp.WrapError(err, "LogDuration")
					return
				}
			}
		case "3456864":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogEnd")
					return
				}
				z.LogEnd = nil
			} else {
				if z.LogEnd == nil {
					z.LogEnd = new(int64)
				}
				*z.LogEnd, err = dc.ReadInt64()
				if err != nil {
					err = msgp.WrapError(err, "LogEnd")
					return
				}
			}
		case "345723":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogFlags")
					return
				}
				z.LogFlags = nil
			} else {
				if z.LogFlags == nil {
					z.LogFlags = new(int32)
				}
				*z.LogFlags, err = dc.ReadInt32()
				if err != nil {
					err = msgp.WrapError(err, "LogFlags")
					return
				}
			}
		case "3456664":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogStart")
					return
				}
				z.LogStart = nil
			} else {
				if z.LogStart == nil {
					z.LogStart = new(int64)
				}
				*z.LogStart, err = dc.ReadInt64()
				if err != nil {
					err = msgp.WrapError(err, "LogStart")
					return
				}
			}
		case "3456031":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogType")
					return
				}
				z.LogType = nil
			} else {
				if z.LogType == nil {
					z.LogType = new(string)
				}
				*z.LogType, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "LogType")
					return
				}
			}
		case "3457831":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "LogTypeDesc")
					return
				}
				z.LogTypeDesc = nil
			} else {
				if z.LogTypeDesc == nil {
					z.LogTypeDesc = new(string)
				}
				*z.LogTypeDesc, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "LogTypeDesc")
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Journal) EncodeMsg(en *msgp.Writer) (err error) {
	// omitempty: check for empty values
	zb0001Len := uint32(10)
	var zb0001Mask uint16 /* 10 bits */
	_ = zb0001Mask
	if z.LogDocumentPosted == nil {
		zb0001Len--
		zb0001Mask |= 0x1
	}
	if z.LogDocumentPrinted == nil {
		zb0001Len--
		zb0001Mask |= 0x2
	}
	if z.LogDocumentRouted == nil {
		zb0001Len--
		zb0001Mask |= 0x4
	}
	if z.LogDocumentSaved == nil {
		zb0001Len--
		zb0001Mask |= 0x8
	}
	if z.LogDuration == nil {
		zb0001Len--
		zb0001Mask |= 0x10
	}
	if z.LogEnd == nil {
		zb0001Len--
		zb0001Mask |= 0x20
	}
	if z.LogFlags == nil {
		zb0001Len--
		zb0001Mask |= 0x40
	}
	if z.LogStart == nil {
		zb0001Len--
		zb0001Mask |= 0x80
	}
	if z.LogType == nil {
		zb0001Len--
		zb0001Mask |= 0x100
	}
	if z.LogTypeDesc == nil {
		zb0001Len--
		zb0001Mask |= 0x200
	}
	// variable map header, size zb0001Len
	err = en.Append(0x80 | uint8(zb0001Len))
	if err != nil {
		return
	}
	if zb0001Len == 0 {
		return
	}
	if (zb0001Mask & 0x1) == 0 { // if not empty
		// write "3457711"
		err = en.Append(0xa7, 0x33, 0x34, 0x35, 0x37, 0x37, 0x31, 0x31)
		if err != nil {
			return
		}
		if z.LogDocumentPosted == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteBool(*z.LogDocumentPosted)
			if err != nil {
				err = msgp.WrapError(err, "LogDocumentPosted")
				return
			}
		}
	}
	if (zb0001Mask & 0x2) == 0 { // if not empty
		// write "3457411"
		err = en.Append(0xa7, 0x33, 0x34, 0x35, 0x37, 0x34, 0x31, 0x31)
		if err != nil {
			return
		}
		if z.LogDocumentPrinted == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteBool(*z.LogDocumentPrinted)
			if err != nil {
				err = msgp.WrapError(err, "LogDocumentPrinted")
				return
			}
		}
	}
	if (zb0001Mask & 0x4) == 0 { // if not empty
		// write "3457611"
		err = en.Append(0xa7, 0x33, 0x34, 0x35, 0x37, 0x36, 0x31, 0x31)
		if err != nil {
			return
		}
		if z.LogDocumentRouted == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteBool(*z.LogDocumentRouted)
			if err != nil {
				err = msgp.WrapError(err, "LogDocumentRouted")
				return
			}
		}
	}
	if (zb0001Mask & 0x8) == 0 { // if not empty
		// write "3457511"
		err = en.Append(0xa7, 0x33, 0x34, 0x35, 0x37, 0x35, 0x31, 0x31)
		if err != nil {
			return
		}
		if z.LogDocumentSaved == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteBool(*z.LogDocumentSaved)
			if err != nil {
				err = msgp.WrapError(err, "LogDocumentSaved")
				return
			}
		}
	}
	if (zb0001Mask & 0x10) == 0 { // if not empty
		// write "345673"
		err = en.Append(0xa6, 0x33, 0x34, 0x35, 0x36, 0x37, 0x33)
		if err != nil {
			return
		}
		if z.LogDuration == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteInt32(*z.LogDuration)
			if err != nil {
				err = msgp.WrapError(err, "LogDuration")
				return
			}
		}
	}
	if (zb0001Mask & 0x20) == 0 { // if not empty
		// write "3456864"
		err = en.Append(0xa7, 0x33, 0x34, 0x35, 0x36, 0x38, 0x36, 0x34)
		if err != nil {
			return
		}
		if z.LogEnd == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteInt64(*z.LogEnd)
			if err != nil {
				err = msgp.WrapError(err, "LogEnd")
				return
			}
		}
	}
	if (zb0001Mask & 0x40) == 0 { // if not empty
		// write "345723"
		err = en.Append(0xa6, 0x33, 0x34, 0x35, 0x37, 0x32, 0x33)
		if err != nil {
			return
		}
		if z.LogFlags == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteInt32(*z.LogFlags)
			if err != nil {
				err = msgp.WrapError(err, "LogFlags")
				return
			}
		}
	}
	if (zb0001Mask & 0x80) == 0 { // if not empty
		// write "3456664"
		err = en.Append(0xa7, 0x33, 0x34, 0x35, 0x36, 0x36, 0x36, 0x34)
		if err != nil {
			return
		}
		if z.LogStart == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteInt64(*z.LogStart)
			if err != nil {
				err = msgp.WrapError(err, "LogStart")
				return
			}
		}
	}
	if (zb0001Mask & 0x100) == 0 { // if not empty
		// write "3456031"
		err = en.Append(0xa7, 0x33, 0x34, 0x35, 0x36, 0x30, 0x33, 0x31)
		if err != nil {
			return
		}
		if z.LogType == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteString(*z.LogType)
			if err != nil {
				err = msgp.WrapError(err, "LogType")
				return
			}
		}
	}
	if (zb0001Mask & 0x200) == 0 { // if not empty
		// write "3457831"
		err = en.Append(0xa7, 0x33, 0x34, 0x35, 0x37, 0x38, 0x33, 0x31)
		if err != nil {
			return
		}
		if z.LogTypeDesc == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = en.WriteString(*z.LogTypeDesc)
			if err != nil {
				err = msgp.WrapError(err, "LogTypeDesc")
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Journal) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// omitempty: check for empty values
	zb0001Len := uint32(10)
	var zb0001Mask uint16 /* 10 bits */
	_ = zb0001Mask
	if z.LogDocumentPosted == nil {
		zb0001Len--
		zb0001Mask |= 0x1
	}
	if z.LogDocumentPrinted == nil {
		zb0001Len--
		zb0001Mask |= 0x2
	}
	if z.LogDocumentRouted == nil {
		zb0001Len--
		zb0001Mask |= 0x4
	}
	if z.LogDocumentSaved == nil {
		zb0001Len--
		zb0001Mask |= 0x8
	}
	if z.LogDuration == nil {
		zb0001Len--
		zb0001Mask |= 0x10
	}
	if z.LogEnd == nil {
		zb0001Len--
		zb0001Mask |= 0x20
	}
	if z.LogFlags == nil {
		zb0001Len--
		zb0001Mask |= 0x40
	}
	if z.LogStart == nil {
		zb0001Len--
		zb0001Mask |= 0x80
	}
	if z.LogType == nil {
		zb0001Len--
		zb0001Mask |= 0x100
	}
	if z.LogTypeDesc == nil {
		zb0001Len--
		zb0001Mask |= 0x200
	}
	// variable map header, size zb0001Len
	o = append(o, 0x80|uint8(zb0001Len))
	if zb0001Len == 0 {
		return
	}
	if (zb0001Mask & 0x1) == 0 { // if not empty
		// string "3457711"
		o = append(o, 0xa7, 0x33, 0x34, 0x35, 0x37, 0x37, 0x31, 0x31)
		if z.LogDocumentPosted == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendBool(o, *z.LogDocumentPosted)
		}
	}
	if (zb0001Mask & 0x2) == 0 { // if not empty
		// string "3457411"
		o = append(o, 0xa7, 0x33, 0x34, 0x35, 0x37, 0x34, 0x31, 0x31)
		if z.LogDocumentPrinted == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendBool(o, *z.LogDocumentPrinted)
		}
	}
	if (zb0001Mask & 0x4) == 0 { // if not empty
		// string "3457611"
		o = append(o, 0xa7, 0x33, 0x34, 0x35, 0x37, 0x36, 0x31, 0x31)
		if z.LogDocumentRouted == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendBool(o, *z.LogDocumentRouted)
		}
	}
	if (zb0001Mask & 0x8) == 0 { // if not empty
		// string "3457511"
		o = append(o, 0xa7, 0x33, 0x34, 0x35, 0x37, 0x35, 0x31, 0x31)
		if z.LogDocumentSaved == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendBool(o, *z.LogDocumentSaved)
		}
	}
	if (zb0001Mask & 0x10) == 0 { // if not empty
		// string "345673"
		o = append(o, 0xa6, 0x33, 0x34, 0x35, 0x36, 0x37, 0x33)
		if z.LogDuration == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendInt32(o, *z.LogDuration)
		}
	}
	if (zb0001Mask & 0x20) == 0 { // if not empty
		// string "3456864"
		o = append(o, 0xa7, 0x33, 0x34, 0x35, 0x36, 0x38, 0x36, 0x34)
		if z.LogEnd == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendInt64(o, *z.LogEnd)
		}
	}
	if (zb0001Mask & 0x40) == 0 { // if not empty
		// string "345723"
		o = append(o, 0xa6, 0x33, 0x34, 0x35, 0x37, 0x32, 0x33)
		if z.LogFlags == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendInt32(o, *z.LogFlags)
		}
	}
	if (zb0001Mask & 0x80) == 0 { // if not empty
		// string "3456664"
		o = append(o, 0xa7, 0x33, 0x34, 0x35, 0x36, 0x36, 0x36, 0x34)
		if z.LogStart == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendInt64(o, *z.LogStart)
		}
	}
	if (zb0001Mask & 0x100) == 0 { // if not empty
		// string "3456031"
		o = append(o, 0xa7, 0x33, 0x34, 0x35, 0x36, 0x30, 0x33, 0x31)
		if z.LogType == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendString(o, *z.LogType)
		}
	}
	if (zb0001Mask & 0x200) == 0 { // if not empty
		// string "3457831"
		o = append(o, 0xa7, 0x33, 0x34, 0x35, 0x37, 0x38, 0x33, 0x31)
		if z.LogTypeDesc == nil {
			o = msgp.AppendNil(o)
		} else {
			o = msgp.AppendString(o, *z.LogTypeDesc)
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Journal) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "3457711":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogDocumentPosted = nil
			} else {
				if z.LogDocumentPosted == nil {
					z.LogDocumentPosted = new(bool)
				}
				*z.LogDocumentPosted, bts, err = msgp.ReadBoolBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentPosted")
					return
				}
			}
		case "3457411":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogDocumentPrinted = nil
			} else {
				if z.LogDocumentPrinted == nil {
					z.LogDocumentPrinted = new(bool)
				}
				*z.LogDocumentPrinted, bts, err = msgp.ReadBoolBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentPrinted")
					return
				}
			}
		case "3457611":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogDocumentRouted = nil
			} else {
				if z.LogDocumentRouted == nil {
					z.LogDocumentRouted = new(bool)
				}
				*z.LogDocumentRouted, bts, err = msgp.ReadBoolBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentRouted")
					return
				}
			}
		case "3457511":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogDocumentSaved = nil
			} else {
				if z.LogDocumentSaved == nil {
					z.LogDocumentSaved = new(bool)
				}
				*z.LogDocumentSaved, bts, err = msgp.ReadBoolBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogDocumentSaved")
					return
				}
			}
		case "345673":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogDuration = nil
			} else {
				if z.LogDuration == nil {
					z.LogDuration = new(int32)
				}
				*z.LogDuration, bts, err = msgp.ReadInt32Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogDuration")
					return
				}
			}
		case "3456864":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogEnd = nil
			} else {
				if z.LogEnd == nil {
					z.LogEnd = new(int64)
				}
				*z.LogEnd, bts, err = msgp.ReadInt64Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogEnd")
					return
				}
			}
		case "345723":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogFlags = nil
			} else {
				if z.LogFlags == nil {
					z.LogFlags = new(int32)
				}
				*z.LogFlags, bts, err = msgp.ReadInt32Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogFlags")
					return
				}
			}
		case "3456664":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogStart = nil
			} else {
				if z.LogStart == nil {
					z.LogStart = new(int64)
				}
				*z.LogStart, bts, err = msgp.ReadInt64Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogStart")
					return
				}
			}
		case "3456031":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogType = nil
			} else {
				if z.LogType == nil {
					z.LogType = new(string)
				}
				*z.LogType, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogType")
					return
				}
			}
		case "3457831":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogTypeDesc = nil
			} else {
				if z.LogTypeDesc == nil {
					z.LogTypeDesc = new(string)
				}
				*z.LogTypeDesc, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "LogTypeDesc")
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Journal) Msgsize() (s int) {
	s = 1 + 8
	if z.LogDocumentPosted == nil {
		s += msgp.NilSize
	} else {
		s += msgp.BoolSize
	}
	s += 8
	if z.LogDocumentPrinted == nil {
		s += msgp.NilSize
	} else {
		s += msgp.BoolSize
	}
	s += 8
	if z.LogDocumentRouted == nil {
		s += msgp.NilSize
	} else {
		s += msgp.BoolSize
	}
	s += 8
	if z.LogDocumentSaved == nil {
		s += msgp.NilSize
	} else {
		s += msgp.BoolSize
	}
	s += 7
	if z.LogDuration == nil {
		s += msgp.NilSize
	} else {
		s += msgp.Int32Size
	}
	s += 8
	if z.LogEnd == nil {
		s += msgp.NilSize
	} else {
		s += msgp.Int64Size
	}
	s += 7
	if z.LogFlags == nil {
		s += msgp.NilSize
	} else {
		s += msgp.Int32Size
	}
	s += 8
	if z.LogStart == nil {
		s += msgp.NilSize
	} else {
		s += msgp.Int64Size
	}
	s += 8
	if z.LogType == nil {
		s += msgp.NilSize
	} else {
		s += msgp.StringPrefixSize + len(*z.LogType)
	}
	s += 8
	if z.LogTypeDesc == nil {
		s += msgp.NilSize
	} else {
		s += msgp.StringPrefixSize + len(*z.LogTypeDesc)
	}
	return
}
