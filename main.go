package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"io"

	"github.com/alistanis/silentinstall/silent"
	"github.com/alistanis/silentinstall/silent/ui"
)

const (
	configVarMsg = "The path of the config file"
	verboseMsg   = "Prints verbose output if true"
)

var (
	configFile = flag.String("f", "", configVarMsg)
)

func init() {
	flag.StringVar(configFile, "file", "", configVarMsg)
	flag.BoolVar(&silent.Verbose, "v", false, verboseMsg)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func parseFlags() {
	flag.Parse()
	if *configFile == "" {
		fmt.Println("Must provide -f or --file for the path of the config file to use.")
		os.Exit(-1)
	}
}

func main() {
	parseFlags()

	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	cmds, err := silent.NewSilentCmdsFromJSON(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	ui := ui.NewColoredUi()
	err = cmds.Exec()
	if err == io.EOF {
		ui.Say("SilentInstall has finished successfully!")
	} else {
		ui.Error(err.Error())
		os.Exit(-1)
	}
}
