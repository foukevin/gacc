package gacc

import (
	"fmt"
	"log"
	"bufio"
	"os"
	"strings"
	"strconv"
	"github.com/foukevin/gacc/vector"
)

func ParseObj(filename string) Mesh {
	file, _ := os.Open(filename)
	defer file.Close()

	var materialCount uint = 0
	mesh := Mesh{ Name: "untitled" }
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		ident, val := line[0], line[1:]
		switch (ident) {
		case "v", "vn":
			x, _ := strconv.ParseFloat(val[0], 64)
			y, _ := strconv.ParseFloat(val[1], 64)
			z, _ := strconv.ParseFloat(val[2], 64)
			v := vector.Vector3{ x, y, z }
			if (ident == "v") {
				mesh.Positions = append(mesh.Positions, v)
			} else {
				mesh.Normals = append(mesh.Normals, v)
			}
		case "vt":
			s, _ := strconv.ParseFloat(val[0], 64)
			t, _ := strconv.ParseFloat(val[1], 64)
			v := vector.Vector3{ s, t, 0 }
			mesh.Texcos = append(mesh.Texcos, v)
		case "usemtl":
			fmt.Println(ident + " new material found")
			materialCount++
		case "f":
			var p Polygon
			for _, s := range val {
				var pos, texco, norm int64
				idx := strings.Split(s, "/")
				pos, _ = strconv.ParseInt(idx[0], 0, 0)
				if len(idx) > 2 {
					texco, _ = strconv.ParseInt(idx[1], 0, 0)
				}
				norm, _ = strconv.ParseInt(idx[len(idx)-1], 0, 0)

				// compensate for indices from obj file starting at 1
				v := FaceVert{ int(pos)-1, int(texco)-1, int(norm)-1 }
				p.FaceVerts = append(p.FaceVerts, v)
				p.Material = materialCount - 1
			}
			mesh.Polygons = append(mesh.Polygons, p)
		default:
			fmt.Println(ident + " not parsed yet")
		}
	}
	mesh.SurfaceCount = materialCount

	fmt.Printf("%+v\n", mesh)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return mesh
}
