package model

import (
	"math"

	"term-render/internal/color"
	"term-render/internal/geo"
)

type Hit struct {
	t      float64
	Point  geo.Vec3
	Normal geo.Vec3
	Color  color.Color
}

func (m *Model) Intersect(ro, rd geo.Vec3) (Hit, bool) {
	return m.root.intersect(ro, rd, m.textures)
}

func (n *bvhNode) intersect(ro, rd geo.Vec3, textures []texture) (Hit, bool) {
	tMin, tMax := aabbHit(ro, rd, n.min, n.max)
	if tMax < tMin || tMax < 0 {
		return Hit{}, false
	}

	best := Hit{t: math.MaxFloat64}
	hit := false

	if n.left == nil && n.right == nil {
		for _, tri := range n.triangles {
			if h, ok := tri.intersect(ro, rd, textures); ok && h.t < best.t {
				best = h
				hit = true
			}
		}
	} else {
		for _, child := range []*bvhNode{n.left, n.right} {
			if child == nil {
				continue
			}
			if h, ok := child.intersect(ro, rd, textures); ok && h.t < best.t {
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

func (t *triangle) intersect(ro, rd geo.Vec3, textures []texture) (Hit, bool) {
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

	colorR := t.c0.R*w + t.c1.R*u + t.c2.R*v
	colorG := t.c0.G*w + t.c1.G*u + t.c2.G*v
	colorB := t.c0.B*w + t.c1.B*u + t.c2.B*v
	c := color.New(colorR, colorG, colorB)

	if t.texIdx >= 0 && t.texIdx < len(textures) {
		uvU := t.t0.U*w + t.t1.U*u + t.t2.U*v
		uvV := t.t0.V*w + t.t1.V*u + t.t2.V*v
		tc := sampleTexture(textures[t.texIdx], uvU, uvV)
		c = tc.Mul(c)
	}

	return Hit{t: dist, Point: point, Normal: normal, Color: c}, true
}

func sampleTexture(tex texture, u, v float64) color.Color {
	u = math.Mod(u, 1.0)
	if u < 0 {
		u += 1.0
	}
	v = math.Mod(v, 1.0)
	if v < 0 {
		v += 1.0
	}

	px := int(u*float64(tex.Width)) % tex.Width
	py := int(v*float64(tex.Height)) % tex.Height
	offset := (py*tex.Width + px) * 4

	return color.New(
		float64(tex.Image.Pix[offset])/255.0,
		float64(tex.Image.Pix[offset+1])/255.0,
		float64(tex.Image.Pix[offset+2])/255.0,
	)
}
