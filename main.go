package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"io"

	"path/filepath"

	"github.com/alistanis/silentinstall/silent"
	"github.com/alistanis/silentinstall/silent/ui"
)

// setup const messages
const (
	configVarMsg = "The path of the config file"
	verboseMsg   = "Prints verbose output if true"
)

var (
	configFile = flag.String("f", "", configVarMsg)
	coloredUi  = ui.NewColoredUi()
)

const (
	_ = iota // skip 0
	// starting at -1, decrement for each additional value
	exitNoFileProvided = -iota
	exitBadFile
	exitBadConfig
	exitCmdError
)

// set our flagvars
func init() {
	flag.StringVar(configFile, "file", "", configVarMsg)
	flag.BoolVar(&silent.Verbose, "v", false, verboseMsg)
}

// parse those flags
func parseFlags() {
	flag.Parse()
	if silent.Verbose {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	}
	if *configFile == "" {
		coloredUi.Err("Must provide -f or --file for the path of the config file to use.")
		os.Exit(exitNoFileProvided)
	}
}

func main() {
	parseFlags()

	file := filepath.Clean(*configFile)
	// read config data
	data, err := ioutil.ReadFile(file)
	if err != nil {
		coloredUi.Err(err)
		os.Exit(exitBadFile)
	}

	// convert json to list of commands
	cmds, err := silent.NewSilentCmdsFromJSON(data)
	if err != nil {
		coloredUi.Err(err)
		os.Exit(exitBadConfig)
	}
	// execute them!
	err = cmds.Exec()
	if err == io.EOF {
		coloredUi.Say("SilentInstall has finished successfully!")
	} else {
		coloredUi.Err(err)
		os.Exit(exitCmdError)
	}
}
