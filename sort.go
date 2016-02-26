package tm

// example: sort.Stable(TimeSorter(frames)) will
// do a stable sort on the timestamp Prim field.

// frameSorter is used to do the merge sort
type TimeSorter []*Frame

// Len is the sorting Len function
func (p TimeSorter) Len() int { return len(p) }

// Less is the sorting Less function.
func (p TimeSorter) Less(i, j int) bool {
	return p[i].Tm() < p[j].Tm()
}

// Swap is the sorting Swap function.
func (p TimeSorter) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
