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

type Meshc struct {
	objFilename, binFilename string
	optionVerbose bool
	optionStruct bool
}

var meshc Meshc

var Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
}

func init() {
	flag.StringVar(&meshc.binFilename, "output", "a.bin", "output file")
	flag.BoolVar(&meshc.optionVerbose, "verbose", false, "display additional information")
	flag.BoolVar(&meshc.optionStruct, "struct", false, "print file format as a C struct")
}

func main() {
	flag.Parse()

	if meshc.optionStruct {
		fmt.Println(gacc.CStruct())
		os.Exit(0)
	}
	meshc.objFilename = flag.Arg(0)
	mesh := gacc.ParseObj(meshc.objFilename)
	mesh.Encode(meshc.binFilename)

}
