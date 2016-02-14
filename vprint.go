package tm

import (
	"fmt"
	"time"
)

// set Verbose to true to debug
var Verbose bool

// set Working to true to show prints currently under investigation
var Working bool

// A debug tool: V acts as a Printf that only prints when Verbose is true.
var V = VPrintf

// A debug tool: W acts as a Printf that only prints when Working is true.
var W = WPrintf

// P is a shortcut for a call to fmt.Printf that implicitly starts
// and ends its message with a newline.
func P(format string, stuff ...interface{}) {
	fmt.Printf("\n "+format+"\n", stuff...)
}

// Q calls are quietly ignored. They allow conversion from P()
// calls to be swapped quickly and easily.
func Q(quietly_ignored ...interface{}) {} // quiet

// get timestamp for logging purposes
func ts() string {
	return time.Now().Format("2006-01-02 15:04:05.999 -0700 MST")
}

// TSPrintf provides a time-stamped printf for debugging timing issues.
func TSPrintf(format string, a ...interface{}) {
	fmt.Printf("%s ", ts())
	fmt.Printf(format, a...)
}

// VPrintf prints via TSPrintf when Verbose is true.
func VPrintf(format string, a ...interface{}) {
	if Verbose {
		TSPrintf(format, a...)
	}
}

// WPrintf prints via TSPrintf when Working is true.
func WPrintf(format string, a ...interface{}) {
	if Working {
		TSPrintf(format, a...)
	}
}
