package model

import (
	"math"

	"term-render/internal/geo"
)

type Hit struct {
	t      float64
	Point  geo.Vec3
	Normal geo.Vec3
}

func (m *Model) Intersect(ro, rd geo.Vec3) (Hit, bool) {
	return m.root.intersect(ro, rd)
}

func (n *bvhNode) intersect(ro, rd geo.Vec3) (Hit, bool) {
	tMin, tMax := aabbHit(ro, rd, n.min, n.max)
	if tMax < tMin || tMax < 0 {
		return Hit{}, false
	}

	best := Hit{t: math.MaxFloat64}
	hit := false

	if n.left == nil && n.right == nil {
		for _, triangle := range n.triangles {
			if h, ok := triangle.intersect(ro, rd); ok && h.t < best.t {
				best = h
				hit = true
			}
		}
	} else {
		for _, child := range []*bvhNode{n.left, n.right} {
			if child == nil {
				continue
			}
			if h, ok := child.intersect(ro, rd); ok && h.t < best.t {
				best = h
				hit = true
			}
		}
	}

	return best, hit
}

func aabbHit(ro, rd, min, max geo.Vec3) (tMin, tMax float64) {
	t1 := (min.X - ro.X) / rd.X
	t2 := (max.X - ro.X) / rd.X
	tMin = math.Min(t1, t2)
	tMax = math.Max(t1, t2)

	t1 = (min.Y - ro.Y) / rd.Y
	t2 = (max.Y - ro.Y) / rd.Y
	tMin = math.Max(tMin, math.Min(t1, t2))
	tMax = math.Min(tMax, math.Max(t1, t2))

	t1 = (min.Z - ro.Z) / rd.Z
	t2 = (max.Z - ro.Z) / rd.Z
	tMin = math.Max(tMin, math.Min(t1, t2))
	tMax = math.Min(tMax, math.Max(t1, t2))

	return tMin, tMax
}

func (t *triangle) intersect(ro, rd geo.Vec3) (Hit, bool) {
	e1 := t.v1.Sub(t.v0)
	e2 := t.v2.Sub(t.v0)
	p := rd.Cross(e2)
	det := e1.Dot(p)

	if math.Abs(det) < 1e-10 {
		return Hit{}, false
	}
	invDet := 1.0 / det

	s := ro.Sub(t.v0)
	u := s.Dot(p) * invDet
	if u < 0 || u > 1 {
		return Hit{}, false
	}

	q := s.Cross(e1)
	v := rd.Dot(q) * invDet
	if v < 0 || u+v > 1 {
		return Hit{}, false
	}

	dist := e2.Dot(q) * invDet
	if dist < 1e-10 {
		return Hit{}, false
	}

	point := ro.Add(rd.Mul(dist))
	w := 1 - u - v
	normal := t.n0.Mul(w).Add(t.n1.Mul(u)).Add(t.n2.Mul(v)).Norm()

	return Hit{t: dist, Point: point, Normal: normal}, true
}
