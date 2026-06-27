package legacy

import (
	"math"

	"test-term/internal/geo"
)

const (
	bound = 1.2
	eps   = 0.01
)

var camera = geo.NewVec3(0, 0, -4.0)

func box(t, px, py float64) (n, hp geo.Vec3, ok bool) {
	ro := camera
	rd := geo.NewVec3(px, py, 1).Norm()

	angleX := 0.5 + 0.3*math.Sin(t*0.6)
	angleY := t * 0.7

	rd = rd.RotY(-angleY).RotX(-angleX)
	ro = ro.RotY(-angleY).RotX(-angleX)

	nBound := geo.NewVec3(-bound, -bound, -bound)
	pBound := geo.NewVec3(bound, bound, bound)

	t1 := nBound.Sub(ro).Div(rd)
	t2 := pBound.Sub(ro).Div(rd)

	tmin := geo.MaxVec(geo.MinVec(t1, t2), geo.NewVec3(-1e9, -1e9, -1e9)).MaxComp()
	tmax := geo.MinVec(geo.MaxVec(t1, t2), geo.NewVec3(1e9, 1e9, 1e9)).MinComp()

	if tmax < tmin || tmax < 0 {
		return geo.Vec3{}, geo.Vec3{}, false
	}

	d := tmin
	if d < 0 {
		d = tmax
	}

	hp = ro.Add(rd.Mul(d))

	switch {
	case hp.X > bound-eps:
		n = geo.NewVec3(1, 0, 0)
	case hp.X < -bound+eps:
		n = geo.NewVec3(-1, 0, 0)
	case hp.Y > bound-eps:
		n = geo.NewVec3(0, 1, 0)
	case hp.Y < -bound+eps:
		n = geo.NewVec3(0, -1, 0)
	case hp.Z > bound-eps:
		n = geo.NewVec3(0, 0, 1)
	case hp.Z < -bound+eps:
		n = geo.NewVec3(0, 0, -1)
	default:
		return geo.Vec3{}, geo.Vec3{}, false
	}
	n = n.RotX(angleX).RotY(angleY)

	return n, hp, true
}

func doughnut(t, px, py float64) (n, hp geo.Vec3, ok bool) {
	ro := camera
	rd := geo.NewVec3(px, py, 1).Norm()

	angleX := 0.5 + 0.3*math.Sin(t*0.6)
	angleY := t * 0.7

	rd = rd.RotY(-angleY).RotX(-angleX)
	ro = ro.RotY(-angleY).RotX(-angleX)

	R := 1.3
	r := 0.5
	RR := R * R
	rr := r * r

	A := rd.Dot(rd)
	B := 2 * ro.Dot(rd)
	C := ro.Dot(ro)

	Axy := rd.X*rd.X + rd.Y*rd.Y
	Bxy := 2 * (ro.X*rd.X + ro.Y*rd.Y)
	Cxy := ro.X*ro.X + ro.Y*ro.Y

	K := C + RR - rr

	a := A * A
	b := 2 * A * B
	c := B*B + 2*A*K - 4*RR*Axy
	d := 2*B*K - 4*RR*Bxy
	e := K*K - 4*RR*Cxy

	roots := solveQuartic(a, b, c, d, e)
	if len(roots) == 0 {
		return geo.Vec3{}, geo.Vec3{}, false
	}

	dist := roots[0]
	hp = ro.Add(rd.Mul(dist))

	dxy := hp.X*hp.X + hp.Y*hp.Y
	if dxy < 1e-10 {
		return geo.Vec3{}, geo.Vec3{}, false
	}
	sdxy := math.Sqrt(dxy)
	factor := 1 - R/sdxy

	n = geo.NewVec3(hp.X*factor, hp.Y*factor, hp.Z).Norm()
	n = n.RotX(angleX).RotY(angleY)

	return n, hp, true
}

func solveQuartic(a, b, c, d, e float64) []float64 {
	if math.Abs(a) < 1e-15 {
		return nil
	}
	p, q, r, s := b/a, c/a, d/a, e/a

	p2 := p * p
	p3 := p2 * p
	p4 := p2 * p2

	alpha := q - 3*p2/8
	beta := r - p*q/2 + p3/8
	gamma := s - p*r/4 + p2*q/16 - 3*p4/256

	roots := solveDepressedQuartic(alpha, beta, gamma)

	shift := p / 4
	for i := range roots {
		roots[i] -= shift
	}

	filtered := make([]float64, 0, len(roots))
	for _, r := range roots {
		if r > 1e-10 {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func solveCubicReal(a, b, c float64) float64 {
	a2 := a * a
	p := b - a2/3
	q := 2*a2*a/27 - a*b/3 + c

	disc := q*q/4 + p*p*p/27

	if disc >= 0 {
		sd := math.Sqrt(disc)
		u := math.Cbrt(-q/2 + sd)
		v := math.Cbrt(-q/2 - sd)
		return u + v - a/3
	}

	sqrtNegP_3 := math.Sqrt(-p / 3)
	val := -q / (2 * sqrtNegP_3 * sqrtNegP_3 * sqrtNegP_3)
	if val > 1 {
		val = 1
	}
	if val < -1 {
		val = -1
	}
	phi := math.Acos(val)
	z := 2 * sqrtNegP_3 * math.Cos(phi/3)
	return z - a/3
}

func solveDepressedQuartic(alpha, beta, gamma float64) []float64 {
	a := -alpha / 2
	b := -gamma
	c := alpha*gamma/2 - beta*beta/8

	y := solveCubicReal(a, b, c)

	R := math.Sqrt(2*y - alpha)

	var roots []float64
	if math.Abs(R) > 1e-10 {
		Q1 := y - beta/(2*R)
		Q2 := y + beta/(2*R)

		disc1 := R*R - 4*Q1
		if disc1 >= 0 {
			sd := math.Sqrt(disc1)
			roots = append(roots, (-R-sd)/2, (-R+sd)/2)
		}
		disc2 := R*R - 4*Q2
		if disc2 >= 0 {
			sd := math.Sqrt(disc2)
			roots = append(roots, (R-sd)/2, (R+sd)/2)
		}
	} else if math.Abs(beta) < 1e-10 {
		disc := alpha*alpha - 4*gamma
		if disc >= 0 {
			sd := math.Sqrt(disc)
			for _, u := range []float64{(-alpha - sd) / 2, (-alpha + sd) / 2} {
				if u >= 0 {
					su := math.Sqrt(u)
					roots = append(roots, -su, su)
				}
			}
		}
	}
	return roots
}
