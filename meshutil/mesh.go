package meshutil

import (
	"encoding/binary"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

// An intermediate representation for 3d vectors
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

//
type vertex struct {
	position, normal Vector3
	color0, color1 Vector3
	texco0, texco1 Vector3
	tangent, binormal Vector3
}

type triangle struct {
	vertices [3]vertex
}

type triangleMesh struct {
	vertexAttribNames []VertexAttribName
	surfaces [][]triangle
}

func (m Mesh) toTriangleMesh() *triangleMesh {
	trimesh := new(triangleMesh)
	has_position := len(m.Positions) > 0
	if has_position {
		trimesh.vertexAttribNames = append(trimesh.vertexAttribNames, Position)
	}
	has_normal := len(m.Normals) > 0
	if has_position {
		trimesh.vertexAttribNames = append(trimesh.vertexAttribNames, Normal)
	}
	has_texco := len(m.Texcos) > 0
	if has_texco {
		trimesh.vertexAttribNames = append(trimesh.vertexAttribNames, Texco0)
	}

	// TODO: multiple surfaces
	var triangles []triangle
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
			if has_texco {
				newVert.texco0 = m.Texcos[f.texcoIndex]
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
			triangles = append(triangles, t, u)
		case 3:
			var t triangle
			t.vertices[0] = poly[0]
			t.vertices[1] = poly[1]
			t.vertices[2] = poly[2]
			triangles = append(triangles, t)
		}
	}
	trimesh.surfaces = append(trimesh.surfaces, triangles)

	return trimesh
}

func contains(vertices []vertex, vertex vertex) (bool, int) {
    for i, v := range vertices {
	    if v == vertex { return true, i }
    }
    return false, len(vertices)
}

func (m *triangleMesh) makeBuffers() (vertices []vertex, indices []int) {
	triangles := m.surfaces[0]
	for _, tri := range triangles {
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

type glHeader struct {
	name [16]byte
	vertAttribCount, vertAttribOffset uint32
	surfDescCount, surfDescOffset uint32
	vertCount, vertDataOffset, vertDataSize uint32
	indCount, indDataOffset, indDataSize uint32
	aabbCenter, aabbExtent [3]float32
}

type VertexAttribName uint32
const (
	Position VertexAttribName = iota
        Normal
        Color
	Texco0
	Texco1
	TangentDet
)

type VertexAttribType uint32
const (
	Float32 VertexAttribType = iota
	Int32
	Int16
	Uint16
	Int8
	Uint8
)

// used for attribute normalization
const (
	MaxFloat32 = float64(^uint32(0))
	MaxInt32   = float64(^uint32(0) >> 1 - 1)
	MaxInt16   = float64(^uint16(0) >> 1 - 1)
	MaxUint16  = float64(^uint16(0) - 1)
	MaxInt8    = float64(^uint8(0) >> 1 - 1)
	MaxUint8   = float64(^uint8(0) - 1)
)

type VertexAttribDesc struct {
	Name VertexAttribName
	Count uint
	Type VertexAttribType
	Normalized bool
}

var attribsGL3 = [...]VertexAttribDesc{
	{Position, 3, Float32, false},
	{Normal, 3, Int16, true},
	{Color, 3, Uint8, true},
	{Texco0, 2, Float32, false},
}

// struct describing what composes a vertex
type BinaryVertexAttrib struct {
	Index, Count, Type, Normalized uint32
	Stride, Offset uint32
}

type BinarySurfaceDesc struct {
	Offset, Count uint32
}

func Dot(a, b Vector3) float64 {
	return a.x*b.x + a.y*b.y + a.z*b.z
}

func (v *Vector3) MultiplyByScalar(s float64) {
	*v = Vector3{ v.x*s, v.y*s, v.z*s }
}

func (v Vector3) Magnitude() float64 {
	return math.Sqrt(Dot(v, v))
}

func (v Vector3) Normalized() Vector3 {
	l := v.Magnitude()
	return Vector3{ v.x/l, v.y/l, v.z/l }
}

func (v *Vector3) Normalize() {
	*v = v.Normalized()
}

func (v Vector3) ToInt16(normalized bool) [3]int16 {
	if normalized {
		v.Normalize()
		v.MultiplyByScalar(MaxInt16)
	}
	return [3]int16{ int16(v.x), int16(v.y), int16(v.z) }
}

func (v Vector3) format(format VertexAttribDesc) []byte {
	buf := new(bytes.Buffer)

	switch (format.Type) {
	case Float32:
		p := [3]float32{ float32(v.x), float32(v.y), float32(v.z), }
		binary.Write(buf, binary.LittleEndian, p)
	case Int16:
		binary.Write(buf, binary.LittleEndian, v.ToInt16(format.Normalized))
	default:
		panic("oops")
	}
	return buf.Bytes()
}

// TODO: rename to Encode() []byte ?
func (m *Mesh) WriteOpenGL(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()

	trimesh := m.toTriangleMesh()

	fmt.Println("Vertex attributes: ", trimesh.vertexAttribNames)
	stride := 0
	var binAttribArray []BinaryVertexAttrib
	for _, va := range trimesh.vertexAttribNames {
		desc := attribsGL3[va]
		fmt.Printf("%+v\n", attribsGL3[va])
		normalized := 0
		if desc.Normalized {
			normalized = 1
		}
		binAttrib := BinaryVertexAttrib{
			Index: uint32(desc.Name),
			Count: uint32(desc.Count),
			Type: uint32(desc.Type),
			Normalized: uint32(normalized),
			Offset: uint32(stride),
		}
		binAttribArray = append(binAttribArray, binAttrib)
		// stride is not complete yet and correspond to the
		// current attribute's offset
		stride += binary.Size(binAttrib)
	}

	vadata := new(bytes.Buffer)
	for _, va := range binAttribArray {
		// Stride was not computable at first, update it now
		// TODO: couldn't the stride be global to the vertex buffer?
		va.Stride = uint32(stride)
		fmt.Printf("%+v\n", va)
		binary.Write(vadata, binary.LittleEndian, va)
	}

	vertices, indices := trimesh.makeBuffers()
	vdata := new(bytes.Buffer)
	for _, v := range vertices {
		binary.Write(vdata, binary.LittleEndian, v.position.format(attribsGL3[Position]))
		binary.Write(vdata, binary.LittleEndian, v.normal.format(attribsGL3[Normal]))
	}

	// TODO: multiple surfaces
	sdata := new(bytes.Buffer)
	binSurf := BinarySurfaceDesc{ Offset: 0, Count: uint32(len(indices)) }
	binary.Write(sdata, binary.LittleEndian, binSurf)

	idata := new(bytes.Buffer)
	for _, i := range indices {
		binary.Write(idata, binary.LittleEndian, uint16(i))
	}

	var offset uint32
	var header glHeader
	copy(header.name[:len(header.name)-1], m.Name)
	offset += uint32(binary.Size(header))

	// Vertex attributes
	header.vertAttribCount = uint32(len(trimesh.vertexAttribNames))
	header.vertAttribOffset = offset
	offset += uint32(vadata.Len())

	// Surface descriptors
	header.surfDescCount = uint32(len(trimesh.surfaces))
	header.surfDescOffset = offset
	offset += uint32(sdata.Len())

	// Vertex data
	header.vertCount = uint32(len(vertices))
	header.vertDataOffset = offset
	header.vertDataSize = uint32(vdata.Len())
	offset += uint32(vdata.Len())

	// Index data
	header.indCount = uint32(len(indices))
	header.indDataOffset = offset
	header.indDataSize = uint32(idata.Len())
	offset += uint32(idata.Len())

	fmt.Printf("%+v\n", header)

	hdata := new(bytes.Buffer)
	binary.Write(hdata, binary.LittleEndian, header)

	file.Write(hdata.Bytes())
	file.Write(vadata.Bytes())
	file.Write(sdata.Bytes())
	file.Write(vdata.Bytes())
	file.Write(idata.Bytes())
}

func ParseObj(filename string) Mesh {
	file, _ := os.Open(filename)
	defer file.Close()

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
			v := Vector3{ x, y, z }
			if (ident == "v") {
				mesh.Positions = append(mesh.Positions, v)
			} else {
				mesh.Normals = append(mesh.Normals, v)
			}
		case "vt":
			s, _ := strconv.ParseFloat(val[0], 64)
			t, _ := strconv.ParseFloat(val[1], 64)
			v := Vector3{ s, t, 0 }
			mesh.Texcos = append(mesh.Texcos, v)
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
