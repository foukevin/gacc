package main

import (
	"fmt"
	"flag"
	"github.com/foukevin/gacc"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Scenec struct {
	sceneFilename, binFilename string
	optionVerbose bool
	optionStruct bool
}

var scenec Scenec

var Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
}

func init() {
	flag.StringVar(&scenec.binFilename, "output", "a.scene.o", "output file")
	flag.BoolVar(&scenec.optionVerbose, "verbose", false, "display additional information")
	flag.BoolVar(&scenec.optionStruct, "struct", false, "print file format as a C struct")
}

func main() {
	flag.Parse()

	if scenec.optionStruct {
		os.Exit(0)
	}
	scenec.sceneFilename = flag.Arg(0)
	_ = gacc.ParseJafScene(scenec.sceneFilename)
	//scene.Encode(scenec.binFilename)
}
