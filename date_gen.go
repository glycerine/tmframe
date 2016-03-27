package tm

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Date) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Year":
			z.Year, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "Month":
			z.Month, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "Day":
			z.Day, err = dc.ReadInt()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Date) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Year"
	err = en.Append(0x83, 0xa4, 0x59, 0x65, 0x61, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Year)
	if err != nil {
		return
	}
	// write "Month"
	err = en.Append(0xa5, 0x4d, 0x6f, 0x6e, 0x74, 0x68)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Month)
	if err != nil {
		return
	}
	// write "Day"
	err = en.Append(0xa3, 0x44, 0x61, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Day)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Date) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Year"
	o = append(o, 0x83, 0xa4, 0x59, 0x65, 0x61, 0x72)
	o = msgp.AppendInt(o, z.Year)
	// string "Month"
	o = append(o, 0xa5, 0x4d, 0x6f, 0x6e, 0x74, 0x68)
	o = msgp.AppendInt(o, z.Month)
	// string "Day"
	o = append(o, 0xa3, 0x44, 0x61, 0x79)
	o = msgp.AppendInt(o, z.Day)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Date) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Year":
			z.Year, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "Month":
			z.Month, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "Day":
			z.Day, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z Date) Msgsize() (s int) {
	s = 1 + 5 + msgp.IntSize + 6 + msgp.IntSize + 4 + msgp.IntSize
	return
}
