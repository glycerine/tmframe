package main

import (
	"flag"
	"fmt"
	tf "github.com/glycerine/tmframe"
	fsnotify "gopkg.in/fsnotify.v1"
	"io"
	"io/ioutil"
	"os"
)

func showUse(myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s displays TMFRAME files. Usage: %s {-p} {-s} {-f} <file1> <file2> ...\n", os.Args[0], os.Args[0])
	myflags.PrintDefaults()
}

func usage(err error, myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	showUse(myflags)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	myflags := flag.NewFlagSet("tfcat", flag.ExitOnError)
	cfg := &tf.TfcatConfig{}
	cfg.DefineFlags(myflags)

	err := myflags.Parse(os.Args[1:])
	err = cfg.ValidateConfig()
	if err != nil {
		usage(err, myflags)
	}

	if cfg.ZebraPackSchemaPath != "" {
		by, err := ioutil.ReadFile(cfg.ZebraPackSchemaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tfcat error reading -zebrapack-schema file '%s': %v\n", cfg.ZebraPackSchemaPath, err)
			os.Exit(1)
		}
		_, err = cfg.ZebraSchema.UnmarshalMsg(by)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tfcat error Unmarshalling the -zebrapack-schema file '%s': %v\n", cfg.ZebraPackSchemaPath, err)
			os.Exit(1)
		}
	}

	leftover := myflags.Args()
	//Q("leftover = %v", leftover)
	if len(leftover) == 0 && cfg.ReadStdin == false {
		fmt.Fprintf(os.Stderr, "no input files given\n")
		showUse(myflags)
		os.Exit(1)
	}
	GlobalPrettyPrint = cfg.PrettyPrint

	if cfg.Follow {
		if len(leftover) != 1 {
			if cfg.ReadStdin {
				fmt.Fprintf(os.Stderr, "cannot follow stdin\n")
				showUse(myflags)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "can only follow a single file\n")
			showUse(myflags)
			os.Exit(1)
		}
		FollowFile(leftover[0], cfg)
		return
	}

	if cfg.ReadStdin {
		if len(leftover) > 0 {
			fmt.Fprintf(os.Stderr, "if reading from stdin, cannot also read from files\n")
			showUse(myflags)
			os.Exit(1)
		}
		leftover = []string{"stdin"}
	}

	if cfg.RawCount > 0 || cfg.RawSkip > 0 {
		if len(leftover) != 1 {
			fmt.Fprintf(os.Stderr, "can only copy raw messages from one file\n")
			showUse(myflags)
			os.Exit(1)
		}
		SendRawBytes(leftover[0], cfg.RawCount, os.Stdout, cfg.RawSkip)
		return
	}

	i := int64(1)
nextfile:
	for _, inputFile := range leftover {

		f := prepInput(inputFile)
		defer f.Close()
		//P("starting on inputFile '%s'", inputFile)

		fr := tf.NewFrameReader(f, 1024*1024)

		var frame tf.Frame

		for ; err == nil; i++ {
			_, _, err, _ = fr.NextFrame(&frame)
			if err != nil {
				if err == io.EOF {
					continue nextfile
				}
				fmt.Fprintf(os.Stderr, "tfcat error from fr.NextFrame() at i=%v: '%v'\n", i, err)
				os.Exit(1)
			}
			frame.DisplayFrame(os.Stdout, i, cfg.PrettyPrint, cfg.SkipPayload, cfg.Rreadable, &cfg.ZebraSchema)
		}
	}
}

func FollowFile(path string, cfg *tf.TfcatConfig) {

	if !FileExists(path) {
		fmt.Fprintf(os.Stderr, "input file '%s' does not exist.\n", path)
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	panicOn(err)
	defer watcher.Close()

	f, err := os.Open(path)
	panicOn(err)
	// move to end for tailing
	_, err = f.Seek(0, 2)
	panicOn(err)

	err = watcher.Add(path)
	panicOn(err)

	fr := tf.NewFrameReader(f, 1024*1024)

	var frame tf.Frame

	i := int64(1)
nextFrame:
	for {
		_, _, err, _ = fr.NextFrame(&frame)
		if err != nil {
			if err == io.EOF {
				select {
				case event := <-watcher.Events:
					if event.Op&fsnotify.Write == fsnotify.Write {
						continue nextFrame
					}
				}
			}
			fmt.Fprintf(os.Stderr, "tfcat error from fr.NextFrame(): '%v'\n", err)
			os.Exit(1)
		}
		frame.DisplayFrame(os.Stdout, i, cfg.PrettyPrint, cfg.SkipPayload, cfg.Rreadable, &cfg.ZebraSchema)
		i++
	}
}

func prepInput(inputPath string) *os.File {

	if inputPath != "stdin" && !FileExists(inputPath) {
		fmt.Fprintf(os.Stderr, "input file '%s' does not exist.\n", inputPath)
		os.Exit(1)
	}

	var f *os.File
	var err error
	if inputPath == "stdin" {
		f = os.Stdin
	} else {
		f, err = os.Open(inputPath)
		panicOn(err)
	}

	return f
}

// copy the raw TMFRAME bytes of messageCount messages read from
// inputPath to w
func SendRawBytes(inputPath string, writeFrameCount int, w io.Writer, skipFrameCount int) {

	f := prepInput(inputPath)
	defer f.Close()
	var err error

	fr := tf.NewFrameReader(f, 1024*1024)
	var frame tf.Frame
	skipByteCount := int64(0)
	writeByteCount := int64(0)

	for i := 0; i < skipFrameCount; i++ {
		_, nbytes, err, _ := fr.NextFrame(&frame)
		if err != nil {
			panic(err)
			//os.Exit(0)
		}
		skipByteCount += nbytes
	}

	for i := 0; i < writeFrameCount; i++ {
		_, nbytes, err, _ := fr.NextFrame(&frame)
		if err != nil {
			panic(err)
			//os.Exit(0)
		}
		writeByteCount += nbytes
	}

	//fmt.Fprintf(os.Stderr, "skipByteCount: %v, writeByteCount: %v, skipFramCount: %v, writeFrameCount: %v\n", skipByteCount, writeByteCount, skipFrameCount, writeFrameCount)

	// seek back to beginning to copy just those bytes
	_, err = f.Seek(skipByteCount, 0)
	panicOn(err)
	if writeFrameCount == 0 {
		// write them all
		_, err = io.Copy(w, f)
		panicOn(err)
	} else {
		var wrote int64
		wrote, err = io.CopyN(w, f, writeByteCount)
		if wrote != writeByteCount {
			panic(fmt.Sprintf("short write: %v vs %v expected", wrote, writeByteCount))
		}
		panicOn(err)
	}
}
