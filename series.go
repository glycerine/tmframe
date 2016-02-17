package tm

import (
	"fmt"
	"sort"
	"time"
)

// Series represents a set of sequential Frames,
// which, taken in order, represent points in a timeseries.
//
// FirstInForceBefore(), LastInForceBefore(),
// FirstAtOrBefore(), and LastAtOrBefore() are the
// main methods to search for a specific timepoint,
// or the last event in force before a specific timepoint.
//
// See the LastInForceBefore() for the most detailed
// description of the arguments and return values.
// The other three functions are analogous.
//
type Series struct {
	Frames []*Frame
}

// create a new Series from a set of Frame pointers
func NewSeriesFromFrames(fr []*Frame) *Series {
	s := &Series{
		Frames: fr,
	}
	return s
}

// SearchStatus is returned by the four search functions
// FirstInForceBefore(), LastInForceBefore(),
// FirstAtOrBefore(), and LastAtOrBefore() to indicate
// the result of the search.
type SearchStatus int

const (
	InPast   SearchStatus = 0
	Avail    SearchStatus = 1
	InFuture SearchStatus = 2
)

// Stringify the SearchStatus, for printing.
func (s SearchStatus) String() string {
	switch s {
	case InPast:
		return "InPast"
	case Avail:
		return "Avail"
	case InFuture:
		return "InFuture"
	}
	panic(fmt.Sprintf("unknown SearchStatus %d", int(s)))
}

// LastInForceBefore():
//
// If tm is greater than any seen Frame, LastInForceBefore()
// will return the last seen Frame and a SearchStatus of InFuture.
//
// If tm is smaller than the oldest Frame available,
// LastInForceBefore will return (nil, InPast). Otherwise,
// it returns the Frame where Frame.Tm() is strictly before
// the tm (using 10 nanosecond resolution; truncating tm using
// the TimeToPrimTm(tm) function.
//
// The 3rd returned argument provides the integer index of
// the returned frame in s.Frames, or -1 if SearchStatus is InPast.
//
// In summary:
//
// LastInForceBefore(): looking at the ties for the nearest timestamp s < tm, return
// the most recent (last in the presented sequence order) of these ties at s. Nearest means
// that there is no other timestamp r such that s < r < tm.
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

// FirstInForceBefore(): looking at the ties for the nearest timestamp s < tm, return
// the earliest (first in the presented sequence order) of these ties at s. Nearest means
// that there is no other timestamp r such that s < r < tm.
func (s *Series) FirstInForceBefore(tm time.Time) (*Frame, SearchStatus, int) {

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
		// find the smallest index such that Tm >= rtm.
		k := sort.Search(m, func(i int) bool {
			return s.Frames[i].Tm() >= rtm
		})
		// k == m is impossible, rtm came from a Frame in s.Frames
		return s.Frames[k], InFuture, k
	}
	// INVAR: at least one entry had Tm >= utm

	if i == 0 {
		return nil, InPast, -1
	}

	// i is the smallest Frame such that itm >= utm.
	// Since we want to go strictly before that i,
	// start at j = i - 1; then find the first of any ties
	// at the Frames[j].Tm() timestamp. If we don't
	// find any, just return s.Frame[j].
	j := i - 1
	jtm := s.Frames[j].Tm()

	// Handling repeated timestamps:
	// Search foward to the last Frame at itm.
	// For worst case efficiency of O(log(n)), rather
	// than O(n), use Search() again to
	// find the smallest index such that Tm >= itm,
	// then subtract 1.
	k := sort.Search(m, func(i int) bool {
		return s.Frames[i].Tm() >= jtm
	})

	// k == m is impossible since jtm comes from the j Frame
	return s.Frames[k], Avail, k
}

// FirstAtOrBefore(): looking at the ties for the nearest timestamp s <= tm, return
// the earliest (first in the presented sequence order) of these ties at s.  Nearest means
// that there is no other timestamp r such that s <= r <= tm.
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
		// find the smallest index such that Tm >= rtm.
		k := sort.Search(m, func(i int) bool {
			return s.Frames[i].Tm() >= rtm
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

// LastAtOrBefore(): looking at the ties for the nearest timestamp s <= tm, return
// the newest (last in the presented sequence order) of these ties at timestamp s.  Nearest means
// that there is no other timestamp r such that s <= r <= tm.
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
