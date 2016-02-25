package tm

import (
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	"os"
	"os/exec"
	"testing"
)

func Test020MergeSortStreams(t *testing.T) {
	cv.Convey(`given several streams with interleaved Prim timestamps, Merge() should combine them in time-sorted order`, t, func() {

		// setup the test streams, with interleaved Frames
		nFrame := 100
		expectedPath := "test.mergesort.expected"
		frames, _, _ := GenTestFramesSequence(nFrame, &expectedPath)

		// deal into different piles, randomly
		nPile := 3
		fds := make([]*os.File, nPile)
		for i := 0; i < nPile; i++ {
			fd, err := os.Create(fmt.Sprintf("test.merge.input.%v", i))
			panicOn(err)
			fds[i] = fd
		}
		for i := 0; i < nFrame; i++ {
			k := cryptoRandNonNegInt() % nPile
			b0, err := frames[i].Marshal(nil)
			panicOn(err)
			_, err = fds[k].Write(b0)
			panicOn(err)
		}

		strms := make([]*BufferedFrameReader, nPile)
		for i := 0; i < nPile; i++ {
			fds[i].Sync()
			fds[i].Seek(0, 0)
			strms[i] = NewBufferedFrameReader(fds[i], 64*1024)
		}

		obsPath := "test.merge.observed"
		writeMe, err := os.Create(obsPath)
		panicOn(err)
		outputStream := NewFrameWriter(writeMe, 64*1024)

		// okay, now create and merge streams
		err = outputStream.Merge(strms...)
		panicOn(err)
		outputStream.Sync()

		if FilesDiff(expectedPath, obsPath) {
			panic(fmt.Errorf("merge failed, '%s' != '%s'", obsPath, expectedPath))
		}
	})
}

func FilesDiff(a, b string) bool {
	co, err := exec.Command("diff", a, b).CombinedOutput()
	p("co = %v", string(co))
	p("err = %v", err)
	if err != nil {
		// don't panic, diff returns 2 on differences
	}

	return len(co) != 0
}
