package silent

import (
	"bytes"
	"testing"

	"io/ioutil"
	"os"

	"io"

	"fmt"

	"strings"

	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	repoPath     = "/src/github.com/alistanis/silentinstall"
	testDataPath = repoPath + "/silent/test_data"
)

func init() {
	Verbose = true
}

func TestSilentCmd_Read(t *testing.T) {
	Convey("We can test the silent command's read function", t, func() {
		s := NewSilentCmd()
		reader := strings.NewReader("Hello!")
		go func() {
			s.Read(reader)
		}()

		// this is kinda dangerous in most cases but we're only testing the read,
		// and to fully test the read function we need to receive it on the buffer and inspect it
		err := s.Receive(nil)

		So(err, ShouldEqual, io.EOF)
		fmt.Println(s.ReceiveBuffer.String())
		fmt.Println(s.ReceiveBuffer.Bytes())
		So(s.ReceiveBuffer.String(), ShouldEqual, "Hello!")
	})

	Convey("We can test the silent command's read function with EOF", t, func() {

		s := NewSilentCmd()
		r, w := io.Pipe()

		Convey("We can read asyncronously with a pipe, not looking for end of line", func() {
			go func() {
				// make sure we can still receive on the pipe while waiting for EOF
				w.Write([]byte("Hello!!!\n"))
				time.Sleep(time.Duration(1) * time.Second)
				w.Write([]byte("Hi!"))
				w.Close()
			}()
			go func() {
				s.Read(r)
			}()
			// nothing will be written to w because there won't be a match because there are no expected cases
			err := s.Receive(w)
			So(err, ShouldEqual, io.EOF)

		})

	})
}

func TestSilentCmd_Write(t *testing.T) {
	Convey("We can test the silent command's write function", t, func() {
		s := &SilentCmd{}
		writer := bytes.NewBuffer([]byte{})
		err := s.Write("Hello!", writer)
		So(err, ShouldBeNil)
		So(writer.String(), ShouldEqual, "Hello!\n")
	})
}

func TestNewSilentCmds(t *testing.T) {

	Convey("We can load up a new set of silent cmds from a json config", t, func() {
		data, err := loadBasicTestConfig()
		So(err, ShouldBeNil)
		cmds, err := NewSilentCmdsFromJSON(data)
		So(err, ShouldBeNil)
		err = cmds.Exec()
		So(err, ShouldEqual, io.EOF)
	})
}

func TestSilentCmd_Exec(t *testing.T) {
	Convey("We can load up a new set of configs execute them, and read/write input to the cmd", t, func() {
		data, err := loadWaitTestConfig()
		So(err, ShouldBeNil)
		cmds, err := NewSilentCmdsFromJSON(data)
		So(err, ShouldBeNil)
		err = cmds.Exec()
		So(err, ShouldEqual, io.EOF)
	})

	Convey("We can load up a new set of configs execute them, and read/write input to the cmd with no newlines", t, func() {
		data, err := loadNoNewlineTestConfig()
		So(err, ShouldBeNil)
		cmds, err := NewSilentCmdsFromJSON(data)
		So(err, ShouldBeNil)
		err = cmds.Exec()
		So(err, ShouldEqual, io.EOF)
	})
}

func TestMultipleCommands(t *testing.T) {
	Convey("We can load a config and execute multiple io ops", t, func() {
		data, err := loadMultipleIOConfig()
		So(err, ShouldBeNil)
		cmds, err := NewSilentCmdsFromJSON(data)
		So(err, ShouldBeNil)
		err = cmds.Exec()
		So(err, ShouldEqual, io.EOF)
	})
}

func loadBasicTestConfig() ([]byte, error) {
	return loadConfig("/basic_example_config.json")
}

func loadWaitTestConfig() ([]byte, error) {
	return loadConfig("/wait_example_config.json")
}

func loadMultipleIOConfig() ([]byte, error) {
	return loadConfig("/multiple_io_example_config.json")
}

func loadNoNewlineTestConfig() ([]byte, error) {
	return loadConfig("/no_newline_example_config.json")
}

func loadConfig(path string) ([]byte, error) {
	gopath := os.Getenv("GOPATH")
	return ioutil.ReadFile(gopath + testDataPath + path)
}
