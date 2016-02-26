package main

import (
	"flag"
	"fmt"
	tf "github.com/glycerine/tmframe"
	fsnotify "gopkg.in/fsnotify.v1"
	"io"
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

	leftover := myflags.Args()
	//Q("leftover = %v", leftover)
	if len(leftover) == 0 {
		fmt.Fprintf(os.Stderr, "no input files given\n")
		showUse(myflags)
		os.Exit(1)
	}
	GlobalPrettyPrint = cfg.PrettyPrint

	if cfg.Follow {
		if len(leftover) != 1 {
			fmt.Fprintf(os.Stderr, "can only follow a single file\n")
			showUse(myflags)
			os.Exit(1)
		}
		FollowFile(leftover[0], cfg)
		return
	}

	if cfg.RawCount > 0 {
		if len(leftover) != 1 {
			fmt.Fprintf(os.Stderr, "can only copy raw messages from one file\n")
			showUse(myflags)
			os.Exit(1)
		}
		SendRawBytes(leftover[0], cfg.RawCount, os.Stdout)
		return
	}

	i := int64(1)
nextfile:
	for _, inputFile := range leftover {
		//P("starting on inputFile '%s'", inputFile)
		if !FileExists(inputFile) {
			fmt.Fprintf(os.Stderr, "input file '%s' does not exist.\n", inputFile)
			os.Exit(1)
		}

		f, err := os.Open(inputFile)
		panicOn(err)
		fr := tf.NewFrameReader(f, 1024*1024)

		var frame tf.Frame

		for ; err == nil; i++ {
			_, _, err = fr.NextFrame(&frame)
			if err != nil {
				if err == io.EOF {
					continue nextfile
				}
				fmt.Fprintf(os.Stderr, "tfcat error from fr.NextFrame() at i=%v: '%v'\n", i, err)
				os.Exit(1)
			}
			frame.DisplayFrame(os.Stdout, i, cfg.PrettyPrint, cfg.SkipPayload)
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
		_, _, err = fr.NextFrame(&frame)
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
		frame.DisplayFrame(os.Stdout, i, cfg.PrettyPrint, cfg.SkipPayload)
		i++
	}
}

// copy the raw TMFRAME bytes of messageCount messages read from
// inputPath to w
func SendRawBytes(inputPath string, messageCount int, w io.Writer) {

	if !FileExists(inputPath) {
		fmt.Fprintf(os.Stderr, "input file '%s' does not exist.\n", inputPath)
		os.Exit(1)
	}

	f, err := os.Open(inputPath)
	panicOn(err)
	defer f.Close()

	fr := tf.NewFrameReader(f, 1024*1024)
	var frame tf.Frame
	byteCount := int64(0)

	for i := 0; i < messageCount; i++ {
		_, nbytes, err := fr.NextFrame(&frame)
		if err != nil {
			panic(err)
			//os.Exit(0)
		}
		byteCount += nbytes
	}

	// seek back to beginning to copy just those bytes
	_, err = f.Seek(0, 0)
	panicOn(err)
	wrote, err := io.CopyN(w, f, byteCount)
	if wrote != byteCount {
		panic(fmt.Sprintf("short write: %v vs %v expected", wrote, byteCount))
	}
	panicOn(err)

}
