package silent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/alistanis/silentinstall/silent/ui"
)

var (
	Verbose bool
)

// SilentCmd is a command that will run silently
// this can be a regular command or it can be one that expects input from the user
type SilentCmd struct {
	Cmd           *exec.Cmd
	CmdString     string         `json:"cmd"`
	Expectations  []*Expectation `json:"expectations"`
	ReceiveBuffer *bytes.Buffer
	ReadChan      chan string
	ErrChan       chan error
	ErrStringChan chan string
	coloredUI     ui.Ui
}

// Expectation is a structure that stores expected input and output coming from and to another application
type Expectation struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// NewSilentCmd returns a new SilentCmd with all of its fields initialized (except expected cases)
func NewSilentCmd() *SilentCmd {
	return &SilentCmd{
		ReceiveBuffer: bytes.NewBuffer([]byte{}),
		ReadChan:      make(chan string),
		ErrChan:       make(chan error),
		ErrStringChan: make(chan string),
		coloredUI:     ui.NewColoredUi(),
	}
}

// SilentCmds is a slice of *SilentCmd
type SilentCmds []*SilentCmd

// Exec executes all commands stored in s
func (s SilentCmds) Exec() error {
	for _, cmd := range s {
		err := cmd.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// NewSilentCmdsFromJSON loads a list of commands and inputs/outputs from a JSON file
func NewSilentCmdsFromJSON(configData []byte) (SilentCmds, error) {
	cmds := SilentCmds{}
	err := json.Unmarshal(configData, &cmds)
	if err != nil {
		return nil, err
	}
	envMap := make(map[string]string)
	for _, s := range os.Environ() {
		kv := strings.Split(s, "=")
		envMap[kv[0]] = kv[1]
	}
	for _, c := range cmds {
		// because we've loaded from json we have to initialize the command's nil fields here
		c.Init()
		err = c.ExecTemplate(envMap)
		if err != nil {
			return nil, err
		}
		// naive but done for speed of dev
		args := strings.Split(c.CmdString, " ")
		if len(args) > 1 {
			fmt.Println("printing args")
			fmt.Println(args)
			c.Cmd = exec.Command(args[0], args[1:]...)
		} else {
			c.Cmd = exec.Command(args[0])
		}
		c.Cmd.Env = os.Environ()

	}
	return cmds, nil
}

// Init initializes this command's nil fields
func (s *SilentCmd) Init() {
	s.ReceiveBuffer = bytes.NewBuffer([]byte{})
	s.ReadChan = make(chan string)
	s.ErrChan = make(chan error)
	s.ErrStringChan = make(chan string)
	s.coloredUI = ui.NewColoredUi()
}

// ExecTemplate parses a map replacing templated values in the command string
func (s *SilentCmd) ExecTemplate(m map[string]string) error {
	t, err := template.New("envBuilder").Parse(s.CmdString)
	if err == nil {
		w := bytes.NewBuffer([]byte{})
		err = t.Execute(w, m)
		if err != nil {
			return err
		}
		s.CmdString = w.String()
	}
	return nil
}

// Exec executes this SilentCmd, blocking until EOF
func (s *SilentCmd) Exec() error {
	if s.Cmd == nil {
		return errors.New("s.Cmd must not be nil")
	}

	// get stdin, stdout, and stderr
	i, err := s.Cmd.StdinPipe()
	if err != nil {
		return err
	}

	o, err := s.Cmd.StdoutPipe()
	if err != nil {
		return err
	}

	e, err := s.Cmd.StderrPipe()
	if err != nil {
		return err
	}

	closeFunc := func() error {
		err := i.Close()
		if err != nil {
			return err
		}
		err = o.Close()
		if err != nil {
			return err
		}
		return e.Close()
	}

	defer closeFunc()

	err = s.Cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		s.Read(o)
	}()

	go func() {
		s.ReadErr(e)
	}()

	return s.Receive(i)
}

// ReadLine - Deprecated - Reads lines of input from the reader returning a string and error
func (s *SilentCmd) ReadLine(reader io.Reader) (string, error) {
	// read for newline character
	r := bufio.NewReader(reader)

	l, err := r.ReadString('\n')
	// eliminate any pesky \r's because windows
	return strings.Replace(l, "\r", "", -1), err
}

// Write writes l (line) to the provided writer, returning an error if any
func (s *SilentCmd) Write(l string, writer io.Writer) error {
	if !strings.HasSuffix(l, "\n") {
		l = l + "\n"
	}
	_, err := writer.Write([]byte(l))
	return err
}

// Read reads data from reader into s.ReadChan
func (s *SilentCmd) Read(reader io.Reader) {
	s.ReadToChannel(reader, s.ReadChan)
}

// ReadErr reads data from reader into s.ErrStringChan
func (s *SilentCmd) ReadErr(reader io.Reader) {
	s.ReadToChannel(reader, s.ErrStringChan)
}

// ReadToChannel reads from reader to the channel ch
func (s *SilentCmd) ReadToChannel(reader io.Reader, ch chan string) {
	// whoa here's a buffer
	data := make([]byte, 256)
	for {
		bytesRead, err := reader.Read(data)
		if err != nil {
			s.ErrChan <- err
		}
		ch <- string(data[:bytesRead])
		// clear the buffer if necessary - i'd love to see a better/more efficient way to do this
		if bytesRead > 0 {
			data = append(data[bytesRead:], make([]byte, bytesRead)...)
		}
	}
}

// Receive loops on s.ReadChan, s.ErrChan, and s.ErrStringChan, selecting the first that occurs each iteration.
// If readchan receives then we are collecting input from stdout, if there is an error sent to s.ErrChan or s.ErrStringChan,
// we return the error. io.EOF is the expected case when no error actually occurred
func (s *SilentCmd) Receive(w io.Writer) error {
	for {
		select {
		case str := <-s.ReadChan:
			if Verbose {
				// gives more specific info for debugging
				log.Println(str)
			}
			s.coloredUI.Say(str)
			s.ReceiveBuffer.WriteString(str)

			match, expected := s.Match(s.ReceiveBuffer.String())
			if match {
				s.Write(expected.Output, w)
				s.ReceiveBuffer.Reset()
			}
		case err := <-s.ErrChan:
			return err
		case errStr := <-s.ErrStringChan:
			return errors.New(errStr)
		}
	}
}

// Match checks the buffer string against expected cases, removing from the list when one is found
func (s *SilentCmd) Match(bufferString string) (bool, *Expectation) {
	match := false
	expectation := &Expectation{}
	index := 0
	for i, e := range s.Expectations {
		// naive check - thinking about fuzzy matching here but open to ideas.
		// Maybe just check for the exact length of what's expected?
		// Don't want to get caught on possible extra white space though.
		if strings.Contains(bufferString, e.Input) {
			match = true
			expectation = e
			index = i
		}
	}

	if match {
		// pop off so we don't hit duplicates
		s.Expectations[index] = nil
		s.Expectations = append(s.Expectations[:index], s.Expectations[index+1:]...)
	}
	return match, expectation
}
