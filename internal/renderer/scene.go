package renderer

import "term-render/internal/geo"

type Hit struct {
	Point  geo.Vec3
	Normal geo.Vec3
}

type Scene interface {
	Intersect(ro, rd geo.Vec3) (Hit, bool)
}
