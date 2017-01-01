package tm

import (
	"flag"
	"fmt"

	"github.com/glycerine/zebrapack/zebra"
)

// configure the tfcat command utility
type TfcatConfig struct {
	PrettyPrint         bool
	SkipPayload         bool
	Follow              bool
	RawCount            int
	RawSkip             int
	ReadStdin           bool
	Rreadable           bool
	ZebraPackSchemaPath string

	ZebraSchema zebra.Schema
}

// call DefineFlags before myflags.Parse()
func (c *TfcatConfig) DefineFlags(fs *flag.FlagSet) {
	fs.IntVar(&c.RawCount, "raw", 0, "count of raw messages to pass to stdout (-p, -s, -f are ignored if -raw is given). -raw is useful for extracting a few messages from the front of a file.")
	fs.IntVar(&c.RawSkip, "rawskip", 0, "count of raw messages to skip before passing the rest to stdout (-p, -s, and -f are ignored if -rawskip is given). -rawskip is useful for extracting messages from the middle of a file. -raw count of messages (after the skip count) are written to stdout if both -raw and -rawskip are given.")
	fs.BoolVar(&c.PrettyPrint, "p", false, "pretty print output.")
	fs.BoolVar(&c.SkipPayload, "s", false, "short display. skip printing any data payload.")
	fs.BoolVar(&c.Follow, "f", false, "follow the file, only printing any new additions.")
	fs.BoolVar(&c.ReadStdin, "stdin", false, "read input from stdin rather than a file. tfcat cannot also -f follow stdin.")
	fs.BoolVar(&c.Rreadable, "r", false, "display in R consumable format")
	fs.StringVar(&c.ZebraPackSchemaPath, "zebrapack-schema", "", "path to ZebraPack schema in msgpack2 format to read for decoding messages")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *TfcatConfig) ValidateConfig() error {
	if c.ZebraPackSchemaPath != "" &&
		!FileExists(c.ZebraPackSchemaPath) {
		return fmt.Errorf("bad -zebrapack-schema path: "+
			"'%s' does not exist.", c.ZebraPackSchemaPath)
	}
	return nil
}

////////////////////////////
// tfindex

// configure the tfindex command utility
type TfindexConfig struct {
}

// call DefineFlags before myflags.Parse()
func (c *TfindexConfig) DefineFlags(fs *flag.FlagSet) {
}

// call c.ValidateConfig() after myflags.Parse()
func (c *TfindexConfig) ValidateConfig() error {
	return nil
}

////////////////////////////
// tfsort

// configure the tfsort command utility
type TfsortConfig struct {
	KeepTmpFiles bool
}

// call DefineFlags before myflags.Parse()
func (c *TfsortConfig) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.KeepTmpFiles, "k", false, "keep .sorted intermediate temp files")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *TfsortConfig) ValidateConfig() error {
	return nil
}

////////////////////////////
// tfdedup

// configure the tfdedup command utility
type TfdedupConfig struct {
	WriteDupsToFile string
	WindowSize      int
	DetectOnly      bool
}

// call DefineFlags before myflags.Parse()
func (c *TfdedupConfig) DefineFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.WriteDupsToFile, "dupsto", "", "write duplicates to this file path")
	fs.IntVar(&c.WindowSize, "window", 1000, "window size; number of Frames in a row to check for duplicates")
	fs.BoolVar(&c.DetectOnly, "detect", false, "detect duplicates and announce that "+
		"fact, but do not write any Frame output")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *TfdedupConfig) ValidateConfig() error {
	if c.WriteDupsToFile != "" && FileExists(c.WriteDupsToFile) {
		return fmt.Errorf("duplicates output file '%s' already exists, aborting.", c.WriteDupsToFile)
	}
	if c.WindowSize <= 1 {
		return fmt.Errorf("-window %v illegal: must be positive integer > 1.", c.WindowSize)
	}
	return nil
}

////////////////
// tfsum

type TfsumConfig struct {
	Help bool
}

// call DefineFlags before myflags.Parse()
func (c *TfsumConfig) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.Help, "h", false, "show this help")
}

func (c *TfsumConfig) ValidateConfig() error {
	return nil
}

////////////////
// tfgroup

type TfgroupConfig struct {
	Help          bool
	GroupInterval string // "min, sec, hour, ..."
}

// call DefineFlags before myflags.Parse()
func (c *TfgroupConfig) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.Help, "h", false, "show this help")
}

func (c *TfgroupConfig) ValidateConfig() error {
	return nil
}

////////////////
// tffilter

type TffilterConfig struct {
	Help           bool
	ExcludeMatches bool
	RegexFile      string
	Any            bool
	Sub            bool
}

// call DefineFlags before myflags.Parse()
func (c *TffilterConfig) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.Help, "h", false, "show this help")
	fs.BoolVar(&c.ExcludeMatches, "x", false, "exclude regex matches rather than, by default, outputing only matches.")
	fs.StringVar(&c.RegexFile, "regexfile", "", "read a newline separated list of regex from this file")
	fs.BoolVar(&c.Any, "any", false, "include the frame if any of the regex matches (effectively OR-ing the regex instead of the default AND-ing)")
	fs.BoolVar(&c.Sub, "sub", false, "print only sub-expression matches of the regular expression")
}

func (c *TffilterConfig) ValidateConfig() error {
	if c.RegexFile != "" && !FileExists(c.RegexFile) {
		return fmt.Errorf("-regexfile '%s' does not exist", c.RegexFile)
	}
	return nil
}
