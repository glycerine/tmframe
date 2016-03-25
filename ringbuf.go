package tm

import "io"

// FrameRingBuf:
//
//  a fixed-size circular ring buffer of *Frame
//
type FrameRingBuf struct {
	A        []*Frame
	N        int // MaxView, the total size of A, whether or not in use.
	Beg      int // start of in-use data in A
	Readable int // number of *Frame available in A (in use)
}

// constructor. NewFrameRingBuf will allocate internally
// a slice of size n.
func NewFrameRingBuf(n int) *FrameRingBuf {
	r := &FrameRingBuf{
		N:        n,
		Beg:      0,
		Readable: 0,
	}
	r.A = make([]*Frame, n, n)

	return r
}

// TwoContig returns all readable *Frame, but in two separate slices,
// to avoid copying. The two slices are from the same buffer, but
// are not contiguous. Either or both may be empty slices.
func (b *FrameRingBuf) TwoContig(makeCopy bool) (first []*Frame, second []*Frame) {

	extent := b.Beg + b.Readable
	if extent <= b.N {
		// we fit contiguously in this buffer without wrapping to the other.
		// Let second stay an empty slice.
		return b.A[b.Beg:(b.Beg + b.Readable)], second
	}

	return b.A[b.Beg:b.N], b.A[0:(extent % b.N)]
}

// ReadFrames():
//
// from bytes.Buffer.Read(): Read reads the next len(p) *Frame
// from the buffer or until the buffer is drained. The return
// value n is the number of bytes read. If the buffer has no data
// to return, err is io.EOF (unless len(p) is zero); otherwise it is nil.
func (b *FrameRingBuf) ReadFrames(p []*Frame) (n int, err error) {
	return b.readAndMaybeAdvance(p, true)
}

// ReadWithoutAdvance(): if you want to Read the data and leave
// it in the buffer, so as to peek ahead for example.
func (b *FrameRingBuf) ReadWithoutAdvance(p []*Frame) (n int, err error) {
	return b.readAndMaybeAdvance(p, false)
}

func (b *FrameRingBuf) readAndMaybeAdvance(p []*Frame, doAdvance bool) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if b.Readable == 0 {
		return 0, io.EOF
	}
	extent := b.Beg + b.Readable
	if extent <= b.N {
		n += copy(p, b.A[b.Beg:extent])
	} else {
		n += copy(p, b.A[b.Beg:b.N])
		if n < len(p) {
			n += copy(p[n:], b.A[0:(extent%b.N)])
		}
	}
	if doAdvance {
		b.Advance(n)
	}
	return
}

//
// WriteFrames writes len(p) *Frame values from p to
// the underlying ring.
// It returns the number of Frames written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
//
func (b *FrameRingBuf) WriteFrames(p []*Frame) (n int, err error) {
	for {
		if len(p) == 0 {
			// nothing (left) to copy in; notice we shorten our
			// local copy p (below) as we read from it.
			return
		}

		writeCapacity := b.N - b.Readable
		if writeCapacity <= 0 {
			// we are all full up already.
			return n, io.ErrShortWrite
		}
		if len(p) > writeCapacity {
			err = io.ErrShortWrite
			// leave err set and
			// keep going, write what we can.
		}

		writeStart := (b.Beg + b.Readable) % b.N

		upperLim := intMin(writeStart+writeCapacity, b.N)

		k := copy(b.A[writeStart:upperLim], p)

		n += k
		b.Readable += k
		p = p[k:]

		// we can fill from b.A[0:something] from
		// p's remainder, so loop
	}
}

// Reset quickly forgets any data stored in the ring buffer.
func (b *FrameRingBuf) Reset() {
	b.Beg = 0
	b.Readable = 0
	for i := range b.A {
		b.A[i] = nil
	}
}

// Advance(): non-standard, but better than Next(),
// because we don't have to unwrap our buffer and pay the cpu time
// for the copy that unwrapping may need.
// Useful in conjuction/after ReadWithoutAdvance() above.
func (b *FrameRingBuf) Advance(n int) {
	if n <= 0 {
		return
	}
	if n > b.Readable {
		n = b.Readable
	}
	b.Readable -= n
	b.Beg = (b.Beg + n) % b.N
}

// Adopt(): non-standard.
//
// For efficiency's sake, (possibly) take ownership of
// already allocated slice offered in me.
//
// If me is large we will adopt it, and we will potentially then
// write to the me buffer.
// If we already have a bigger buffer, copy me into the existing
// buffer instead.
func (b *FrameRingBuf) Adopt(me []*Frame) {
	n := len(me)
	if n > b.N {
		b.A = me
		b.N = n
		b.Beg = 0
		b.Readable = n
	} else {
		// we already have a larger buffer, reuse it.
		copy(b.A, me)
		b.Beg = 0
		b.Readable = n
	}
}

func intMin(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// Avail returns the number of available readable
// Frames stored in the ring.
func (f *FrameRingBuf) Avail() int {
	return f.Readable
}

// Kth presents the contents of the
// ring as a strictly linear sequence,
// so the user doesn't need to think
// about modular arithmetic. Here k indexes from
// [0, f.Readable-1], assuming f.Avail()
// is greater than 0. Kth() returns the
// k-th frame starting from f.Beg, where
// f.Beg itself is at k = 0. If k is
// out of bounds, or the ring is empty,
// nil is returned.
func (f *FrameRingBuf) Kth(k int) *Frame {
	if f.Readable == 0 || k < 0 || k >= f.Readable {
		return nil
	}
	pos := (f.Beg + k) % f.N
	return f.A[pos]
}

// First returns the earliest index, or -1 if
// the ring is empty
func (f *FrameRingBuf) First() int {
	if f.Readable == 0 {
		return -1
	}
	return f.Beg
}

// Nextpos returns the index of the element after
// from, or -1 if no more. returns -2 if erroneous
// input (bad from).
func (f *FrameRingBuf) Nextpos(from int) int {
	if from >= f.N || from < 0 {
		return -2
	}
	if f.Readable == 0 {
		return -1
	}

	last := f.Last()
	if from == last {
		return -1
	}
	a0, a1, b0, b1 := f.LegalPos()
	switch {
	case from >= a0 && from < a1:
		return from + 1
	case from == a1:
		return b0 // can be -1
	case from >= b0 && from < b1:
		return from + 1
	case from == b1:
		return -1
	}
	return -1
}

// LegalPos returns the legal index positions,
// [a0,aLast] and [b0,bLast] inclusive, where the
// [a0,aLast] holds the first FIFO ordered segment,
// and the [b0,bLast] holds the second ordered segment,
// if any.
// A position of -1 means the segment is not used,
// perhaps because b.Readable is zero, or because
// the second segment [b0,bLast] is not in use (when
// everything fits in the first [a0,aLast] segment).
//
func (b *FrameRingBuf) LegalPos() (a0, aLast, b0, bLast int) {
	a0 = -1
	aLast = -1
	b0 = -1
	bLast = -1
	if b.Readable == 0 {
		return
	}
	a0 = b.Beg
	last := b.Beg + b.Readable - 1
	if last < b.N {
		aLast = last
		return
	}
	aLast = b.N - 1
	b0 = 0
	bLast = last % b.N
	return
}

// Prevpos returns the index of the element before
// from, or -1 if no more and from is the
// first in the ring. Returns -2 on bad
// from position.
func (f *FrameRingBuf) Prevpos(from int) int {
	if from >= f.N || from < 0 {
		return -2
	}
	if f.Readable == 0 {
		return -1
	}
	if from == f.Beg {
		return -1
	}
	a0, a1, b0, b1 := f.LegalPos()
	switch {
	case from == a0:
		return -1
	case from > a0 && from <= a1:
		return from - 1
	case from == b0:
		return a1
	case from > b0 && from <= b1:
		return from - 1
	}
	return -1
}

// Last returns the index of the last element,
// or -1 if the ring is empty.
func (f *FrameRingBuf) Last() int {
	if f.Readable == 0 {
		return -1
	}

	last := f.Beg + f.Readable - 1
	if last < f.N {
		// we fit without wrapping
		return last
	}

	return last % f.N
}

// WriteCapacity returns the number of spaces
// left to write in the ring before it is full.
// When the ring is full, 0 is returned.
func (b *FrameRingBuf) WriteCapacity() int {
	return b.N - b.Readable
}
