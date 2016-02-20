package main

import (
	"flag"
	"fmt"
	"github.com/glycerine/tmframe/archiver"
	"os"
)

func usage(err error, myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s\n", err)

	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	myflags.PrintDefaults()
	os.Exit(1)
}

func main() {
	myflags := flag.NewFlagSet("archver", flag.ExitOnError)
	cfg := &archiver.ArchiverConfig{}
	cfg.DefineFlags(myflags)

	err := myflags.Parse(os.Args[1:])
	err = cfg.ValidateConfig()
	if err != nil {
		usage(err, myflags)
	}

	archiv := archiver.NewFileMgr(cfg)
	err = archiv.Run()
	if err != nil {
		panic(err)
	}
}
