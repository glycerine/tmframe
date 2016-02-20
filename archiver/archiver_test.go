package archiver

import (
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	ts "github.com/glycerine/tmframe"
	"github.com/nats-io/gnatsd/server"
	gnatsd "github.com/nats-io/gnatsd/test"
	"github.com/nats-io/nats"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func Test001ArchiverFileMgr(t *testing.T) {

	cv.Convey("given an archiver, messages that roll to different dates should be stored in distinct files", t, func() {
		tmp, err := ioutil.TempDir("", "test-archiver-filemgr")
		panicOn(err)
		defer os.RemoveAll(tmp)
		fm := NewFileMgr(&ArchiverConfig{WriteDir: tmp})

		tm1, err := time.Parse(time.RFC3339, "2016-01-01T00:00:00Z")
		panicOn(err)
		tm2, err := time.Parse(time.RFC3339, "2016-01-02T00:00:00Z")
		panicOn(err)

		streamName := "test"
		data1 := []byte("data1")
		data2 := []byte("data2")

		frame1, err := ts.NewFrame(tm1, ts.EvUtf8, 0, 0, data1)
		panicOn(err)
		by1, err := frame1.Marshal(nil)
		file1, err := fm.Store(tm1, streamName, by1)
		panicOn(err)

		frame2, err := ts.NewFrame(tm2, ts.EvUtf8, 0, 0, data2)
		panicOn(err)
		by2, err := frame2.Marshal(nil)
		file2, err := fm.Store(tm2, streamName, by2)
		panicOn(err)

		p("For inspection, we stored into file1 = '%v'\n", file1.Path)
		q("file2 = '%#v'\n", file2)
		cv.So(strings.Contains(file1.Path, "2016/01/01"), cv.ShouldBeTrue)
		cv.So(strings.Contains(file2.Path, "2016/01/02"), cv.ShouldBeTrue)

		cv.So(FileExists(file1.Path), cv.ShouldBeTrue)
		cv.So(FileExists(file2.Path), cv.ShouldBeTrue)

		by, err := ioutil.ReadFile(file1.Path)
		panicOn(err)

		var fr ts.Frame
		_, err = fr.Unmarshal(by, true)
		p("fr = %#v", fr)
		panicOn(err)

		cv.So(fr.GetEvtnum(), cv.ShouldEqual, ts.EvUtf8)
		cv.So(fr.GetUlen(), cv.ShouldEqual, len(frame1.Data)+1)

		p("Given that we've written an event to its file, we should be able to recover what we've written")
		cv.So(string(fr.Data), cv.ShouldResemble, string(data1))
	})
}

func Test003ArchiverAcknowledgesStorage(t *testing.T) {

	cv.Convey("given a running gnatsd, the archiver should store messages to disk and acknowledge their storage on the 'servicename.storage-ack.(hostname)' subject.", t, func() {

		user := "user"
		pw := "password"
		host := "127.0.0.1"
		port := 4444
		serverList := fmt.Sprintf("nats://%v:%v@%v:%v", user, pw, host, port)

		// start yourself an embedded gnatsd server
		opts := server.Options{
			Host:     host,
			Port:     port,
			Username: user,
			Password: pw,
			Trace:    true,
			Debug:    true,
			//NoLog:  true,
			//NoSigs: true,
		}
		gnats := gnatsd.RunServer(&opts)
		gnats.SetLogger(&Logger{}, true, true)

		//logger := log.New(os.Stderr, "gnatsd: ", log.LUTC|log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
		defer func() {
			p("calling gnats.Shutdown()")
			gnats.Shutdown() // when done
		}()
		addr := fmt.Sprintf("%v:%v", host, port)
		if !PortIsBound(addr) {
			panic("port not bound " + addr)
		}

		// start client
		asyncHandler := nats.ErrorHandler(func(c *nats.Conn, s *nats.Subscription, e error) {
			fmt.Printf("\n *** async error handler sees error: '%s'\n", e)
			panic(e)
		})
		p("about to connect with client")
		nc, err := nats.Connect(serverList, asyncHandler)
		panicOn(err)
		defer nc.Close()

		// prep data
		p("prepping data")
		streamName := "test"
		data1 := []byte("data1")

		tm1, err := time.Parse(time.RFC3339, "2016-01-01T00:00:00Z")
		panicOn(err)

		frame1, err := ts.NewFrame(tm1, ts.EvUtf8, 0, 0, data1)
		panicOn(err)
		framed, err := frame1.Marshal(nil)
		panicOn(err)

		// subscribe to reply in advance of publishing archive request
		gotAck := make(chan *nats.Msg)
		nc.Subscribe(ServiceName+".storage-ack.>", func(msgNats *nats.Msg) {
			p(ServiceName+".storage-ack received msg: '%#v'", msgNats)
			select {
			case <-gotAck:
				// already closed; don't do it again
				p(ServiceName + ".storage-ack already closed, skipping this time")
			default:
				p(ServiceName + ".storage-ack closing gotAck")
				close(gotAck)
			}
		})

		// start an archiver to catch our request
		tmp, err := ioutil.TempDir("", "test-archiver-filemgr")
		panicOn(err)
		defer os.RemoveAll(tmp)
		fm := NewFileMgr(&ArchiverConfig{WriteDir: tmp,
			ServerList: serverList})
		go func() {
			err := fm.Run()
			panicOn(err)
		}()
		<-fm.Ready
		p("archiver_test: fm.Ready received")
		defer fm.Stop()

		// make the achiving request: write data to server
		subj := ServiceName + ".archive." + streamName

		err = nc.Publish(subj, framed)
		if err != nil {
			panic(fmt.Errorf("Got an error on nc.Publish(): %+v\n", err))
		}
		nc.Flush()
		p("Published to subject '%s' %v bytes", subj, len(framed))
		reply, err := nc.Request(subj, framed, 1000*time.Millisecond)
		if err != nil {
			p("\n ----------->>>>>>> Request got err back: '%s'\n", err)
			panic(fmt.Errorf("Error in Request: %v\n", err))
		}
		p("003ArchiverAck test sent request, got reply: '%#v' with Data '%s'", reply, string(reply.Data))
		cv.So(strings.HasPrefix(reply.Subject, "_INBOX."), cv.ShouldBeTrue)
		// verify reply or panic after timeout
		timeout := 2 * time.Second
		select {
		case <-gotAck:
			// cool
			p("successfully got ack!")
		case <-time.After(timeout):
			panic(fmt.Errorf("timeout after %v waiting for reply", timeout))
		}
	})
}

func PortIsBound(addr string) bool {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

type Logger struct{}

func (c *Logger) String() string {
	return ""
}

func (c *Logger) Errorf(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	fmt.Printf("\n%v ts-archiver.Logger.Errorf(): %s", time.Now(), str)
}

func (c *Logger) Debugf(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	fmt.Printf("\n%v ts-archiver.Logger.Debugf(): %s", time.Now(), str)
}

func (c *Logger) Noticef(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	fmt.Printf("\n%v ts-archiver.Logger.Noticef(): %s", time.Now(), str)
}

func (c *Logger) Tracef(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	fmt.Printf("\n%v ts-archiver.Logger.Tracef(): %s", time.Now(), str)
}

func (c *Logger) Fatalf(format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	fmt.Printf("\n%v ts-archiver.Logger.Fatalf(): %s", time.Now(), str)
}
