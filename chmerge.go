package tm

import (
	"github.com/nats-io/nats"
	"io"
	"sort"
)

// FrameChWriter provides merge-sort (via Merge) with
// output sent to a channel rather than an io.Writer.
// Output frames are sent on the SendOnMe channel.
// FrameChWriter may buffer frames and does not force
// sending immediately.
type FrameChWriter struct {
	Frames   []*Frame
	fr       *FrameReader
	buf      []byte
	SendOnMe chan *nats.Msg
}

// NewFrameToChannelWriter construts a new FrameChWriter for buffering
// and sending Frames on sendOnMe.  It imposes a
// message size limit of maxFrameBytes in order to size
// its internal marshalling buffer.
func NewFrameChWriter(maxFrameBytes int64, sendOnMe chan *nats.Msg) *FrameChWriter {

	s := &FrameChWriter{
		SendOnMe: sendOnMe,
		buf:      make([]byte, maxFrameBytes),
	}
	return s
}

// SendOnCh sends f on the fw.SendOnMe channel, using
// a nats.Msg with subject: "tseries.replay." + subject + "." + datestr.
// SendOnCh marshals the frame, effectively copying it.
func (fw *FrameChWriter) SendOnCh(f *Frame, subject string, datestr string) {
	fullSubj := "tseries.replay." + subject + "." + datestr
	by, err := f.Marshal(nil)
	panicOn(err)
	fw.SendOnMe <- &nats.Msg{Subject: fullSubj, Data: by}
}

// Merge merges the strms input into timestamp order, based on
// the Frame.Tm() timestamp, and writes the ordered sequence
// out to fw.SendOnMe
func (fw *FrameChWriter) Merge(datestr string, strms ...*BufferedFrameReader) error {

	n := len(strms)
	peeks := make([]*frameElem, n)
	for i := 0; i < n; i++ {
		peeks[i] = &frameElem{bfr: strms[i], index: i, name: strms[i].Name}
	}

	var err error

	// initialize the Frames in peeks
	newlist := []*frameElem{}
	for i := range peeks {
		peeks[i].frame, err = peeks[i].bfr.Peek()
		if err != nil {
			if err == io.EOF {
				// peeks[i].frame will be nil.
				//p("err = %v, omitting peeks[i=%v] from newlist", err, i)
			} else {
				return err
			}
		} else {
			newlist = append(newlist, peeks[i])
		}
	}
	peeks = newlist

	for len(peeks) > 0 {
		if len(peeks) == 1 {
			// just copy over the rest of this stream and we're done
			//p("down to just one (%v), copying it directly over", peeks[0].index)

			fw.SendOnCh(peeks[0].frame, peeks[0].name, datestr)
			var fr *Frame
			for err == nil {
				fr, err = peeks[0].bfr.ReadOne()
				if err != nil {
					fw.SendOnCh(fr, peeks[0].name, datestr)
				}
			}
			if err == io.EOF {
				return nil
			}
			return err
		}
		// have 2 or more source left, sort and pick the earliest
		sort.Sort(frameSorter(peeks))
		fw.SendOnCh(peeks[0].frame, peeks[0].name, datestr)
		peeks[0].bfr.Advance()
		peeks[0].frame, err = peeks[0].bfr.Peek()
		if err != nil {
			if err == io.EOF {
				//p("saw err '%v',  finished with stream %v", err, peeks[0].index)
				peeks = peeks[1:]
			} else {
				return err
			}
		}
	}
	return nil
}
