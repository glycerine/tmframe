package tm

import (
	"fmt"
)

// p is a shortcut for a call to fmt.Printf that implicitly starts
// and ends its message with a newline.
func p(format string, stuff ...interface{}) {
	fmt.Printf("\n "+format+"\n", stuff...)
}

// q calls are quietly ignored. They allow conversion from p()
// calls to be swapped quickly and easily.
func q(quietly_ignored ...interface{}) {} // quiet
