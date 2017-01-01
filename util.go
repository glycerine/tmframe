package tm

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/glycerine/tmframe/testdata"
)

// generate n test Frames, with 4 different frame types, and randomly varying sizes
// if outpath is non-nill, write to that file.
func GenTestFrames(n int, outpath *string) (frames []*Frame, tms []time.Time, by []byte) {

	t0, err := time.Parse(time.RFC3339, "2016-02-16T00:00:00Z")
	panicOn(err)
	t0 = t0.UTC()

	var f0 *Frame
	for i := 0; i < n; i++ {
		t := t0.Add(time.Second * time.Duration(i))
		tms = append(tms, t)
		switch i % 3 {
		case 0:
			// generate a random length message payload
			m := cryptoRandNonNegInt() % 254
			data := make([]byte, m)
			for j := 0; j < m; j++ {
				data[j] = byte(j)
			}
			f0, err = NewFrame(t, EvMsgpKafka, 0, 0, data)
			panicOn(err)
		case 1:
			f0, err = NewFrame(t, EvZero, 0, 0, nil)
			panicOn(err)
		case 2:
			f0, err = NewFrame(t, EvTwo64, float64(i), int64(i), nil)
			panicOn(err)
		case 3:
			f0, err = NewFrame(t, EvOneFloat64, float64(i), 0, nil)
			panicOn(err)
		}
		frames = append(frames, f0)
		b0, err := f0.Marshal(nil)
		panicOn(err)
		by = append(by, b0...)
	}

	if outpath != nil {
		f, err := os.Create(*outpath)
		panicOn(err)
		_, err = f.Write(by)
		panicOn(err)
		f.Close()
	}

	return
}

func cryptoRandNonNegInt() int {
	b := make([]byte, 8)
	_, err := cryptorand.Read(b)
	panicOn(err)
	r := int(binary.LittleEndian.Uint64(b))
	if r < 0 {
		return -r
	}
	return r
}

// generate 0..(n-1) as floating point EvOneFloat64 frames
func GenTestFramesSequence(n int, outpath *string) (frames []*Frame, tms []time.Time, by []byte) {

	t0, err := time.Parse(time.RFC3339, "2016-02-16T00:00:00Z")
	panicOn(err)
	t0 = t0.UTC()

	var f0 *Frame
	for i := 0; i < n; i++ {
		t := t0.Add(time.Second * time.Duration(i))
		tms = append(tms, t)
		f0, err = NewFrame(t, EvOneFloat64, float64(i), 0, nil)
		panicOn(err)
		frames = append(frames, f0)
		b0, err := f0.Marshal(nil)
		panicOn(err)
		by = append(by, b0...)
	}

	if outpath != nil {
		f, err := os.Create(*outpath)
		panicOn(err)
		_, err = f.Write(by)
		panicOn(err)
		f.Close()
	}

	return
}

func GenerateSeriesWithRepeats(reps []int) *Series {

	var frames []*Frame
	t0, err := time.Parse(time.RFC3339, "2016-02-16T00:00:00Z")
	panicOn(err)
	t0 = t0.UTC()

	var f0 *Frame
	for i := range reps {
		j := reps[i]
		jtm := t0.Add(time.Second * time.Duration(i))
		for k := 0; k < j; k++ {
			f0, err = NewFrame(jtm, EvOneFloat64, float64(i), 0, nil)
			panicOn(err)
			frames = append(frames, f0)
		}
	}
	return NewSeriesFromFrames(frames)
}

// MakeTwo64Frames creates a slice of *Frames,
// starting at tm, and occuring thereafter as
// determined by the offset in the sec (seconds)
// slice.
//
// The deltas argument determines the length and the
// content of the slice. Each element is
// an EvOneFloat64 valued Frame, use Frame.GetV0()
// to observe the value. deltas of 0 are skipped.
// typ gives the int64 V1 value for each frame,
// accessed with Frame.GetV1().
func MakeTwo64Frames(tm time.Time,
	deltas []float64,
	typ []int64,
	addSec []int64) []*Frame {

	frames := make([]*Frame, 0)
	t0 := tm.UTC()

	var f0 *Frame
	var err error
	for i, val := range deltas {
		if val != 0 {
			t := t0.Add(time.Second * time.Duration(addSec[i]))
			f0, err = NewFrame(t, EvTwo64, val, typ[i], nil)
			panicOn(err)
			frames = append(frames, f0)
		}
	}

	return frames
}

// GetProperFilesInDir returns a slice
// listing the files (not directories) found in dir.
func GetProperFilesInDir(dir string) ([]string, error) {

	f, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("GetProperFilesInDir() could not open '%s': %v", dir, err)
	}
	defer f.Close()

	fiSlice, err := f.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("GetProperFilesInDir() could not Readdir on '%s': %v", dir, err)
	}
	res := []string{}
	for _, fi := range fiSlice {
		if !fi.IsDir() {
			res = append(res, fi.Name())
		}
	}
	return res, nil
}

func chompPathSep(path string) string {
	if strings.HasSuffix(path, sep) {
		return path[:len(path)-1]
	}
	return path
}

var sep = string(os.PathSeparator)

func ReadAvailDays(readDir string) (res []string, err error) {
	q("top of ReadAvailDays")
	defer func() {
		q("leaving ReadAvailDays with err = %v, res = %#v", err, res)
	}()

	root := chompPathSep(readDir)
	if root == "" || !DirExists(root) {
		return nil, fmt.Errorf("bad ReadDir '%s'", root)
	}

	years, err := GetDateSubdirs(root)
	if err != nil {
		q("finding years, GetDateSubdirs returned err = %v", err)
		return nil, err
	}
	q("years = %#v", years)

	// for each year, get the months
	for _, y := range sorted(years) {
		months, err := GetDateSubdirs(root + sep + y)
		if err != nil {
			return nil, err
		}
		q("months = %#v", months)

		// for each month, get the days
		for _, m := range sorted(months) {
			days, err := GetDateSubdirs(root + sep + y + sep + m)
			if err != nil {
				return nil, err
			}
			q("days = %#v", days)

			for _, d := range sorted(days) {
				res = append(res, y+sep+m+sep+d)
			}
		}

	}
	q("res = '%#v'", res)
	return res, nil
}

// IntersectDays: avail must already be in sorted calendar increasing order.
// endx can be nil, begin cannot be nil.
func IntersectDays(begin *Date, endx *Date, avail []string) (readDays []string, err error) {
	q("top of IntersectDays, begin = %v", begin.String())

	need := begin.String()
	if endx == nil {
		// only this one day, yes or no.
		for i := range avail {
			if avail[i] == need {
				q("found our one, returning just it since endx == nil")
				return []string{need}, nil
			}
		}
		// this day is not available
		q("our day is not available")
		return []string{}, nil
	}

	last := PrevDate(endx)
	q("last = %s", last.String())

	// scan for days >= being and < endx
	for _, a := range avail {
		ad, err := ParseDate(a)
		if err != nil {
			return nil, fmt.Errorf("could not ParseDate('%s'): %v", a, err)
		}
		if DateAfter(ad, last) {
			// can stop early, since avail is sorted ascending in calendar days
			break
		}
		// INVAR: ad <= last
		if DatesEqual(ad, begin) || DateAfter(ad, begin) {
			readDays = append(readDays, a)
		}
	}

	return readDays, nil
}

func GetDateSubdirs(dir string) ([]string, error) {
	q("top of GetDateSubdirs")
	f, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("GetDateSubdirs() could not open '%s': %v", dir, err)
	}
	defer f.Close()

	fiSlice, err := f.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("GetDateSubdirs() could not Readdir on '%s': %v", dir, err)
	}
	res := []string{}
	for _, fi := range fiSlice {
		nm := fi.Name()
		if fi.IsDir() && IsDateDir(nm) {
			res = append(res, fi.Name())
		}
	}
	return res, nil
}

func sorted(s []string) []string {
	sort.StringSlice(s).Sort()
	return s
}

// IsDateDir says true to '2016', '01', and '31',
// but rejects '1999', '41' or '00'.
func IsDateDir(d string) bool {
	n := len(d)
	if n != 4 && n != 2 {
		return false
	}
	for _, r := range d {
		if r > '9' || r < '0' {
			return false
		}
	}
	if n == 4 {
		if d[0] != '2' && d[0] != '3' {
			return false
		}
	}
	if n == 2 {
		if d[0] > '3' {
			return false
		}
		if d[0] == '3' {
			if d[1] > '1' {
				return false
			}
		}
	}
	return true
}

// generate n test Two64Frames
func GenTestTwo64Frames(n int, outpath *string) (frames []*Frame, tms []time.Time, by []byte) {

	t0, err := time.Parse(time.RFC3339, "2016-02-16T00:00:00Z")
	panicOn(err)
	t0 = t0.UTC()

	var f0 *Frame
	for i := 0; i < n; i++ {
		t := t0.Add(time.Second * time.Duration(i))
		tms = append(tms, t)
		f0, err = NewFrame(t, EvTwo64, float64(i), 0, nil)
		panicOn(err)
		frames = append(frames, f0)
		b0, err := f0.Marshal(nil)
		panicOn(err)
		by = append(by, b0...)
	}

	if outpath != nil {
		f, err := os.Create(*outpath)
		panicOn(err)
		_, err = f.Write(by)
		panicOn(err)
		f.Close()
	}

	return
}

// generate n test Frames of ZebraPack encoded.
// if outpath is non-nill, write to that file.
func GenTestdataZebraPackTestFrames(n int, outpath *string) (frames []*Frame, tms []time.Time, by []byte) {

	t0, err := time.Parse(time.RFC3339, "2016-02-16T00:00:00Z")
	panicOn(err)
	t0 = t0.UTC()

	var f0 *Frame
	for i := 0; i < n; i++ {

		var e testdata.LogEntry
		e.LogSequenceNum = int64(i)
		e.Operation = fmt.Sprintf("0x%x", i)
		data, err := e.MarshalMsg(nil)
		panicOn(err)
		t := t0.Add(time.Second * time.Duration(i))
		tms = append(tms, t)
		f0, err = NewFrame(t, EvZebraPack, 0, 0, data)
		panicOn(err)

		frames = append(frames, f0)
		b0, err := f0.Marshal(nil)

		panicOn(err)
		by = append(by, b0...)
	}

	if outpath != nil {
		if *outpath == "-" {
			_, err = os.Stdout.Write(by)
			panicOn(err)
		} else {
			f, err := os.Create(*outpath)
			panicOn(err)
			_, err = f.Write(by)
			panicOn(err)
			f.Close()
		}
	}

	return
}
