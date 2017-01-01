package tm

import (
	"bytes"
	"strings"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
	"github.com/glycerine/tmframe/testdata"
	"github.com/glycerine/zebrapack/zebra"
)

func Test060DisplayZebraPack(t *testing.T) {

	cv.Convey("DisplayFrame should handle ZebraPack if supplied with a schema\n", t, func() {
		// read the schema
		msgp2schema := testdata.ZebraSchemaInMsgpack2Format()
		var zSchema zebra.Schema
		_, err := zSchema.UnmarshalMsg(msgp2schema)
		panicOn(err)

		fn := "./test.zebrapack"
		panicOn(err)

		frs, _, _ := GenTestdataZebraPackTestFrames(5, &fn)

		prettyPrint := false
		skipPayload := false
		rReadable := false
		var out bytes.Buffer
		for i, fr := range frs {
			fr.DisplayFrame(&out, int64(i), prettyPrint, skipPayload, rReadable, &zSchema)
		}
		//fmt.Printf("out = '%s'\n", string(out.Bytes()))
		cv.So(strings.HasPrefix(string(out.Bytes()), `000000 TMFRAME 2016-02-16T00:00:00Z EVTNUM Ev.16 [33 bytes] (UCOUNT 17) {"op":"0x0"}`), cv.ShouldBeTrue)
	})
}
