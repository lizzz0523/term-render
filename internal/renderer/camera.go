package renderer

import "term-render/internal/geo"

type Camera struct {
	Pos        geo.Vec3
	Yaw, Pitch float64
}
