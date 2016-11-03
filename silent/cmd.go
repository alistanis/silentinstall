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
)

var (
	Verbose bool
)

type SilentCmd struct {
	Cmd           *exec.Cmd
	CmdString     string `json:"cmd"`
	ExpectedCases []*IO  `json:"io"`
}

type SilentCmds []*SilentCmd

func (s SilentCmds) Exec() error {
	for _, cmd := range s {
		err := cmd.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func NewSilentCmds(configData []byte) (SilentCmds, error) {
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
		t, err := template.New("envBuilder").Parse(c.CmdString)
		if err == nil {
			w := bytes.NewBuffer([]byte{})
			err = t.Execute(w, envMap)
			if err != nil {
				return nil, err
			}
			c.CmdString = w.String()
		}
		// naive but done for speed of dev
		args := strings.Split(c.CmdString, " ")
		fmt.Println(args)
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

type IO struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func (s *SilentCmd) Exec() error {
	if s.Cmd == nil {
		return errors.New("s.Cmd must not be nil")
	}

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

	errChan := make(chan error)
	finishedChan := make(chan bool)

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

	err = s.Cmd.Start()
	if err != nil {
		return err
	}

	// loop through cmd output and write to command input if there is a match
	// currently expects exact matches
	go func() {
		for {
			l, err := s.ReadLine(o)
			if Verbose {
				log.Println(l)
			}
			if err != nil {
				if err == io.EOF {
					finishedChan <- true
					break
				} else {
					errChan <- err
					break
				}
			}
			for _, inputOutput := range s.ExpectedCases {
				if strings.Contains(l, inputOutput.Input) {
					err = s.Write(inputOutput.Output, i)
					if err != nil {
						errChan <- err
						break
					}
				}
			}
		}
	}()

	// loop through stderr, assemble any lines that are read until it is closed, write to err chan combined lines
	// or a separate error if one occured while reading
	go func() {
		errLines := bytes.NewBuffer([]byte{})
		for {
			l, err := s.ReadLine(e)
			if err != nil {
				if err == io.EOF {
					if errLines.Len() > 0 {
						if Verbose {
							log.Println(errLines)
						}
						errChan <- errors.New(errLines.String())
						break
					}
				} else {
					errChan <- err
					break
				}
			}
			errLines.WriteString(l)
		}
	}()

	select {
	case e := <-errChan:
		if Verbose {
			log.Println(e)
		}
		return e
	case _ = <-finishedChan:
		err = closeFunc()
		if err != nil {
			return err
		}
	}
	return s.Cmd.Wait()
}

func (s *SilentCmd) ReadLine(reader io.Reader) (string, error) {
	// read for newline character
	r := bufio.NewReader(reader)

	l, err := r.ReadString('\n')
	// eliminate any pesky \r's because windows
	return strings.Replace(l, "\r", "", -1), err
}

func (s *SilentCmd) Write(l string, writer io.Writer) error {
	if !strings.HasSuffix(l, "\n") {
		l = l + "\n"
	}
	_, err := writer.Write([]byte(l))
	return err
}
