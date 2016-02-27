package archiver

import (
	"encoding/binary"
	"fmt"
	"github.com/nats-io/nats"
	"io"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"
)

// the "/" path seperator on osx/linux.
var sep = string(os.PathSeparator)

// ServiceName is how this service is addressed on the Nats bus.
var ServiceName = "service-name"

// in UTC time/day boundaries, give the Date
type Date struct {
	Year  int
	Month int
	Day   int
}

func TimeToDate(tm time.Time) Date {
	utc := tm.UTC()
	return Date{
		Year:  utc.Year(),
		Month: int(utc.Month()),
		Day:   utc.Day(),
	}
}

// File is an open file in our cache of file handles
type File struct {
	Key        string
	Date       Date
	StreamName string
	Path       string
	Fd         *os.File
	EndOffset  int64
	LastErr    error
	LastWrite  time.Time
}

type FileMgr struct {
	Cfg   ArchiverConfig
	Files map[string]*File

	Ready   chan bool
	ReqStop chan bool
	Done    chan bool

	NatsAsyncErrCh   chan asyncErr
	NatsConnClosedCh chan *nats.Conn
	NatsConnDisconCh chan *nats.Conn
	NatsConnReconCh  chan *nats.Conn
	ArchiveReqCh     chan *nats.Msg

	SignalInterruptCh chan os.Signal

	subscription *nats.Subscription
	nc           *nats.Conn

	// nats async handlers and options
	opts []nats.Option

	certs certConfig
}

func NewFileMgr(cfg *ArchiverConfig) *FileMgr {
	fm := &FileMgr{
		Cfg:               *cfg,
		Files:             make(map[string]*File),
		Ready:             make(chan bool),
		ReqStop:           make(chan bool),
		Done:              make(chan bool),
		NatsConnClosedCh:  make(chan *nats.Conn),
		NatsConnDisconCh:  make(chan *nats.Conn),
		NatsConnReconCh:   make(chan *nats.Conn),
		ArchiveReqCh:      make(chan *nats.Msg),
		SignalInterruptCh: make(chan os.Signal, 5),
		NatsAsyncErrCh:    make(chan asyncErr),
	}
	fm.certs.init(cfg.TlsDir)
	signal.Notify(fm.SignalInterruptCh, os.Interrupt)
	return fm
}

// provides index into File map
func (fm *FileMgr) GetDateNameString(tm time.Time, streamName string) (string, Date) {
	date := TimeToDate(tm)
	path := path.Join(fmt.Sprintf("%04d", date.Year), fmt.Sprintf("%02d", date.Month), fmt.Sprintf("%02d", date.Day), streamName)
	return path, date
}

func (fm *FileMgr) GetPath(tm time.Time, streamName string) string {
	dateNameStr, _ := fm.GetDateNameString(tm, streamName)
	path := fm.Cfg.WriteDir + string(os.PathSeparator) + dateNameStr
	return path
}

// XtraDirs allows better scalabilty for directory file counts by adding a
// 3 layer deep directory hierarchy.
func XtraDirs(id string) string {
	a := string(id[0])
	b := string(id[1])
	c := string(id[2])
	return fmt.Sprintf("%s%s%s%s%s%s%s%s", a, sep, b, sep, c, sep, id, sep)
}

func (fm *FileMgr) Store(tm time.Time, streamName string, data []byte) (f *File, err error) {
	q("starting Store() for streamName '%s' with data len = %v", streamName, len(data))
	if len(data) == 0 {
		p("data length is zero, returning nil from Store()")
		return nil, nil
	}

	dateNameStr, date := fm.GetDateNameString(tm, streamName)
	var fd *os.File
	var ok bool
	var pth string
	f, ok = fm.Files[dateNameStr]
	if ok {
		q("got hit for dateNameStr == '%s'; f.Fd=%#v", dateNameStr, f.Fd)
		fd = f.Fd
		pth = f.Path
		f.LastWrite = time.Now()
	} else {
		pth = fm.GetPath(tm, streamName)
		if FileExists(pth) {
			// must re-use the existing fd here
			fd, err = os.OpenFile(pth, os.O_CREATE|os.O_RDWR, 0664)
			if err != nil {
				return nil, fmt.Errorf("FileMgr.Store() error opening path '%s': '%s'", pth, err)
			}
			if NoLockErr == LockFile(fd) {
				fd.Close()
				return nil, fmt.Errorf("FileMgr.Store() error opening path '%s': "+
					"already flocked by another process.", pth)
			}
			// move to end to appending
			_, err = fd.Seek(0, 2)
			if err != nil {
				return nil, fmt.Errorf("FileMgr.Store() error seeking to end of path '%s': '%s'", pth, err)
			}
		} else {
			err = os.MkdirAll(path.Dir(pth), 0775)
			if err != nil {
				return nil, fmt.Errorf("FileMgr.Store() error during MkdirAll on path '%s': '%s'", pth, err)
			}
			fd, err = os.Create(pth)
			if err != nil {
				return nil, fmt.Errorf("FileMgr.Store() error creating path '%s': '%s'", pth, err)
			}
			if NoLockErr == LockFile(fd) {
				fd.Close()
				return nil, fmt.Errorf("FileMgr.Store() error opening path '%s': "+
					"already flocked by another process -- odd as we just Create()-ed this file.", pth)
			}

		}
		// INVAR: fd good to go.
		f = &File{
			Key:        dateNameStr,
			Date:       date,
			StreamName: streamName,
			Path:       pth,
			Fd:         fd,
			LastWrite:  time.Now(),
		}
		fm.Files[dateNameStr] = f
	}
	// INVAR f and fd are good to go

	attemptsShort := 0
	total := len(data)
	wrote := 0
	var m int
writeloop:
	for {
		m, err = fd.Write(data)
		f.EndOffset += int64(m)
		wrote += m
		data = data[m:]
		// guaranteed full length write if err is nil.
		if err == nil || wrote >= total {
			return f, nil
		}
		if err == io.ErrShortWrite {
			attemptsShort++
			if attemptsShort > 100 {
				return nil, fmt.Errorf("FileMgr.Store(): error writing data to "+
					"'%v' to path '%s': (over 100 short writes)'%s'", fd, pth, err)
			}
			continue writeloop
		}
		if err != nil {
			f.LastErr = err
			return nil, fmt.Errorf("FileMgr.Store(): error writing data to '%v'"+
				" to path '%s': '%s'", fd, pth, err)
		}
	}

	return f, nil
}

// sync all open files
func (fm *FileMgr) Sync() {
	for _, v := range fm.Files {
		v.Fd.Sync()
	}
}

// close unaccessed files, to keep file handle count low.
func (fm *FileMgr) CloseUnusedFiles(olderThan time.Duration) {
	delme := []string{}
	for key, v := range fm.Files {
		if time.Since(v.LastWrite) > olderThan {
			v.Fd.Sync()
			v.Fd.Close()
			UnlockFile(v.Fd)
			delme = append(delme, key)
		}
	}
	for _, k := range delme {
		delete(fm.Files, k)
	}
}

func (fm *FileMgr) SyncAndCloseAllFiles() {
	for _, v := range fm.Files {
		v.Fd.Sync()
		v.Fd.Close()
		UnlockFile(v.Fd)
	}
	// clear the map
	fm.Files = make(map[string]*File)
}

type asyncErr struct {
	conn *nats.Conn
	sub  *nats.Subscription
	err  error
}

func (fm *FileMgr) setupNatsOptions() {

	if !fm.certs.skipTLS {
		err := fm.certs.certLoad()
		if err != nil {
			panic(err)
		}
	}

	o := []nats.Option{}
	o = append(o, nats.MaxReconnects(-1)) // -1 => keep trying forever
	o = append(o, nats.ReconnectWait(2*time.Second))
	o = append(o, nats.Name("archiver"))

	o = append(o, nats.ErrorHandler(func(c *nats.Conn, s *nats.Subscription, e error) {
		fm.NatsAsyncErrCh <- asyncErr{conn: c, sub: s, err: e}
	}))
	o = append(o, nats.DisconnectHandler(func(conn *nats.Conn) {
		fm.NatsConnDisconCh <- conn
	}))
	o = append(o, nats.ReconnectHandler(func(conn *nats.Conn) {
		fm.NatsConnReconCh <- conn
	}))
	o = append(o, nats.ClosedHandler(func(conn *nats.Conn) {
		fm.NatsConnClosedCh <- conn
	}))

	if !fm.certs.skipTLS {
		o = append(o, nats.Secure(&fm.certs.tlsConfig))
		o = append(o, fm.certs.rootCA)
	}

	fm.opts = o
}

// blocks until done
func (fm *FileMgr) Run() error {
	p("FileMgr Run() started")
	var err error

	fm.setupNatsOptions()
	subj := ServiceName + ".archive.>"

	minuteTimer := time.After(time.Minute)
	hourTimer := time.After(time.Hour)

	i := 0

	// poll and dispatch archive requests
	for {
		// set fm.nc to nil if we need to connect again.
		if fm.nc == nil {
			p("about to connect with FileMgr client using server list: '%s'", fm.Cfg.ServerList)
			nc, err := nats.Connect(fm.Cfg.ServerList, fm.opts...)
			if err != nil {
				panic(err)
			}
			p("FileMgr nats.Connect(fm.Cfg.ServerList='%s') done.", fm.Cfg.ServerList)

			fm.nc = nc
			err = fm.Subscribe(subj)
			panicOn(err)
			p("Listening on [%s] after Flush() has returned", subj)
			close(fm.Ready)
			p("closed fm.Ready")
		}

		select {
		case <-minuteTimer:
			// sync files to disk every minute.
			fm.Sync()
			minuteTimer = time.After(time.Minute)

		case <-hourTimer:
			// Guarantee that files unaccessed for more than two hours are closed.
			// Keeps file handle count from growing too large.
			fm.CloseUnusedFiles(time.Hour * 2)
			hourTimer = time.After(time.Hour)

		case <-fm.SignalInterruptCh:
			return fm.shutdown()

		case <-fm.ReqStop:
			return fm.shutdown()

		case asErr := <-fm.NatsAsyncErrCh:
			p("*** async error handler sees error: '%s'", asErr.err)

		case <-fm.NatsConnClosedCh:
			p("%v closed handler called", time.Now())
			// indicate manual reconnect is required.
			// with MaxReconnects of -1 we should never get here.
			fm.nc = nil
		case <-fm.NatsConnDisconCh:
			p("%v disco handler called", time.Now())
			// let the client take care of reconnecting
		case <-fm.NatsConnReconCh:
			p("%v recon handler called", time.Now())
		case msg := <-fm.ArchiveReqCh:
			q("filemgr listener sees msg with subject:'%v' and reply-to: '%s'",
				msg.Subject, msg.Reply)
			i += 1

			if i%1000 == 0 {
				p("%s.archive.* sees its %v msg since this archiver started; with subject '%s'. at %v",
					ServiceName, i, msg.Subject, time.Now())
			}
			streamName, id := ExtractStreamFromSubject(msg.Subject)
			if streamName == "" {
				panic(fmt.Errorf("no stream in subject '%s'!", msg.Subject))
			}

			// grab the timestamp off the TMFRAME, if we have at least 8 bytes.
			var tm time.Time
			if len(msg.Data) >= 8 {
				tm = time.Unix(0, int64(binary.LittleEndian.Uint64(msg.Data[:8])))
			} else {
				tm = time.Now()
			}

			_, err = fm.Store(tm, streamName, msg.Data)
			panicOn(err)
			q("id=%v", id)
			// ack

			if len(msg.Reply) > 0 {
				err = fm.nc.Publish(msg.Reply, []byte(id))
				p("FileMgr published to msg.Reply:'%s' with id '%s'",
					msg.Reply, id)
				panicOn(err)
				err = fm.nc.Publish(ServiceName+".storage-ack."+streamName, []byte(id))
				p("FileMgr published to ack.%s with id '%s'",
					streamName, id)
				panicOn(err)
			}
		}

	}
	return nil
}

func (fm *FileMgr) shutdown() error {
	err := fm.subscription.Unsubscribe()

	// avoid deadlock on the nats close callback
	go func() { <-fm.NatsConnClosedCh }()

	fm.nc.Close()
	fm.SyncAndCloseAllFiles()
	close(fm.Done)
	return err
}

func (fm *FileMgr) Subscribe(subj string) error {
	var err error
	fm.subscription, err = fm.nc.Subscribe(subj, func(msg *nats.Msg) {
		fm.ArchiveReqCh <- msg
	})
	if err != nil {
		return fmt.Errorf("Got an error on nc.Subscribe(subj:'%s'): %+v\n", subj, err)
	}
	p("subscription on subject '%s' obtained.", subj)

	// nc.Flush() is required or our subscription may not be ready
	// and we will race the Request against it, leading to
	// intermittant test failures.
	return fm.nc.Flush()
}

func (fm *FileMgr) Stop() {
	close(fm.ReqStop)
	<-fm.Done
}

func ExtractStreamFromSubject(subj string) (stream string, id string) {
	splt := strings.Split(subj, ".")
	q("splt = '%#v'", splt)
	m := len(splt)
	switch {
	case m <= 2:
		return "", ""
	case splt[0] != ServiceName || splt[1] != "archive":
		return "", ""
	case m == 3:
		return splt[2], ""
	case m > 3:
		return splt[2], strings.Join(splt[3:], ".")
	}
	return "", ""
}

func GetYearMonthDayString(tm time.Time) (string, Date) {
	date := TimeToDate(tm)
	path := path.Join(fmt.Sprintf("%04d", date.Year), fmt.Sprintf("%02d", date.Month), fmt.Sprintf("%02d", date.Day))
	return path, date
}

func GetYearMonthString(tm time.Time) (string, Date) {
	date := TimeToDate(tm)
	path := path.Join(fmt.Sprintf("%04d", date.Year), fmt.Sprintf("%02d", date.Month))
	return path, date
}
