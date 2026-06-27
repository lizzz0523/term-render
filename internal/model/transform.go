package model

import "term-render/internal/geo"

func (m *Model) Scale(s float64) {
	m.root.scale(s)
	m.radius *= s
}

func (m *Model) RotateX(angle float64) {
	m.root.rotateX(angle)
}

func (m *Model) RotateY(angle float64) {
	m.root.rotateY(angle)
}

func (m *Model) RotateZ(angle float64) {
	m.root.rotateZ(angle)
}

func (n *bvhNode) scale(s float64) {
	if n == nil {
		return
	}
	n.min = n.min.Mul(s)
	n.max = n.max.Mul(s)
	for i := range n.triangles {
		n.triangles[i].scale(s)
	}
	n.left.scale(s)
	n.right.scale(s)
}

func (t *triangle) scale(s float64) {
	t.v0 = t.v0.Mul(s)
	t.v1 = t.v1.Mul(s)
	t.v2 = t.v2.Mul(s)
}

func (n *bvhNode) rotateX(angle float64) {
	if n == nil {
		return
	}
	for i := range n.triangles {
		n.triangles[i].rotateX(angle)
	}
	n.left.rotateX(angle)
	n.right.rotateX(angle)
	n.recalcAABB()
}

func (n *bvhNode) rotateY(angle float64) {
	if n == nil {
		return
	}
	for i := range n.triangles {
		n.triangles[i].rotateY(angle)
	}
	n.left.rotateY(angle)
	n.right.rotateY(angle)
	n.recalcAABB()
}

func (n *bvhNode) rotateZ(angle float64) {
	if n == nil {
		return
	}
	for i := range n.triangles {
		n.triangles[i].rotateZ(angle)
	}
	n.left.rotateZ(angle)
	n.right.rotateZ(angle)
	n.recalcAABB()
}

func (n *bvhNode) recalcAABB() {
	if len(n.triangles) > 0 {
		n.min = n.triangles[0].v0
		n.max = n.triangles[0].v0
		for _, t := range n.triangles {
			for _, v := range []geo.Vec3{t.v0, t.v1, t.v2} {
				n.min = geo.MinVec(n.min, v)
				n.max = geo.MaxVec(n.max, v)
			}
		}
	} else {
		n.min = geo.MinVec(n.left.min, n.right.min)
		n.max = geo.MaxVec(n.left.max, n.right.max)
	}
}

func (t *triangle) rotateX(angle float64) {
	t.v0 = t.v0.RotX(angle)
	t.v1 = t.v1.RotX(angle)
	t.v2 = t.v2.RotX(angle)
	t.n0 = t.n0.RotX(angle)
	t.n1 = t.n1.RotX(angle)
	t.n2 = t.n2.RotX(angle)
}

func (t *triangle) rotateY(angle float64) {
	t.v0 = t.v0.RotY(angle)
	t.v1 = t.v1.RotY(angle)
	t.v2 = t.v2.RotY(angle)
	t.n0 = t.n0.RotY(angle)
	t.n1 = t.n1.RotY(angle)
	t.n2 = t.n2.RotY(angle)
}

func (t *triangle) rotateZ(angle float64) {
	t.v0 = t.v0.RotZ(angle)
	t.v1 = t.v1.RotZ(angle)
	t.v2 = t.v2.RotZ(angle)
	t.n0 = t.n0.RotZ(angle)
	t.n1 = t.n1.RotZ(angle)
	t.n2 = t.n2.RotZ(angle)
}
