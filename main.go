package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alistanis/silentinstall/silent"
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
	flag.BoolVar(&silent.Verbose, "v", false, verboseMsg)

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

	cmds, err := silent.NewSilentCmds(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	err = cmds.Exec()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

}
