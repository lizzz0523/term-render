package color

import "math"

type Color struct {
	R, G, B float64
}

func New(r, g, b float64) Color {
	return Color{r, g, b}
}

func (c Color) Mul(other Color) Color {
	return New(
		clamp(c.R*other.R),
		clamp(c.G*other.G),
		clamp(c.B*other.B),
	)
}

func clamp(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
