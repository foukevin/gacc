package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"flag"
	"github.com/foukevin/gacc/meshutil"
	"github.com/foukevin/gacc/meshutil/vector"
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
	mesh := ParseObj(meshc.objFilename)
	Encode(&mesh, meshc.binFilename)
}

var attribsGL3 = [...]VertexAttribDesc{
	{Position, 3, Float32, false},
	{Normal, 3, Int16, true},
	{Color, 3, Uint8, true},
	{Texco0, 2, Float32, false},
}

type glHeader struct {
	name [16]byte
	vertAttribCount, vertAttribOffset uint32
	surfDescCount, surfDescOffset uint32
	vertCount, vertDataOffset, vertDataSize uint32
	indCount, indDataOffset, indDataSize uint32
	aabbCenter, aabbExtent [3]float32
}

func FormatVector(v vector.Vector3, format VertexAttribDesc) []byte {
	buf := new(bytes.Buffer)

	switch (format.Type) {
	case Float32:
		p := [3]float32{ float32(v.X), float32(v.Y), float32(v.Z), }
		binary.Write(buf, binary.LittleEndian, p)
	case Int16:
		binary.Write(buf, binary.LittleEndian, VectorToInt16(v, format.Normalized))
	default:
		panic("oops")
	}
	return buf.Bytes()
}

func VectorToInt16(v vector.Vector3, normalized bool) [3]int16 {
	if normalized {
		v.Normalize()
		v.MultiplyByScalar(MaxInt16)
	}
	return [3]int16{ int16(v.X), int16(v.Y), int16(v.Z) }
}

// used for attribute normalization
const (
	MaxFloat32 = float64(^uint32(0))
	MaxInt32   = float64(^uint32(0) >> 1 - 1)
	MaxInt16   = float64(^uint16(0) >> 1 - 1)
	MaxUint16  = float64(^uint16(0) - 1)
	MaxInt8    = float64(^uint8(0) >> 1 - 1)
	MaxUint8   = float64(^uint8(0) - 1)
)

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

type VertexAttribDesc struct {
	Name VertexAttribName
	Count uint
	Type VertexAttribType
	Normalized bool
}

// struct describing what composes a vertex
type BinaryVertexAttrib struct {
	Index, Count, Type, Normalized uint32
	Stride, Offset uint32
}

type BinarySurfaceDesc struct {
	StartIndex, Count uint32
}
// TODO: rename to Encode() []byte ?
func Encode(m *meshutil.Mesh, filename string) {
	file, _ := os.Create(filename)
	defer file.Close()

	trimesh := MeshToTriangleMesh(m)

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

	// Vertex attributes
	vadata := new(bytes.Buffer)
	for _, va := range binAttribArray {
		// Stride was not computable at first, update it now
		// TODO: couldn't the stride be global to the vertex buffer?
		va.Stride = uint32(stride)
		fmt.Printf("%+v\n", va)
		binary.Write(vadata, binary.LittleEndian, va)
	}


	vertices, surfaces := trimesh.makeBuffers()

	// Vertices
	vdata := new(bytes.Buffer)
	for _, v := range vertices {
		binary.Write(vdata, binary.LittleEndian, FormatVector(v.position, attribsGL3[Position]))
		binary.Write(vdata, binary.LittleEndian, FormatVector(v.normal, attribsGL3[Normal]))
	}

	// Indices
	indexCount := 0
	idata := new(bytes.Buffer)
	sdata := new(bytes.Buffer)
	for _, indices := range surfaces {
		binSurf := BinarySurfaceDesc{
			StartIndex: uint32(indexCount),
			Count: uint32(len(indices)),
		}
		fmt.Printf("Surface descriptor:\n%+v\n", binSurf)
		binary.Write(sdata, binary.LittleEndian, binSurf)

		indexCount += len(indices)
		for _, i := range indices {
			binary.Write(idata, binary.LittleEndian, uint16(i))
		}
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
	header.surfDescCount = uint32(len(surfaces))
	header.surfDescOffset = offset
	offset += uint32(sdata.Len())

	// Vertex data
	header.vertCount = uint32(len(vertices))
	header.vertDataOffset = offset
	header.vertDataSize = uint32(vdata.Len())
	offset += uint32(vdata.Len())

	// Index data
	header.indCount = uint32(indexCount)
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

func ParseObj(filename string) meshutil.Mesh {
	file, _ := os.Open(filename)
	defer file.Close()

	var materialCount uint = 0
	mesh := meshutil.Mesh{ Name: "untitled" }
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
			var p meshutil.Polygon
			for _, s := range val {
				var pos, texco, norm int64
				idx := strings.Split(s, "/")
				pos, _ = strconv.ParseInt(idx[0], 0, 0)
				if len(idx) > 2 {
					texco, _ = strconv.ParseInt(idx[1], 0, 0)
				}
				norm, _ = strconv.ParseInt(idx[len(idx)-1], 0, 0)

				// compensate for indices from obj file starting at 1
				v := meshutil.FaceVert{ int(pos)-1, int(texco)-1, int(norm)-1 }
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

type vertex struct {
	position, normal vector.Vector3
	color0, color1 vector.Vector3
	texco0, texco1 vector.Vector3
	tangent, binormal vector.Vector3
}

type triangle struct {
	vertices [3]vertex
}

type triangleMesh struct {
	vertexAttribNames []VertexAttribName
	surfaces [][]triangle
}

func contains(vertices []vertex, vertex vertex) (bool, int) {
    for i, v := range vertices {
	    if v == vertex { return true, i }
    }
    return false, len(vertices)
}

// TODO: return surfaces?
func MeshToTriangleMesh(m *meshutil.Mesh) *triangleMesh {
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

	trimesh.surfaces = make([][]triangle, m.SurfaceCount)

	for _, p := range m.Polygons {
		var poly []vertex
		for _, f := range p.FaceVerts {
			var newVert vertex
			if has_position {
				newVert.position = m.Positions[f.PositionIndex]
			}
			if has_normal {
				newVert.normal = m.Normals[f.NormalIndex]
			}
			if has_texco {
				newVert.texco0 = m.Texcos[f.TexcoIndex]
			}
			poly = append(poly, newVert)
		}

		triangles := &trimesh.surfaces[p.Material]
		switch (len(poly)) {
		case 4:
			var t, u triangle
			t.vertices[0] = poly[0]
			t.vertices[1] = poly[1]
			t.vertices[2] = poly[2]
			u.vertices[0] = poly[0]
			u.vertices[1] = poly[2]
			u.vertices[2] = poly[3]
			*triangles = append(*triangles, t, u)
		case 3:
			var t triangle
			t.vertices[0] = poly[0]
			t.vertices[1] = poly[1]
			t.vertices[2] = poly[2]
			*triangles = append(*triangles, t)
		}
	}

	return trimesh
}

func (m *triangleMesh) makeBuffers() (vertices []vertex, surfaces [][]int) {
	surfaces = make([][]int, len(m.surfaces))

	for i, triangles := range m.surfaces {
		for _, tri := range triangles {
			for _, vert := range tri.vertices {
				alreadyIn, idx := contains(vertices, vert)
				if !alreadyIn {
					vertices = append(vertices, vert)
				}
				surfaces[i] = append(surfaces[i], idx)
			}
		}
	}

	return
}

