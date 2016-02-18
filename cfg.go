package tm

import (
	"flag"
)

// configure the tfcat command utility
type TfcatConfig struct {
	PrettyPrint bool
	SkipPayload bool
	Follow      bool
	RawCount    int
}

// call DefineFlags before myflags.Parse()
func (c *TfcatConfig) DefineFlags(fs *flag.FlagSet) {
	fs.IntVar(&c.RawCount, "raw", 0, "count of raw messages to pass to stdout (-p, -s, -f are ignored if -raw is given). -raw is useful for extracting a few messages from the front of a file.")
	fs.BoolVar(&c.PrettyPrint, "p", false, "pretty print output.")
	fs.BoolVar(&c.SkipPayload, "s", false, "short display. skip printing any data payload.")
	fs.BoolVar(&c.Follow, "f", false, "follow the file, only printing any new additions.")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *TfcatConfig) ValidateConfig() error {
	return nil
}
