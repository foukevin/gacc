package meshutil

import (
	"github.com/foukevin/gacc/meshutil/vector"
)

type FaceVert struct {
	PositionIndex, TexcoIndex, NormalIndex int
}

type Polygon struct {
	FaceVerts []FaceVert
	Material uint
}

// A generic mesh structure
type Mesh struct {
	Name string
	SurfaceCount uint
	Positions, Texcos, Normals []vector.Vector3
	Polygons []Polygon
}
