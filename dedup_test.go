package tm

import (
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	"os"
	"testing"
)

func Test030DedupStreams(t *testing.T) {
	cv.Convey(`given a stream with duplicate frames, Dedup() filter out duplicates within the given window`, t, func() {

		// setup the test streams, with duplicated Frames
		nFrame := 100
		expectedPath := "test.dedup.expected"

		frames, _, _ := GenTestFrames(nFrame, &expectedPath)

		// prep to write out duplicates
		dupsPath := "test.dups"
		writeMe, err := os.Create(dupsPath)
		panicOn(err)
		ds := NewFrameWriter(writeMe, 64*1024)

		// generate duplictes, randomly
		fold := 4
		tot := nFrame * fold
		k := 0
		for i := 0; i < nFrame; i++ {
			ds.Append(frames[i])
			k++
		}

		for i := 1; i <= tot; i++ {
			fr := frames[cryptoRandNonNegInt()%((i%(nFrame-1))+1)]
			ds.Append(fr)
			k++
		}
		ds.Flush()
		ds.Sync()
		writeMe.Seek(0, 0)
		hasDupsFd := writeMe

		obsPath := "test.dedup.obs"
		dedupFd, err := os.Create(obsPath)

		// dedup the stream
		// all of it, so we give the full total as the window
		p("k = %v", k)
		err = Dedup(hasDupsFd, dedupFd, k+1)
		cv.So(err, cv.ShouldEqual, nil)

		// check
		diff := FilesDiff(expectedPath, obsPath)
		if diff {
			panic(fmt.Errorf("dedup failed, '%s' != '%s'", obsPath, expectedPath))
		}
		cv.So(diff, cv.ShouldBeFalse)
		os.Remove(obsPath)
		os.Remove(expectedPath)
	})
}
