package model

func (m *Model) Scale(s float64) {
	m.root.scale(s)
	m.radius *= s
}

func (m *Model) RotateY(angle float64) {
	m.root.rotateY(angle)
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

func (n *bvhNode) rotateY(angle float64) {
	if n == nil {
		return
	}
	n.min = n.min.RotY(angle)
	n.max = n.max.RotY(angle)
	for i := range n.triangles {
		n.triangles[i].rotateY(angle)
	}
	n.left.rotateY(angle)
	n.right.rotateY(angle)
}

func (t *triangle) rotateY(angle float64) {
	t.v0 = t.v0.RotY(angle)
	t.v1 = t.v1.RotY(angle)
	t.v2 = t.v2.RotY(angle)
	t.n0 = t.n0.RotY(angle)
	t.n1 = t.n1.RotY(angle)
	t.n2 = t.n2.RotY(angle)
}
