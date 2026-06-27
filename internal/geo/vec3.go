package geo

import "math"

type Vec3 struct {
	X, Y, Z float64
}

func NewVec3(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

func (v Vec3) Norm() Vec3 {
	d := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	return Vec3{v.X / d, v.Y / d, v.Z / d}
}

func (v Vec3) Dot(u Vec3) float64 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z
}

func (v Vec3) Add(u Vec3) Vec3 {
	return Vec3{v.X + u.X, v.Y + u.Y, v.Z + u.Z}
}

func (v Vec3) Sub(u Vec3) Vec3 {
	return Vec3{v.X - u.X, v.Y - u.Y, v.Z - u.Z}
}

func (v Vec3) Div(u Vec3) Vec3 {
	return Vec3{v.X / u.X, v.Y / u.Y, v.Z / u.Z}
}

func (v Vec3) MinComp() float64 {
	return math.Min(v.X, math.Min(v.Y, v.Z))
}

func (v Vec3) MaxComp() float64 {
	return math.Max(v.X, math.Max(v.Y, v.Z))
}

func MinVec(a, b Vec3) Vec3 {
	return Vec3{math.Min(a.X, b.X), math.Min(a.Y, b.Y), math.Min(a.Z, b.Z)}
}

func MaxVec(a, b Vec3) Vec3 {
	return Vec3{math.Max(a.X, b.X), math.Max(a.Y, b.Y), math.Max(a.Z, b.Z)}
}

func (v Vec3) Mul(s float64) Vec3 {
	return Vec3{v.X * s, v.Y * s, v.Z * s}
}

func (v Vec3) Cross(u Vec3) Vec3 {
	return Vec3{v.Y*u.Z - v.Z*u.Y, v.Z*u.X - v.X*u.Z, v.X*u.Y - v.Y*u.X}
}

func (v Vec3) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vec3) Comp(axis int) float64 {
	switch axis {
	case 0:
		return v.X
	case 1:
		return v.Y
	default:
		return v.Z
	}
}

func (v Vec3) RotY(angle float64) Vec3 {
	s, c := math.Sin(angle), math.Cos(angle)
	return Vec3{v.X*c + v.Z*s, v.Y, -v.X*s + v.Z*c}
}

func (v Vec3) RotX(angle float64) Vec3 {
	s, c := math.Sin(angle), math.Cos(angle)
	return Vec3{v.X, v.Y*c - v.Z*s, v.Y*s + v.Z*c}
}
