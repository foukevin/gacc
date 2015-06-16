package meshutil

import (
	"encoding/binary"
	"bufio"
	"bytes"
//	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Vector3 struct {
	x, y, z float64
}

type Mesh struct {
	Name string
	Vertices []Vector3
}

/*
func (m *Mesh) parseObjLine(line string) {
}
*/

func (m Mesh) WriteOpenGL(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()

	buf := new(bytes.Buffer)
	numVerts := uint32(len(m.Vertices))
	binary.Write(buf, binary.LittleEndian, numVerts)

	for _, v := range m.Vertices {
		v32 := [3]float32{ float32(v.x), float32(v.y), float32(v.z) }
		binary.Write(buf, binary.LittleEndian, v32)
	//	fmt.Println(v)
	}
	file.Write(buf.Bytes())
}

func ParseObj(filename string) Mesh {
	file, _ := os.Open(filename)
	defer file.Close()

	mesh := Mesh{
		Name: "woot",
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		if line[0] == "v" {
			x, _ := strconv.ParseFloat(line[1], 64)
			y, _ := strconv.ParseFloat(line[2], 64)
			z, _ := strconv.ParseFloat(line[3], 64)
			v := Vector3{ x, y, z }
			mesh.Vertices = append(mesh.Vertices, v)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return mesh
}
