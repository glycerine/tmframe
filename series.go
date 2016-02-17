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

// FirstInForceBefore(), LastInForceBefore(),
// FirstAtOrBefore(), and LastAtOrBefore()
// provide searching functionality for through a
// timeseries that may have duplicated timestamps.
//
// LastInForceBefore():
//
// If tm is greater than any seen Frame, InForceBefore()
// will return the last seen Frame and a SearchStatus of InFuture.
// If tm is smaller than the oldest Frame available,
// InForceBefore will return (nil, InPast). Otherwise,
// it returns the Frame where Frame.Tm() is strictly before
// the tm (using 10 nanosecond resolution; truncating tm using
// the TimeToPrimTm(tm) function. The 3rd return argument
// is the integer index of the returned frame, or -1 if
// SearchStatus is InPast.
func (s *Series) LastInForceBefore(tm time.Time) (*Frame, SearchStatus, int) {

	m := len(s.Frames)
	utm := TimeToPrimTm(tm)

	// Search returns the smallest index i in [0, m) at which f(i) is true.
	// If i == m, this means no such index had f(i) true.
	i := sort.Search(m, func(i int) bool {
		return s.Frames[i].Tm() >= utm
	})
	if i == m {
		return s.Frames[m-1], InFuture, m - 1
	}

	if i == 0 {
		return nil, InPast, -1
	}

	return s.Frames[i-1], Avail, i - 1
}

// Like LastInForceBefore(), but if there are ties we return
// the first in the sequence of tied timestamp Frames.
func (s *Series) FirstInForceBefore(tm time.Time) (*Frame, SearchStatus, int) {

	m := len(s.Frames)
	utm := TimeToPrimTm(tm)

	// Search returns the smallest index i in [0, m) at which f(i) is true.
	// If i == m, this means no such index had f(i) true.
	i := sort.Search(m, func(i int) bool {
		return s.Frames[i].Tm() >= utm
	})
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

	// Search returns the smallest index i in [0, m) at which f(i) is true.
	// If i == m, this means no such index had f(i) true.
	i := sort.Search(m, func(i int) bool {
		return s.Frames[i].Tm() >= utm
	})

	if i == m {
		// all frames Tm < utm
		rtm := s.Frames[m-1].Tm()

		// Handling repeated timestamps:
		// Need to search back to the first Frame at rtm.
		// For worst case efficiency of O(log(n)), rather
		// than O(n), use Search() again to
		// find the smallest index such that Tm == rtm.
		k := sort.Search(m, func(i int) bool {
			return s.Frames[i].Tm() == rtm
		})
		// k == m is impossible, rtm came from a Frame in s.Frames
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

	// Search returns the smallest index i in [0, m) at which f(i) is true.
	// If i == m, this means no such index had f(i) true.
	i := sort.Search(m, func(i int) bool {
		return s.Frames[i].Tm() >= utm
	})

	if i == m {
		// all frames Tm < utm
		return s.Frames[m-1], InFuture, m - 1
	}
	// INVAR: at least one entry had Tm >= utm.

	// i is the smallest Frame such that itm >= utm.
	itm := s.Frames[i].Tm()
	if i == 0 {
		if itm > utm {
			return nil, InPast, -1
		}
	}
	// But: there can be many at itm and we want the largest index that ties.

	// Handling repeated timestamps:
	// Search foward to the last Frame at itm.
	// For worst case efficiency of O(log(n)), rather
	// than O(n), use Search() again to
	// find the smallest index such that Tm > itm,
	// then subtract 1.
	k := sort.Search(m, func(i int) bool {
		return s.Frames[i].Tm() > itm
	})

	if k == m {
		// no Frames had Tm > itm
		return s.Frames[m-1], Avail, m - 1
	}
	// k == 0 is impossible, since itm came from an existing Frame in s.Frames.
	return s.Frames[k-1], Avail, k - 1
}
