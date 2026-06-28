package geo

type Vec2 struct {
	U, V float64
}

func NewVec2(u, v float64) Vec2 {
	return Vec2{u, v}
}

func (v Vec2) Add(o Vec2) Vec2 {
	return Vec2{v.U + o.U, v.V + o.V}
}

func (v Vec2) Mul(s float64) Vec2 {
	return Vec2{v.U * s, v.V * s}
}
