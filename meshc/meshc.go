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

var outPath string

func init() {
	flag.StringVar(&outPath, "output", "a.bin", "output file")
}

func main() {
	flag.Parse()
	objPath := flag.Arg(0)
	mesh := meshutil.ParseObj(objPath)
	mesh.WriteOpenGL(outPath)
}
