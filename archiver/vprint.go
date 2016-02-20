package archiver

import (
	"fmt"
)

func p(format string, stuff ...interface{}) {
	fmt.Printf("\n "+format+"\n", stuff...)
}

func q(quietly_ignored ...interface{}) {} // quiet
