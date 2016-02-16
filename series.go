package tm

import (
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

func (s *Series) InForceBefore(tm time.Time) (*Frame, SearchStatus) {
	return s.Frames[0], Avail
}
