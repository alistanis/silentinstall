package silent

import (
	"bytes"
	"testing"

	"io/ioutil"
	"os"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	repoPath     = "/src/github.com/alistanis/silentinstall"
	testDataPath = repoPath + "/silent/test_data"
)

func init() {
	Verbose = true
}

func TestSilentCmd_ReadLine(t *testing.T) {

	Convey("We can test the silent command's readline function", t, func() {
		s := &SilentCmd{}
		reader := bytes.NewReader([]byte("Hello!\n"))
		l, err := s.ReadLine(reader)
		So(err, ShouldBeNil)
		So(l, ShouldEqual, "Hello!\n")
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
		cmds, err := NewSilentCmds(data)
		So(err, ShouldBeNil)
		err = cmds.Exec()
		So(err, ShouldBeNil)
	})
}

func TestSilentCmd_Exec(t *testing.T) {
	Convey("We can load up a new set of configs execute them, and read/write input to the cmd", t, func() {
		data, err := loadWaitTestConfig()
		So(err, ShouldBeNil)
		cmds, err := NewSilentCmds(data)
		So(err, ShouldBeNil)
		err = cmds.Exec()
		So(err, ShouldBeNil)
	})
}

func TestMultipleCommands(t *testing.T) {
	Convey("We can load a config and execute multiple io ops", t, func() {
		data, err := loadMultipleIOConfig()
		So(err, ShouldBeNil)
		cmds, err := NewSilentCmds(data)
		So(err, ShouldBeNil)
		err = cmds.Exec()
		So(err, ShouldBeNil)
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

func loadConfig(path string) ([]byte, error) {
	gopath := os.Getenv("GOPATH")
	return ioutil.ReadFile(gopath + testDataPath + path)
}
