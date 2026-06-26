package main

import "math"

type Triangle struct {
	V0, V1, V2 vec3
	N0, N1, N2 vec3
}

type Hit struct {
	T      float64
	Point  vec3
	Normal vec3
}

type BVHNode struct {
	Min, Max  vec3
	Left      *BVHNode
	Right     *BVHNode
	Triangles []Triangle
}

func buildBVH(tris []Triangle) *BVHNode {
	if len(tris) == 0 {
		return nil
	}
	node := &BVHNode{}
	node.Min = tris[0].V0
	node.Max = tris[0].V0
	for _, t := range tris {
		for _, v := range []vec3{t.V0, t.V1, t.V2} {
			node.Min = minVec(node.Min, v)
			node.Max = maxVec(node.Max, v)
		}
	}

	if len(tris) <= 4 {
		node.Triangles = tris
		return node
	}

	size := node.Max.sub(node.Min)
	var axis int
	if size.x >= size.y && size.x >= size.z {
		axis = 0
	} else if size.y >= size.x && size.y >= size.z {
		axis = 1
	} else {
		axis = 2
	}

	mid := 0.5 * (getComp(node.Min, axis) + getComp(node.Max, axis))

	leftTris := make([]Triangle, 0, len(tris)/2)
	rightTris := make([]Triangle, 0, len(tris)/2)

	for _, t := range tris {
		center := t.V0.add(t.V1).add(t.V2).mul(1.0 / 3)
		if getComp(center, axis) < mid {
			leftTris = append(leftTris, t)
		} else {
			rightTris = append(rightTris, t)
		}
	}

	if len(leftTris) == 0 || len(rightTris) == 0 {
		node.Triangles = tris
		return node
	}

	node.Left = buildBVH(leftTris)
	node.Right = buildBVH(rightTris)
	return node
}

func getComp(v vec3, axis int) float64 {
	switch axis {
	case 0:
		return v.x
	case 1:
		return v.y
	default:
		return v.z
	}
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
