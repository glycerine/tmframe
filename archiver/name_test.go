package archiver

import (
	cv "github.com/glycerine/goconvey/convey"
	"testing"
)

func Test005StreamNameRegexp(t *testing.T) {

	cv.Convey("given a nats subect 'ServiceName.archive.mystream', we should be able to extract the stream name 'mystream' from it", t, func() {
		var stream, id string

		stream, id = ExtractStreamFromSubject(ServiceName + ".archive.mystream")
		cv.So(stream, cv.ShouldEqual, "mystream")
		cv.So(id, cv.ShouldEqual, "")
		stream, id = ExtractStreamFromSubject(ServiceName + ".archive.my-stream")
		cv.So(stream, cv.ShouldEqual, "my-stream")
		cv.So(id, cv.ShouldEqual, "")

		stream, id = ExtractStreamFromSubject(ServiceName + ".archive.123.456.789-10")
		cv.So(stream, cv.ShouldEqual, "123")
		cv.So(id, cv.ShouldEqual, "456.789-10")

		stream, id = ExtractStreamFromSubject(ServiceName + ".archive.")
		cv.So(stream, cv.ShouldEqual, "")
		cv.So(id, cv.ShouldEqual, "")
		stream, id = ExtractStreamFromSubject(ServiceName + ".archive")
		cv.So(stream, cv.ShouldEqual, "")
		cv.So(id, cv.ShouldEqual, "")
		stream, id = ExtractStreamFromSubject("wrong.blah.hello.")
		cv.So(stream, cv.ShouldEqual, "")
		cv.So(id, cv.ShouldEqual, "")

	})
}
