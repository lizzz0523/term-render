package main

import "math"

type vec3 struct {
	x, y, z float64
}

func (v vec3) norm() vec3 {
	d := math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
	return vec3{v.x / d, v.y / d, v.z / d}
}

func (v vec3) dot(u vec3) float64 {
	return v.x*u.x + v.y*u.y + v.z*u.z
}

func (v vec3) add(u vec3) vec3 {
	return vec3{v.x + u.x, v.y + u.y, v.z + u.z}
}

func (v vec3) sub(u vec3) vec3 {
	return vec3{v.x - u.x, v.y - u.y, v.z - u.z}
}

func (v vec3) div(u vec3) vec3 {
	return vec3{v.x / u.x, v.y / u.y, v.z / u.z}
}

func (v vec3) minComp() float64 {
	return math.Min(v.x, math.Min(v.y, v.z))
}

func (v vec3) maxComp() float64 {
	return math.Max(v.x, math.Max(v.y, v.z))
}

func minVec(a, b vec3) vec3 {
	return vec3{math.Min(a.x, b.x), math.Min(a.y, b.y), math.Min(a.z, b.z)}
}

func maxVec(a, b vec3) vec3 {
	return vec3{math.Max(a.x, b.x), math.Max(a.y, b.y), math.Max(a.z, b.z)}
}

func (v vec3) mul(s float64) vec3 {
	return vec3{v.x * s, v.y * s, v.z * s}
}

func (v vec3) cross(u vec3) vec3 {
	return vec3{v.y*u.z - v.z*u.y, v.z*u.x - v.x*u.z, v.x*u.y - v.y*u.x}
}

func (v vec3) len() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

func (v vec3) comp(axis int) float64 {
	switch axis {
	case 0:
		return v.x
	case 1:
		return v.y
	default:
		return v.z
	}
}

func (v vec3) rotY(angle float64) vec3 {
	s, c := math.Sin(angle), math.Cos(angle)
	return vec3{v.x*c + v.z*s, v.y, -v.x*s + v.z*c}
}

func (v vec3) rotX(angle float64) vec3 {
	s, c := math.Sin(angle), math.Cos(angle)
	return vec3{v.x, v.y*c - v.z*s, v.y*s + v.z*c}
}
