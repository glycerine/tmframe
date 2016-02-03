package main

import (
	"flag"
)

type TfcatConfig struct {
	PrettyPrint bool
}

// call DefineFlags before myflags.Parse()
func (c *TfcatConfig) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.PrettyPrint, "p", false, "pretty print output")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *TfcatConfig) ValidateConfig() error {
	return nil
}
