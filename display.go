package tm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ugorji/go/codec"
	"io"
	"reflect"
	"time"
)

// display and pretty print message payloads in json/msgpack format.

// DisplayFrame prints a frame to w (e.g. pass os.Stdout as w),
// along with optional number i.
//
// If i < 0, the i is not printed. If prettyPrint is true and the payload
// is json or msgpack, we will display in an easier to ready pretty-printed
// json format. If skipPayload is true we will only print the Frame header
// information.
func (frame *Frame) DisplayFrame(w io.Writer, i int64, prettyPrint bool, skipPayload bool) {

	if i >= 0 {
		fmt.Fprintf(w, "%06d %s", i, frame.String())
	} else {
		fmt.Fprintf(w, "%s", frame.String())
	}
	if !skipPayload {
		evtnum := frame.GetEvtnum()
		if evtnum == EvJson || (evtnum >= 2000 && evtnum <= 9999) {
			pp := prettyPrintJson(prettyPrint, frame.Data)
			fmt.Fprintf(w, "  %s", string(pp))
		}
		if evtnum == EvMsgpKafka || evtnum == EvMsgpack {
			// decode msgpack to json with ugorji/go/codec

			var iface interface{}
			dec := codec.NewDecoderBytes(frame.Data, &msgpHelper.mh)
			err := dec.Decode(&iface)
			panicOn(err)

			//Q("iface = '%#v'", iface)

			var wbuf bytes.Buffer
			enc := codec.NewEncoder(&wbuf, &msgpHelper.jh)
			err = enc.Encode(&iface)
			panicOn(err)
			pp := prettyPrintJson(prettyPrint, wbuf.Bytes())
			fmt.Fprintf(w, " %s", string(pp))
		}
	}
	fmt.Fprintf(w, "\n")
}

// StringifyFrame is like DisplayFrame but it returns
// a string.
func (frame *Frame) Stringify(i int64, prettyPrint bool, skipPayload bool) string {
	var s string

	if i >= 0 {
		s += fmt.Sprintf("%06d %s", i, frame.String())
	} else {
		s += fmt.Sprintf("%s", frame.String())
	}
	if !skipPayload {
		evtnum := frame.GetEvtnum()
		if evtnum == EvJson {
			pp := prettyPrintJson(prettyPrint, frame.Data)
			s += fmt.Sprintf("  %s", string(pp))
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
			pp := prettyPrintJson(prettyPrint, w.Bytes())
			s += fmt.Sprintf(" %s", string(pp))
		}
	}
	return s
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

// TimeExt allows github.com/ugorji/go/codec to understand Go time.Time
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

// JsonBytesAsStringExt allows github.com/ugorji/go/codec to passthrough json bytes without conversion.
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
	//return v
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
