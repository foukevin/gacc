package meshutil

import (
	"encoding/binary"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Vector3 struct {
	x, y, z float64
}

type FaceVert struct {
	positionIndex, texcoIndex, normalIndex int
}

type Polygon struct {
	verts []FaceVert
}

type Mesh struct {
	Name string
	Positions, Texcos, Normals []Vector3
	Polygons []Polygon
}

type vertex struct {
	position, normal Vector3
	color0, color1 Vector3
	texco0, texco1 Vector3
	tangent, binormal Vector3
}

type vertexAttributes struct {
	position, normal bool
	color0, color1 bool
	texco0, texco1 bool
	tangent, binormal bool
}

type triangle struct {
	vertices [3]vertex
}

type triangleMesh struct {
	attributes vertexAttributes
	triangles []triangle
}

func (m *Mesh) toTriangleMesh() *triangleMesh {
	trimesh := new(triangleMesh)
	has_position := len(m.Positions) > 0
	trimesh.attributes.position = has_position
	has_normal := len(m.Normals) > 0
	trimesh.attributes.normal = has_normal

	for _, p := range m.Polygons {
		var poly []vertex
		for _, f := range p.verts {
			var newVert vertex
			if has_position {
				newVert.position = m.Positions[f.positionIndex]
			}
			if has_normal {
				newVert.normal = m.Normals[f.normalIndex]
			}
			poly = append(poly, newVert)
		}

		switch (len(poly)) {
		case 4:
			var t, u triangle
			t.vertices[0] = poly[0]
			t.vertices[1] = poly[1]
			t.vertices[2] = poly[2]
			u.vertices[0] = poly[0]
			u.vertices[1] = poly[2]
			u.vertices[2] = poly[3]
			trimesh.triangles = append(trimesh.triangles, t, u)
		case 3:
			var t triangle
			t.vertices[0] = poly[0]
			t.vertices[1] = poly[1]
			t.vertices[2] = poly[2]
			trimesh.triangles = append(trimesh.triangles, t)
		}
	}

	return trimesh
}
/*
func (m *Mesh) parseObjLine(line string) {
}
*/
func contains(vertices []vertex, vertex vertex) (bool, int) {
    for i, v := range vertices {
	    if v == vertex { return true, i }
    }
    return false, len(vertices)
}

func (m *triangleMesh) makeBuffers() (vertices []vertex, indices []int) {
	for _, tri := range m.triangles {
		for _, vert := range tri.vertices {
			alreadyIn, idx := contains(vertices, vert)
			if !alreadyIn {
				vertices = append(vertices, vert)
			}
			indices = append(indices, idx)
		}
	}

	return
}

func (v *vertex) format() []byte {
	buf := new(bytes.Buffer)
	pos := [3]float32{
		float32(v.position.x),
		float32(v.position.y),
		float32(v.position.z),
	}
	norm := [3]float32{
		float32(v.normal.x),
		float32(v.normal.y),
		float32(v.normal.z),
	}
	binary.Write(buf, binary.LittleEndian, pos)
	binary.Write(buf, binary.LittleEndian, norm)
	return buf.Bytes()
}

// TODO: use triangleMesh instead of Mesh here?
func (m *Mesh) WriteOpenGL(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()

	vertices, indices := m.toTriangleMesh().makeBuffers()

	vdata := new(bytes.Buffer)
	for _, v := range vertices {
		binary.Write(vdata, binary.LittleEndian, v.format())
	}

	idata := new(bytes.Buffer)
	for _, i := range indices {
		binary.Write(idata, binary.LittleEndian, uint16(i))
	}

	hdr := new(bytes.Buffer)
	binary.Write(hdr, binary.LittleEndian, uint32(len(vertices)))
	binary.Write(hdr, binary.LittleEndian, uint32(vdata.Len()))
	binary.Write(hdr, binary.LittleEndian, uint32(len(indices)))
	binary.Write(hdr, binary.LittleEndian, uint32(idata.Len()))

	fmt.Println("vertex count: ", len(vertices))
	fmt.Println("vertex data size: ", vdata.Len())
	//fmt.Println(vertices)
	fmt.Println("index data size: ", idata.Len())
	fmt.Println("index count: ", len(indices))
	//fmt.Println(indices)

	file.Write(hdr.Bytes())
	file.Write(vdata.Bytes())
	file.Write(idata.Bytes())
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
		ident, val := line[0], line[1:]
		switch (ident) {
		case "v", "vn":
			x, _ := strconv.ParseFloat(val[0], 64)
			y, _ := strconv.ParseFloat(val[1], 64)
			z, _ := strconv.ParseFloat(val[2], 64)
			v := Vector3{ x, y, z }
			if (ident == "v") {
				mesh.Positions = append(mesh.Positions, v)
			} else {
				mesh.Normals = append(mesh.Normals, v)
			}
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
				p.verts = append(p.verts, v)
			}
			mesh.Polygons = append(mesh.Polygons, p)
		default:
			fmt.Println(ident + " not parsed yet")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return mesh
}
