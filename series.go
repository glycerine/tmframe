package tm

import (
	"fmt"
	"sort"
	"time"
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
	InFuture SearchStatus = 2
)

func (s SearchStatus) String() string {
	switch s {
	case InPast:
		return "InPast"
	case Avail:
		return "Avail"
	case InFuture:
		return "InFuture"
	}
	panic(fmt.Sprintf("unknown SearchStatus %v", s))
}

// If tm is greater than any seen Frame, InForceBefore()
// will return the last seen Frame and a SearchStatus of InFuture.
// If tm is smaller than the oldest Frame available,
// InForceBefore will return (nil, InPast). Otherwise,
// it returns the Frame where Frame.Tm() is strictly before
// the tm (using 10 nanosecond resolution; truncating tm using
// the TimeToPrimTm(tm) function. The 3rd return argument
// is the integer index of the returned frame, or -1 if
// SearchStatus is InPast.
func (s *Series) InForceBefore(tm time.Time) (*Frame, SearchStatus, int) {

	m := len(s.Frames)
	utm := TimeToPrimTm(tm)
	// Search returns the smallest index i in [0, n) at which f(i) is true
	i := sort.Search(m, func(i int) bool {
		//P("sort called at i=%v, returning %v b/c %v vs %v", i, s.Frames[i].Tm() >= utm, s.Frames[i].Tm(), utm)
		return s.Frames[i].Tm() >= utm
	})
	//P("i = %v", i)
	if i == m {
		return s.Frames[m-1], InFuture, m - 1
	}

	if i == 0 {
		return nil, InPast, -1
	}

	return s.Frames[i-1], Avail, i - 1
}

func (s *Series) FirstAtOrBefore(tm time.Time) (*Frame, SearchStatus, int) {

	m := len(s.Frames)
	utm := TimeToPrimTm(tm)
	// Search returns the smallest index i in [0, n) at which f(i) is true.
	// If i == n, this means no such index had f(i) true.
	i := sort.Search(m, func(i int) bool {
		P("FirstAtOrBefore sort called at i=%v, returning %v b/c %v vs %v", i, s.Frames[i].Tm() >= utm, s.Frames[i].Tm(), utm)
		return s.Frames[i].Tm() >= utm
	})
	P("FirstAtOrBefore i = %v", i)
	if i == m {
		// all frames Tm < utm
		rtm := s.Frames[m-1].Tm()
		// spin back to the first at rtm
		k := m - 1
		for k >= 1 && s.Frames[k-1].Tm() == rtm {
			P("spinning down to k=%v", k-1)
			k--
		}
		return s.Frames[k], InFuture, k
	}
	// INVAR: at least one entry had Tm >= utm

	itm := s.Frames[i].Tm()
	if i == 0 {
		if itm == utm {
			return s.Frames[i], Avail, i
		}
		// even s.Frames[0] was > utm
		return nil, InPast, -1
	}
	if itm == utm {
		return s.Frames[i], Avail, i
	}
	return s.Frames[i-1], Avail, i - 1
}

func (s *Series) LastAtOrBefore(tm time.Time) (*Frame, SearchStatus, int) {

	m := len(s.Frames)
	utm := TimeToPrimTm(tm)
	// Search returns the smallest index i in [0, n) at which f(i) is true
	i := sort.Search(m, func(i int) bool {
		//P("sort called at i=%v, returning %v b/c %v vs %v", i, s.Frames[i].Tm() >= utm, s.Frames[i].Tm(), utm)
		return s.Frames[i].Tm() >= utm
	})
	//P("i = %v", i)
	if i == m {
		return s.Frames[m-1], InFuture, m - 1
	}

	if i == 0 {
		return nil, InPast, -1
	}

	return s.Frames[i-1], Avail, i - 1
}
