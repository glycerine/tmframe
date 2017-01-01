package testdata

// NOTE: THIS FILE WAS PRODUCED BY THE
// ZEBRAPACK CODE GENERATION TOOL (github.com/glycerine/zebrapack)
// DO NOT EDIT

import "github.com/glycerine/zebrapack/msgp"

// DecodeMsg implements msgp.Decodable
// We treat empty fields as if we read a Nil from the wire.
func (z *LogEntry) DecodeMsg(dc *msgp.Reader) (err error) {
	var sawTopNil bool
	if dc.IsNil() {
		sawTopNil = true
		err = dc.ReadNil()
		if err != nil {
			return
		}
		dc.PushAlwaysNil()
	}

	var field []byte
	_ = field
	const maxFields0zzad = 3

	// -- templateDecodeMsgZid starts here--
	var totalEncodedFields0zzad uint32
	totalEncodedFields0zzad, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	encodedFieldsLeft0zzad := totalEncodedFields0zzad
	missingFieldsLeft0zzad := maxFields0zzad - totalEncodedFields0zzad

	var nextMiss0zzad int = -1
	var found0zzad [maxFields0zzad]bool
	var curField0zzad int

doneWithStruct0zzad:
	// First fill all the encoded fields, then
	// treat the remaining, missing fields, as Nil.
	for encodedFieldsLeft0zzad > 0 || missingFieldsLeft0zzad > 0 {
		//fmt.Printf("encodedFieldsLeft: %v, missingFieldsLeft: %v, found: '%v', fields: '%#v'\n", encodedFieldsLeft0zzad, missingFieldsLeft0zzad, msgp.ShowFound(found0zzad[:]), decodeMsgFieldOrder0zzad)
		if encodedFieldsLeft0zzad > 0 {
			encodedFieldsLeft0zzad--
			curField0zzad, err = dc.ReadInt()
			if err != nil {
				return
			}
		} else {
			//missing fields need handling
			if nextMiss0zzad < 0 {
				// tell the reader to only give us Nils
				// until further notice.
				dc.PushAlwaysNil()
				nextMiss0zzad = 0
			}
			for nextMiss0zzad < maxFields0zzad && (found0zzad[nextMiss0zzad] || decodeMsgFieldSkip0zzad[nextMiss0zzad]) {
				nextMiss0zzad++
			}
			if nextMiss0zzad == maxFields0zzad {
				// filled all the empty fields!
				break doneWithStruct0zzad
			}
			missingFieldsLeft0zzad--
			curField0zzad = nextMiss0zzad
		}
		//fmt.Printf("switching on curField: '%v'\n", curField0zzad)
		switch curField0zzad {
		// -- templateDecodeMsgZid ends here --

		case 0:
			// zid 0 for "lsn"
			found0zzad[0] = true
			z.LogSequenceNum, err = dc.ReadInt64()
			if err != nil {
				panic(err)
			}
		case 1:
			// zid 1 for "op"
			found0zzad[1] = true
			z.Operation, err = dc.ReadString()
			if err != nil {
				panic(err)
			}
		case 2:
			// zid 2 for "args"
			found0zzad[2] = true
			var zxmc uint32
			zxmc, err = dc.ReadMapHeader()
			if err != nil {
				panic(err)
			}
			if z.OpArgs == nil && zxmc > 0 {
				z.OpArgs = make(map[string]string, zxmc)
			} else if len(z.OpArgs) > 0 {
				for key, _ := range z.OpArgs {
					delete(z.OpArgs, key)
				}
			}
			for zxmc > 0 {
				zxmc--
				var znks string
				var zxfd string
				znks, err = dc.ReadString()
				if err != nil {
					panic(err)
				}
				zxfd, err = dc.ReadString()
				if err != nil {
					panic(err)
				}
				z.OpArgs[znks] = zxfd
			}
		default:
			err = dc.Skip()
			if err != nil {
				panic(err)
			}
		}
	}
	if nextMiss0zzad != -1 {
		dc.PopAlwaysNil()
	}

	if sawTopNil {
		dc.PopAlwaysNil()
	}

	return
}

// fields of LogEntry
var decodeMsgFieldOrder0zzad = []string{"lsn", "op", "args"}

var decodeMsgFieldSkip0zzad = []bool{false, false, false}

// fieldsNotEmpty supports omitempty tags
func (z *LogEntry) fieldsNotEmpty(isempty []bool) uint32 {
	if len(isempty) == 0 {
		return 3
	}
	var fieldsInUse uint32 = 3
	isempty[0] = (z.LogSequenceNum == 0) // number, omitempty
	if isempty[0] {
		fieldsInUse--
	}
	isempty[1] = (len(z.Operation) == 0) // string, omitempty
	if isempty[1] {
		fieldsInUse--
	}
	isempty[2] = (len(z.OpArgs) == 0) // string, omitempty
	if isempty[2] {
		fieldsInUse--
	}

	return fieldsInUse
}

// EncodeMsg implements msgp.Encodable
func (z *LogEntry) EncodeMsg(en *msgp.Writer) (err error) {

	// honor the omitempty tags
	var empty_zaww [3]bool
	fieldsInUse_zkzs := z.fieldsNotEmpty(empty_zaww[:])

	// map header
	err = en.WriteMapHeader(fieldsInUse_zkzs + 1)
	if err != nil {
		return err
	}

	// runtime struct type identification for 'LogEntry'
	err = en.Append(0xff)
	if err != nil {
		return err
	}
	err = en.WriteStringFromBytes([]byte{0x4c, 0x6f, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79})
	if err != nil {
		return err
	}

	if !empty_zaww[0] {
		// zid 0 for "lsn"
		err = en.Append(0x0)
		if err != nil {
			return err
		}
		err = en.WriteInt64(z.LogSequenceNum)
		if err != nil {
			panic(err)
		}
	}

	if !empty_zaww[1] {
		// zid 1 for "op"
		err = en.Append(0x1)
		if err != nil {
			return err
		}
		err = en.WriteString(z.Operation)
		if err != nil {
			panic(err)
		}
	}

	if !empty_zaww[2] {
		// zid 2 for "args"
		err = en.Append(0x2)
		if err != nil {
			return err
		}
		err = en.WriteMapHeader(uint32(len(z.OpArgs)))
		if err != nil {
			panic(err)
		}
		for znks, zxfd := range z.OpArgs {
			err = en.WriteString(znks)
			if err != nil {
				panic(err)
			}
			err = en.WriteString(zxfd)
			if err != nil {
				panic(err)
			}
		}
	}

	return
}

// MarshalMsg implements msgp.Marshaler
func (z *LogEntry) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())

	// honor the omitempty tags
	var empty [3]bool
	fieldsInUse := z.fieldsNotEmpty(empty[:])
	o = msgp.AppendMapHeader(o, fieldsInUse+1)

	// runtime struct type identification for 'LogEntry'
	o = msgp.AppendNegativeOneAndStringAsBytes(o, []byte{0x4c, 0x6f, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79})

	if !empty[0] {
		// zid 0 for "lsn"
		o = append(o, 0x0)
		o = msgp.AppendInt64(o, z.LogSequenceNum)
	}

	if !empty[1] {
		// zid 1 for "op"
		o = append(o, 0x1)
		o = msgp.AppendString(o, z.Operation)
	}

	if !empty[2] {
		// zid 2 for "args"
		o = append(o, 0x2)
		o = msgp.AppendMapHeader(o, uint32(len(z.OpArgs)))
		for znks, zxfd := range z.OpArgs {
			o = msgp.AppendString(o, znks)
			o = msgp.AppendString(o, zxfd)
		}
	}

	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *LogEntry) UnmarshalMsg(bts []byte) (o []byte, err error) {
	return z.UnmarshalMsgWithCfg(bts, nil)
}
func (z *LogEntry) UnmarshalMsgWithCfg(bts []byte, cfg *msgp.RuntimeConfig) (o []byte, err error) {
	var nbs msgp.NilBitsStack
	nbs.Init(cfg)
	var sawTopNil bool
	if msgp.IsNil(bts) {
		sawTopNil = true
		bts = nbs.PushAlwaysNil(bts[1:])
	}

	var field []byte
	_ = field
	const maxFields1zefq = 3

	// -- templateUnmarshalMsgZid starts here--
	var totalEncodedFields1zefq uint32
	if !nbs.AlwaysNil {
		totalEncodedFields1zefq, bts, err = nbs.ReadMapHeaderBytes(bts)
		if err != nil {
			panic(err)
			return
		}
	}
	encodedFieldsLeft1zefq := totalEncodedFields1zefq
	missingFieldsLeft1zefq := maxFields1zefq - totalEncodedFields1zefq

	var nextMiss1zefq int = -1
	var found1zefq [maxFields1zefq]bool
	var curField1zefq int

doneWithStruct1zefq:
	// First fill all the encoded fields, then
	// treat the remaining, missing fields, as Nil.
	for encodedFieldsLeft1zefq > 0 || missingFieldsLeft1zefq > 0 {
		//fmt.Printf("encodedFieldsLeft: %v, missingFieldsLeft: %v, found: '%v', fields: '%#v'\n", encodedFieldsLeft1zefq, missingFieldsLeft1zefq, msgp.ShowFound(found1zefq[:]), unmarshalMsgFieldOrder1zefq)
		if encodedFieldsLeft1zefq > 0 {
			encodedFieldsLeft1zefq--
			curField1zefq, bts, err = nbs.ReadIntBytes(bts)
			if err != nil {
				panic(err)
				return
			}
		} else {
			//missing fields need handling
			if nextMiss1zefq < 0 {
				// set bts to contain just mnil (0xc0)
				bts = nbs.PushAlwaysNil(bts)
				nextMiss1zefq = 0
			}
			for nextMiss1zefq < maxFields1zefq && (found1zefq[nextMiss1zefq] || unmarshalMsgFieldSkip1zefq[nextMiss1zefq]) {
				nextMiss1zefq++
			}
			if nextMiss1zefq == maxFields1zefq {
				// filled all the empty fields!
				break doneWithStruct1zefq
			}
			missingFieldsLeft1zefq--
			curField1zefq = nextMiss1zefq
		}
		//fmt.Printf("switching on curField: '%v'\n", curField1zefq)
		switch curField1zefq {
		// -- templateUnmarshalMsgZid ends here --

		case 0:
			// zid 0 for "lsn"
			found1zefq[0] = true
			z.LogSequenceNum, bts, err = nbs.ReadInt64Bytes(bts)

			if err != nil {
				panic(err)
			}
		case 1:
			// zid 1 for "op"
			found1zefq[1] = true
			z.Operation, bts, err = nbs.ReadStringBytes(bts)

			if err != nil {
				panic(err)
			}
		case 2:
			// zid 2 for "args"
			found1zefq[2] = true
			if nbs.AlwaysNil {
				if len(z.OpArgs) > 0 {
					for key, _ := range z.OpArgs {
						delete(z.OpArgs, key)
					}
				}

			} else {

				var zbav uint32
				zbav, bts, err = nbs.ReadMapHeaderBytes(bts)
				if err != nil {
					panic(err)
				}
				if z.OpArgs == nil && zbav > 0 {
					z.OpArgs = make(map[string]string, zbav)
				} else if len(z.OpArgs) > 0 {
					for key, _ := range z.OpArgs {
						delete(z.OpArgs, key)
					}
				}
				for zbav > 0 {
					var znks string
					var zxfd string
					zbav--
					znks, bts, err = nbs.ReadStringBytes(bts)
					if err != nil {
						panic(err)
					}
					zxfd, bts, err = nbs.ReadStringBytes(bts)

					if err != nil {
						panic(err)
					}
					z.OpArgs[znks] = zxfd
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				panic(err)
			}
		}
	}
	if nextMiss1zefq != -1 {
		bts = nbs.PopAlwaysNil()
	}

	if sawTopNil {
		bts = nbs.PopAlwaysNil()
	}
	o = bts
	return
}

// fields of LogEntry
var unmarshalMsgFieldOrder1zefq = []string{"lsn", "op", "args"}

var unmarshalMsgFieldSkip1zefq = []bool{false, false, false}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *LogEntry) Msgsize() (s int) {
	s = 1 + 12 + msgp.Int64Size + 12 + msgp.StringPrefixSize + len(z.Operation) + 12 + msgp.MapHeaderSize
	if z.OpArgs != nil {
		for znks, zxfd := range z.OpArgs {
			_ = zxfd
			_ = znks
			s += msgp.StringPrefixSize + len(znks) + msgp.StringPrefixSize + len(zxfd)
		}
	}
	return
}

// ZebraSchemaInMsgpack2Format provides the ZebraPack Schema in msgpack2 format, length 573 bytes
func ZebraSchemaInMsgpack2Format() []byte { return zebraSchemaInMsgpack2Format }

var zebraSchemaInMsgpack2Format = []byte{
	0x85, 0xaa, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x50, 0x61, 0x74, 0x68, 0xab, 0x6c, 0x6f, 0x67,
	0x65, 0x6e, 0x74, 0x72, 0x79, 0x2e, 0x67, 0x6f, 0xad, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x50,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0xa8, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0xad,
	0x5a, 0x65, 0x62, 0x72, 0x61, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x49, 0x64, 0xd3, 0x0, 0x0,
	0xa9, 0x56, 0x5e, 0xd3, 0x24, 0x17, 0xa7, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x73, 0x81, 0xa8,
	0x4c, 0x6f, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x82, 0xaa, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74,
	0x4e, 0x61, 0x6d, 0x65, 0xa8, 0x4c, 0x6f, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0xa6, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x73, 0x93, 0x87, 0xa3, 0x5a, 0x69, 0x64, 0x0, 0xab, 0x46, 0x69, 0x65, 0x6c,
	0x64, 0x47, 0x6f, 0x4e, 0x61, 0x6d, 0x65, 0xae, 0x4c, 0x6f, 0x67, 0x53, 0x65, 0x71, 0x75, 0x65,
	0x6e, 0x63, 0x65, 0x4e, 0x75, 0x6d, 0xac, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x54, 0x61, 0x67, 0x4e,
	0x61, 0x6d, 0x65, 0xa3, 0x6c, 0x73, 0x6e, 0xac, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x54, 0x79, 0x70,
	0x65, 0x53, 0x74, 0x72, 0xa5, 0x69, 0x6e, 0x74, 0x36, 0x34, 0xad, 0x46, 0x69, 0x65, 0x6c, 0x64,
	0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x17, 0xae, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x50,
	0x72, 0x69, 0x6d, 0x69, 0x74, 0x69, 0x76, 0x65, 0x11, 0xad, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x46,
	0x75, 0x6c, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x82, 0xa4, 0x4b, 0x69, 0x6e, 0x64, 0x11, 0xa3, 0x53,
	0x74, 0x72, 0xa5, 0x69, 0x6e, 0x74, 0x36, 0x34, 0x88, 0xa3, 0x5a, 0x69, 0x64, 0x1, 0xab, 0x46,
	0x69, 0x65, 0x6c, 0x64, 0x47, 0x6f, 0x4e, 0x61, 0x6d, 0x65, 0xa9, 0x4f, 0x70, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0xac, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x54, 0x61, 0x67, 0x4e, 0x61, 0x6d,
	0x65, 0xa2, 0x6f, 0x70, 0xac, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x54, 0x79, 0x70, 0x65, 0x53, 0x74,
	0x72, 0xa6, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xad, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x43, 0x61,
	0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x17, 0xae, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x50, 0x72, 0x69,
	0x6d, 0x69, 0x74, 0x69, 0x76, 0x65, 0x2, 0xad, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x46, 0x75, 0x6c,
	0x6c, 0x54, 0x79, 0x70, 0x65, 0x82, 0xa4, 0x4b, 0x69, 0x6e, 0x64, 0x2, 0xa3, 0x53, 0x74, 0x72,
	0xa6, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xa9, 0x4f, 0x6d, 0x69, 0x74, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0xc3, 0x87, 0xa3, 0x5a, 0x69, 0x64, 0x2, 0xab, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x47, 0x6f,
	0x4e, 0x61, 0x6d, 0x65, 0xa6, 0x4f, 0x70, 0x41, 0x72, 0x67, 0x73, 0xac, 0x46, 0x69, 0x65, 0x6c,
	0x64, 0x54, 0x61, 0x67, 0x4e, 0x61, 0x6d, 0x65, 0xa4, 0x61, 0x72, 0x67, 0x73, 0xac, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x54, 0x79, 0x70, 0x65, 0x53, 0x74, 0x72, 0xb1, 0x6d, 0x61, 0x70, 0x5b, 0x73,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x5d, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xad, 0x46, 0x69, 0x65,
	0x6c, 0x64, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x18, 0xad, 0x46, 0x69, 0x65, 0x6c,
	0x64, 0x46, 0x75, 0x6c, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x84, 0xa4, 0x4b, 0x69, 0x6e, 0x64, 0x18,
	0xa3, 0x53, 0x74, 0x72, 0xa3, 0x4d, 0x61, 0x70, 0xa6, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x82,
	0xa4, 0x4b, 0x69, 0x6e, 0x64, 0x2, 0xa3, 0x53, 0x74, 0x72, 0xa6, 0x73, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0xa5, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x82, 0xa4, 0x4b, 0x69, 0x6e, 0x64, 0x2, 0xa3, 0x53,
	0x74, 0x72, 0xa6, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xa9, 0x4f, 0x6d, 0x69, 0x74, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0xc3, 0xa7, 0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x90,
}

// ZebraSchemaInJsonCompact provides the ZebraPack Schema in compact JSON format, length 738 bytes
func ZebraSchemaInJsonCompact() []byte { return zebraSchemaInJsonCompact }

var zebraSchemaInJsonCompact = []byte(`{"SourcePath":"logentry.go","SourcePackage":"testdata","ZebraSchemaId":186188423177239,"Structs":{"LogEntry":{"StructName":"LogEntry","Fields":[{"Zid":0,"FieldGoName":"LogSequenceNum","FieldTagName":"lsn","FieldTypeStr":"int64","FieldCategory":23,"FieldPrimitive":17,"FieldFullType":{"Kind":17,"Str":"int64"}},{"Zid":1,"FieldGoName":"Operation","FieldTagName":"op","FieldTypeStr":"string","FieldCategory":23,"FieldPrimitive":2,"FieldFullType":{"Kind":2,"Str":"string"},"OmitEmpty":true},{"Zid":2,"FieldGoName":"OpArgs","FieldTagName":"args","FieldTypeStr":"map[string]string","FieldCategory":24,"FieldFullType":{"Kind":24,"Str":"Map","Domain":{"Kind":2,"Str":"string"},"Range":{"Kind":2,"Str":"string"}},"OmitEmpty":true}]}},"Imports":[]}`)

// ZebraSchemaInJsonPretty provides the ZebraPack Schema in pretty JSON format, length 1861 bytes
func ZebraSchemaInJsonPretty() []byte { return zebraSchemaInJsonPretty }

var zebraSchemaInJsonPretty = []byte(`{
    "SourcePath": "logentry.go",
    "SourcePackage": "testdata",
    "ZebraSchemaId": 186188423177239,
    "Structs": {
        "LogEntry": {
            "StructName": "LogEntry",
            "Fields": [
                {
                    "Zid": 0,
                    "FieldGoName": "LogSequenceNum",
                    "FieldTagName": "lsn",
                    "FieldTypeStr": "int64",
                    "FieldCategory": 23,
                    "FieldPrimitive": 17,
                    "FieldFullType": {
                        "Kind": 17,
                        "Str": "int64"
                    }
                },
                {
                    "Zid": 1,
                    "FieldGoName": "Operation",
                    "FieldTagName": "op",
                    "FieldTypeStr": "string",
                    "FieldCategory": 23,
                    "FieldPrimitive": 2,
                    "FieldFullType": {
                        "Kind": 2,
                        "Str": "string"
                    },
                    "OmitEmpty": true
                },
                {
                    "Zid": 2,
                    "FieldGoName": "OpArgs",
                    "FieldTagName": "args",
                    "FieldTypeStr": "map[string]string",
                    "FieldCategory": 24,
                    "FieldFullType": {
                        "Kind": 24,
                        "Str": "Map",
                        "Domain": {
                            "Kind": 2,
                            "Str": "string"
                        },
                        "Range": {
                            "Kind": 2,
                            "Str": "string"
                        }
                    },
                    "OmitEmpty": true
                }
            ]
        }
    },
    "Imports": []
}`)
