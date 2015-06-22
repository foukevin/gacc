package gacc

import (
	"reflect"
	"io"
	"os"
	"encoding/binary"
	"bytes"
	"fmt"
	"math"
	"github.com/foukevin/gacc/vector"
)

type BinaryMeshHeader struct {
	Name [16]byte
	VertAttribCount, VertAttribOffset uint32
	SurfDescCount, SurfDescOffset uint32
	VertCount, VertDataOffset, VertDataSize uint32
	IndCount, IndDataOffset, IndDataSize uint32
	AabbCenter, AabbExtent [3]float32
}

// struct describing what composes a vertex
type BinaryVertexAttrib struct {
	Index, Count, Type, Normalized uint32
	Stride, Offset uint32
}

type BinarySurfaceDesc struct {
	StartIndex, Count uint32
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

func (t VertexAttribType)ByteSize() (size uint) {
	switch (t) {
	case Float32, Int32:
		size = 4
	case Int16, Uint16:
		size = 2
	case Int8, Uint8:
		size = 1
	}
	return
}

type VertexAttribDesc struct {
	Name VertexAttribName
	Count uint
	Type VertexAttribType
	Normalized bool
}

// TODO: rename to Encode() []byte ?
func (m *Mesh) Encode(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()

	trimesh := MeshToTriangleMesh(m)

	var stride uint
	var binAttribArray []BinaryVertexAttrib
	for _, va := range trimesh.vertexAttribNames {
		desc := attributes[va]
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
		stride += desc.Count * desc.Type.ByteSize()
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
	aabbCenter, aabbExtent := AxisAlignedBoudingBox(vertices)
	fmt.Printf("AABB: %+v %+v\n", aabbCenter, aabbExtent)

	// Vertices
	vdata := new(bytes.Buffer)
	for _, v := range vertices {
		binary.Write(vdata, binary.LittleEndian, VectorToBytes(v.position, attributes[Position]))
		binary.Write(vdata, binary.LittleEndian, VectorToBytes(v.normal, attributes[Normal]))
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
		binary.Write(sdata, binary.LittleEndian, binSurf)

		indexCount += len(indices)
		for _, i := range indices {
			binary.Write(idata, binary.LittleEndian, uint16(i))
		}
	}

	var offset uint32
	var header BinaryMeshHeader
	copy(header.Name[:len(header.Name)-1], m.Name)
	offset += uint32(binary.Size(header))

	// Vertex attributes
	header.VertAttribCount = uint32(len(trimesh.vertexAttribNames))
	header.VertAttribOffset = offset
	offset += uint32(vadata.Len())

	// Surface descriptors
	header.SurfDescCount = uint32(len(surfaces))
	header.SurfDescOffset = offset
	offset += uint32(sdata.Len())

	// Vertex data
	header.VertCount = uint32(len(vertices))
	header.VertDataOffset = offset
	header.VertDataSize = uint32(vdata.Len())
	offset += uint32(vdata.Len())

	// Index data
	header.IndCount = uint32(indexCount)
	header.IndDataOffset = offset
	header.IndDataSize = uint32(idata.Len())
	offset += uint32(idata.Len())

	// AABB
	header.AabbCenter = [3]float32 {
		float32(aabbCenter.X),
		float32(aabbCenter.Y),
		float32(aabbCenter.Z),
	}
	header.AabbExtent = [3]float32 {
		float32(aabbExtent.X),
		float32(aabbExtent.Y),
		float32(aabbExtent.Z),
	}

	fmt.Printf("%+v\n", header)

	hdata := new(bytes.Buffer)
	binary.Write(hdata, binary.LittleEndian, header)

	file.Write(hdata.Bytes())
	file.Write(vadata.Bytes())
	file.Write(sdata.Bytes())
	file.Write(vdata.Bytes())
	file.Write(idata.Bytes())
}

var attributes = [...]VertexAttribDesc{
	{Position, 3, Float32, false},
	{Normal, 3, Int16, true},
	{Color, 3, Uint8, true},
	{Texco0, 2, Float32, false},
}

func VectorToBytes(v vector.Vector3, format VertexAttribDesc) []byte {
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

func AxisAlignedBoudingBox(vertices []vertex) (center, extent vector.Vector3) {
	min := vector.Vector3{ math.Inf(1), math.Inf(1), math.Inf(1) }
	max := vector.Vector3{ math.Inf(-1), math.Inf(-1), math.Inf(-1) }

	for _, v := range vertices {
		p := v.position
		if p.X < min.X { min.X = p.X }
		if p.Y < min.Y { min.Y = p.Y }
		if p.Z < min.Z { min.Z = p.Z }
		if p.X > max.X { max.X = p.X }
		if p.Y > max.Y { max.Y = p.Y }
		if p.Z > max.Z { max.Z = p.Z }
	}

	fmt.Println(min)
	fmt.Println(max)

	center.X = (min.X + max.X) / 2
	center.Y = (min.Y + max.Y) / 2
	center.Z = (min.Z + max.Z) / 2

	min = vector.Substract(min, center)
	max = vector.Substract(max, center)

	extent.X = math.Max(math.Abs(min.X), math.Abs(max.X))
	extent.Y = math.Max(math.Abs(min.Y), math.Abs(max.Y))
	extent.Z = math.Max(math.Abs(min.Z), math.Abs(max.Z))
	return
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

type vertex struct {
	position, normal vector.Vector3
	color0, color1 vector.Vector3
	texco0, texco1 vector.Vector3
	tangent, binormal vector.Vector3
}

// TODO: return surfaces?
func MeshToTriangleMesh(m *Mesh) *triangleMesh {
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

func ReadBinaryMeshHeader(filename string) (BinaryMeshHeader, []BinaryVertexAttrib) {
	file, _ := os.Open(filename)
	defer file.Close()

	var header BinaryMeshHeader

	sr := io.NewSectionReader(file, 0, 1<<63-1)
	sr.Seek(0, os.SEEK_SET)
	binary.Read(sr, binary.LittleEndian, &header)

	vertAttribs := make([]BinaryVertexAttrib, header.VertAttribCount)
	sr.Seek(int64(header.VertAttribOffset), os.SEEK_SET)

	for i := 0; i < int(header.VertAttribCount); i++ {
		binary.Read(sr, binary.LittleEndian, &vertAttribs[i])
	}

	return header, vertAttribs
}

func CStruct() string {
	var t BinaryMeshHeader
	s := reflect.ValueOf(&t).Elem()

	res := "struct Mesh {\n"

	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		var carray, ctype string
		var baseType reflect.Type
		if f.Kind() == reflect.Array {
			carray = fmt.Sprintf("[%d]", f.Len())
			baseType = f.Type().Elem()
		} else {
			baseType = f.Type()
		}

		switch (baseType.Kind()) {
		case reflect.Float32:
			ctype = "float"
		case reflect.Uint32, reflect.Int32, reflect.Uint16, reflect.Int16, reflect.Uint8, reflect.Int8:
			ctype = baseType.Name() + "_t"
		default:
			ctype = "unknown_type"
		}
		res += fmt.Sprintf("\t%s %s%s;\n", ctype, typeOfT.Field(i).Name, carray)
	}

	return res + "};"
}
