package tm

import (
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	"github.com/nats-io/nats"
	"os"
	"testing"
)

func Test050MergeSortToChannel(t *testing.T) {
	cv.Convey("Given that we have a directory of subject-sorted and invidually time-sorted TMFRAME files, when we merge sort them they should appear in chronological order with subjects interleaved.", t, func() {

		datestr := "2016/03/10"
		dir := "testdata/" + datestr

		// setup the test streams, with duplicated Frames
		nFrame := 100

		frames, _, _ := GenTestTwo64Frames(nFrame, nil)

		// prep to write out files to files a, b, or c
		lab := []string{"a", "b", "c"}
		expected := []int64{}

		fds := []*os.File{}
		wri := []*FrameWriter{}
		for j := range lab {
			writeMe, err := os.Create(dir + "/" + lab[j])
			panicOn(err)
			ds := NewFrameWriter(writeMe, 64*1024)
			fds = append(fds, writeMe)
			wri = append(wri, ds)
		}

		nLab := len(lab)
		for i := 0; i < nFrame; i++ {
			k := int64(cryptoRandNonNegInt() % nLab)
			frames[i].SetV1(k)
			p("frames[i=%v] = '%v'", i, frames[i].String())
			wri[k].Append(frames[i])
			expected = append(expected, k)
		}

		for j := range lab {
			wri[j].Flush()
			wri[j].Sync()
			fds[j].Close()
		}

		defer func() {
			// cleanup
			for j := range lab {
				os.Remove(dir + "/" + lab[j])
			}
		}()

		// test the merge sorting to channel
		inputFiles, err := GetProperFilesInDir(dir)
		ch := make(chan *nats.Msg)

		go func() {
			err = SendDirOnChannel(ch, inputFiles, dir, datestr)
			panicOn(err)
		}()

		var fr Frame
		for i, expV1 := range expected {
			obs := <-ch
			_, err := fr.Unmarshal(obs.Data, false)
			p("i: %v, exp: %v, fr observed: %v", i, expected[i], fr.String())
			panicOn(err)
			obsV1 := fr.GetV1()
			cv.So(obsV1, cv.ShouldEqual, expV1)
		}
	})
}

func SendDirOnChannel(ch chan *nats.Msg, inputFiles []string, dir string, datestr string) error {

	n := len(inputFiles)
	if n == 0 {
		return fmt.Errorf("no input files given")
	}

	const MB = 1024 * 1024

	outputStream := NewFrameChWriter(MB, ch)

	strms := make([]*BufferedFrameReader, n)
	for i := 0; i < n; i++ {
		pth := dir + "/" + inputFiles[i]
		if !FileExists(pth) {
			return fmt.Errorf("path '%s' not found", pth)
		}
		f, err := os.Open(pth)
		if err != nil {
			return fmt.Errorf("could not open path '%s': '%s'", pth, err)
		}
		strms[i] = NewBufferedFrameReader(f, MB, inputFiles[i])
	}

	// okay, now create and merge streams
	return outputStream.Merge(datestr, strms...)
}
