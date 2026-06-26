package main

import "math"

type Hit struct {
	T      float64
	Point  vec3
	Normal vec3
}

func (node *BVHNode) intersect(ro, rd vec3) (Hit, bool) {
	tMin, tMax := aabbHit(ro, rd, node.Min, node.Max)
	if tMax < tMin || tMax < 0 {
		return Hit{}, false
	}

	best := Hit{T: math.MaxFloat64}
	hit := false

	if node.Left == nil && node.Right == nil {
		for _, t := range node.Triangles {
			if h, ok := intersectTriangle(ro, rd, t); ok && h.T < best.T {
				best = h
				hit = true
			}
		}
	} else {
		for _, child := range []*BVHNode{node.Left, node.Right} {
			if child == nil {
				continue
			}
			if h, ok := child.intersect(ro, rd); ok && h.T < best.T {
				best = h
				hit = true
			}
		}
	}

	return best, hit
}

func aabbHit(ro, rd, min, max vec3) (tMin, tMax float64) {
	t1 := (min.x - ro.x) / rd.x
	t2 := (max.x - ro.x) / rd.x
	tMin = math.Min(t1, t2)
	tMax = math.Max(t1, t2)

	t1 = (min.y - ro.y) / rd.y
	t2 = (max.y - ro.y) / rd.y
	tMin = math.Max(tMin, math.Min(t1, t2))
	tMax = math.Min(tMax, math.Max(t1, t2))

	t1 = (min.z - ro.z) / rd.z
	t2 = (max.z - ro.z) / rd.z
	tMin = math.Max(tMin, math.Min(t1, t2))
	tMax = math.Min(tMax, math.Max(t1, t2))

	return tMin, tMax
}

func intersectTriangle(ro, rd vec3, t Triangle) (Hit, bool) {
	e1 := t.V1.sub(t.V0)
	e2 := t.V2.sub(t.V0)
	p := rd.cross(e2)
	det := e1.dot(p)

	if math.Abs(det) < 1e-10 {
		return Hit{}, false
	}
	invDet := 1.0 / det

	s := ro.sub(t.V0)
	u := s.dot(p) * invDet
	if u < 0 || u > 1 {
		return Hit{}, false
	}

	q := s.cross(e1)
	v := rd.dot(q) * invDet
	if v < 0 || u+v > 1 {
		return Hit{}, false
	}

	dist := e2.dot(q) * invDet
	if dist < 1e-10 {
		return Hit{}, false
	}

	point := ro.add(rd.mul(dist))
	w := 1 - u - v
	normal := t.N0.mul(w).add(t.N1.mul(u)).add(t.N2.mul(v)).norm()

	return Hit{T: dist, Point: point, Normal: normal}, true
}
