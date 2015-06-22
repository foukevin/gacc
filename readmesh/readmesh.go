package main

import (
	"fmt"
	"flag"
	"github.com/foukevin/gacc"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type ReadMesh struct {
	binFilename string
	optionVerbose bool
}

var readMesh ReadMesh

func init() {
	flag.BoolVar(&readMesh.optionVerbose, "verbose", false, "display additional information")
}

func main() {
	flag.Parse()
	readMesh.binFilename = flag.Arg(0)
	header, vertAttribs := gacc.ReadBinaryMeshHeader(readMesh.binFilename)

	fmt.Println("Name: " + string(header.Name[:]))
	fmt.Printf("Vertex attribute count   : %d\n", header.VertAttribCount)
	fmt.Printf("Vertex attribute offset  : %#d (%d)\n", header.VertAttribOffset, header.VertAttribOffset)
	fmt.Printf("Surface descriptor count : %d\n", header.SurfDescCount)
	fmt.Printf("Surface descriptor offset: %#x (%d)\n", header.SurfDescOffset, header.SurfDescOffset)
	fmt.Printf("Vertex count             : %d\n", header.VertCount)
	fmt.Printf("Vertex data offset       : %#x (%d)\n", header.VertDataOffset, header.VertDataOffset)
	fmt.Printf("Vertex data size         : %d\n", header.VertDataSize)
	fmt.Printf("Index count              : %d (%d triangles)\n", header.IndCount, header.IndCount / 3)
	fmt.Printf("Index data offset        : %#x (%d)\n", header.IndDataOffset, header.IndDataOffset)
	fmt.Printf("Index data size          : %d\n", header.IndDataSize)
	fmt.Printf("Bounding box center      : %+v\n", header.AabbCenter)
	fmt.Printf("Bounding box extent      : %+v\n", header.AabbExtent)

	fmt.Println("Vertex attributes:")
	for _, va := range vertAttribs {
		fmt.Printf("  %+v\n", va)
	}
}



