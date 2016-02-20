package archiver

import (
	"flag"
	"fmt"
)

type ArchiverConfig struct {
	WriteDir   string
	ServerList string
	TlsDir     string
}

// call DefineFlags before myflags.Parse()
func (c *ArchiverConfig) DefineFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.WriteDir, "write-dir", "", "write messages to this directory")
	fs.StringVar(&c.ServerList, "gnatsd-servers", "nats://user:password@127.0.0.1:4444", "comma separated URLs to find gnatsd daemons(s)")
	fs.StringVar(&c.TlsDir, "tlsdir", "/etc/ssl", "which dir to load certificates, key, and CA from.")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *ArchiverConfig) ValidateConfig() error {
	if c.WriteDir == "" {
		return fmt.Errorf("missing and required: -write-dir not given")
	}
	if !DirExists(c.TlsDir) {
		return fmt.Errorf("-tlsdir refers to non existant directory '%s'", c.TlsDir)
	}
	return nil
}
