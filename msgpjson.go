package tm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ugorji/go/codec"
	"reflect"
	"time"
)

// display and pretty print json/msgpack

// print a frame
func DisplayFrame(frame *Frame, cfg *TfcatConfig, i int64) {

	if i >= 0 {
		fmt.Printf("%06d %v", i, frame)
	} else {
		fmt.Printf("%06d %v", frame)
	}
	if cfg == nil || !cfg.SkipPayload {
		evtnum := frame.GetEvtnum()
		if evtnum == EvJson {
			pp := prettyPrintJson(cfg != nil && cfg.PrettyPrint, frame.Data)
			fmt.Printf("  %s", string(pp))
		}
		if evtnum == EvMsgpKafka || evtnum == EvMsgpack {
			// decode msgpack to json with ugorji/go/codec

			var iface interface{}
			dec := codec.NewDecoderBytes(frame.Data, &msgpHelper.mh)
			err := dec.Decode(&iface)
			panicOn(err)

			//Q("iface = '%#v'", iface)

			var w bytes.Buffer
			enc := codec.NewEncoder(&w, &msgpHelper.jh)
			err = enc.Encode(&iface)
			panicOn(err)
			pp := prettyPrintJson(cfg != nil && cfg.PrettyPrint, w.Bytes())
			fmt.Printf(" %s", string(pp))
		}
	}
	fmt.Printf("\n")
}

func prettyPrintJson(doPretty bool, input []byte) []byte {
	if doPretty {
		var prettyBB bytes.Buffer
		jsErr := json.Indent(&prettyBB, input, "      ", "    ")
		if jsErr != nil {
			return input
		} else {
			return prettyBB.Bytes()
		}
	} else {
		return input
	}
}

// msgpack and json helper

type msgpackHelper struct {
	initialized bool
	mh          codec.MsgpackHandle
	jh          codec.JsonHandle
}

func (m *msgpackHelper) init() {
	if m.initialized {
		return
	}

	m.mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	// configure extensions
	// e.g. for msgpack, define functions and enable Time support for tag 1
	//does this make a differenece? m.mh.AddExt(reflect.TypeOf(time.Time{}), 1, timeEncExt, timeDecExt)
	m.mh.RawToString = true
	m.mh.WriteExt = true
	m.mh.SignedInteger = true
	m.mh.Canonical = true // sort maps before writing them

	timeTyp := reflect.TypeOf(time.Time{})
	var timeExt TimeExt
	m.mh.SetExt(timeTyp, 1, timeExt)

	// JSON
	m.jh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	m.jh.SignedInteger = true
	m.jh.Canonical = true // sort maps before writing them
	m.jh.SetExt(timeTyp, 1, timeExt)

	var jsonBytesExt JsonBytesAsStringExt
	m.jh.RawBytesExt = &jsonBytesExt
	m.initialized = true
}

var msgpHelper msgpackHelper

func init() {
	msgpHelper.init()
}

type TimeExt struct{}

func (x TimeExt) WriteExt(interface{}) []byte { panic("unsupported") }
func (x TimeExt) ReadExt(interface{}, []byte) { panic("unsupported") }
func (x TimeExt) ConvertExt(v interface{}) interface{} {
	switch v2 := v.(type) {
	case time.Time:
		return v2.UTC().UnixNano()
	case *time.Time:
		return v2.UTC().UnixNano()
	default:
		panic(fmt.Sprintf("unsupported format for time conversion: expecting time.Time; got %T", v))
	}
}
func (x TimeExt) UpdateExt(dest interface{}, v interface{}) {
	tt := dest.(*time.Time)
	switch v2 := v.(type) {
	case int64:
		*tt = time.Unix(0, v2).UTC()
	case uint64:
		*tt = time.Unix(0, int64(v2)).UTC()
	default:
		panic(fmt.Sprintf("unsupported format for time conversion: expecting int64/uint64; got %T", v))
	}
}

type JsonBytesAsStringExt struct{}

//func (x JsonBytesAsStringExt) WriteExt(interface{}) []byte { panic("unsupported") }
//func (x JsonBytesAsStringExt) ReadExt(interface{}, []byte) { panic("unsupported") }
func (x JsonBytesAsStringExt) ConvertExt(v interface{}) interface{} {
	//P("in JsonBytesAsStringExt.ConvertExt(): v is %T/val='%#v'", v, v)
	switch v2 := v.(type) {
	case []byte:
		//Q("v2 is []byte")
		return string(v2)
	default:
		panic(fmt.Sprintf("unsupported format for JsonBytesAsStringExt conversion: expecting []byte; got %T", v))
	}
	return v
}
func (x JsonBytesAsStringExt) UpdateExt(dest interface{}, v interface{}) {
	//Q("in JsonBytesAsStringExt.UpdateExt(): v is %T/val=%#v    dest is %T/val=%#v", v, v, dest, dest)

	tt := dest.(*[]byte)
	switch v2 := v.(type) {
	case []byte:
		*tt = v2
	default:
		panic(fmt.Sprintf("unsupported format for JsonBytesAsStringExt conversion: expecting []byte; got %T", v))
	}

}
