package tm

import (
	"sort"
	"time"
	//"github.com/glycerine/rbuf"
)

type Series struct {
	Frames []*Frame
}

func NewSeriesFromFrames(fr []*Frame) *Series {
	s := &Series{
		Frames: fr,
	}
	return s
}

// result of searching with JustBefore() and AtOrAfter()
type SearchStatus int

const (
	InPast   SearchStatus = 0
	Avail    SearchStatus = 1
	InFuture SearchStatus = 1
)

// If tm is greater than any seen Frame, InForceBefore()
// will return the last seen Frame and a SearchStatus of InFuture.
// If tm is smaller than the oldest Frame available,
// InForceBefore will return (nil, InPast). Otherwise,
// it returns the Frame where Frame.Tm() is strictly before
// the tm (using 10 nanosecond resolution; truncating tm using
// the TimeToPrimTm(tm) function.
func (s *Series) InForceBefore(tm time.Time) (*Frame, SearchStatus) {

	m := len(s.Frames)
	utm := TimeToPrimTm(tm)
	// Search returns the smallest index i in [0, n) at which f(i) is true
	i := sort.Search(m, func(i int) bool {
		//P("sort called at i=%v, returning %v b/c %v vs %v", i, s.Frames[i].Tm() >= utm, s.Frames[i].Tm(), utm)
		return s.Frames[i].Tm() >= utm
	})
	//P("i = %v", i)
	if i == m {
		return s.Frames[m-1], InFuture
	}

	if i == 0 {
		return nil, InPast
	}

	return s.Frames[i-1], Avail
}
