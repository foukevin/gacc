package main

import (
	"flag"
	//	"fmt"

	"github.com/foukevin/gacc/meshutil"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Meshc struct {
	objFilename, binFilename string
	optionVerbose bool
}

var meshc Meshc

func init() {
	flag.StringVar(&meshc.binFilename, "output", "a.bin", "output file")
	flag.BoolVar(&meshc.optionVerbose, "verbose", false, "display additional information")
}

func main() {
	flag.Parse()
	meshc.objFilename = flag.Arg(0)
	mesh := meshutil.ParseObj(meshc.objFilename)
	mesh.WriteOpenGL(meshc.binFilename)
}
