package archiver

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"time"
)

func demoFlocks() {
	f, err := os.Create("demo.flocks.file")
	panicOn(err)
	fmt.Fprintf(f, "hello")
	f.Close()

	g, err := os.Open("demo.flocks.file")

	res := LockFile(g)
	p("res = '%v'\n", res)
	have := HaveLock(g)
	p("have lock = %v\n", have)
	res2 := UnlockFile(g)
	p("Unlock result2 = %v\n", res2)

	have2 := HaveLock(g)
	p("have2 lock = %v\n", have2)

	time.Sleep(time.Hour)
}

// NoLockErr is returned by LockFile when it
// fails to lock the file.
var NoLockErr = fmt.Errorf("no flock lock was obtained")

// LockFile returns nil if the lock was obtained.
// If the lock was not obtained, NoLockErr is
// returned
func LockFile(f *os.File) error {
	var how int = unix.LOCK_EX | unix.LOCK_NB
	err := unix.Flock(int(f.Fd()), how)
	if err == unix.EWOULDBLOCK {
		return NoLockErr
	} else {
		return nil
	}
}

// HaveLock checks if we have an exclusive flock lock on the file.
// flocks(2) can be inherited by child processes.
func HaveLock(f *os.File) bool {
	var how int = unix.LOCK_EX | unix.LOCK_NB
	err := unix.Flock(int(f.Fd()), how)
	if err == nil {
		return true
	}
	return false
}

// Unlock calls flock with LOCK_UN, and returns any error.
// Typically the returned error is nil whether we succeeded
// or not in unlocking the file. Note that the operating
// system automatically unlocks flocks when the process
// exits.
func UnlockFile(f *os.File) error {
	var how int = unix.LOCK_UN
	return unix.Flock(int(f.Fd()), how)
}
