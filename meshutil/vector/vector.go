package vector

import (
	"math"
)

type Vector3 struct {
	X, Y, Z float64
}

func Dot(a, b Vector3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (v *Vector3) MultiplyByScalar(s float64) {
	*v = Vector3{ v.X*s, v.Y*s, v.Z*s }
}

func Substract(a, b Vector3) Vector3 {
	return Vector3{ a.X-b.X, a.Y-b.Y, a.Z-b.Z }
}

func (v Vector3) Magnitude() float64 {
	return math.Sqrt(Dot(v, v))
}

func (v Vector3) Normalized() Vector3 {
	l := v.Magnitude()
	return Vector3{ v.X/l, v.Y/l, v.Z/l }
}

func (v *Vector3) Normalize() {
	*v = v.Normalized()
}

