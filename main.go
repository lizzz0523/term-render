package main

import (
	"math"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	bound   = 1.2
	eps     = 0.01
	chars   = " .:;+=*#@"
	cameraZ = -4.0
)

var camera = vec3{0, 0, cameraZ}

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := s.Init(); err != nil {
		panic(err)
	}
	defer s.Fini()

	start := time.Now()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	eventCh := make(chan tcell.Event)
	go func() {
		for {
			eventCh <- s.PollEvent()
		}
	}()

	for {
		select {
		case <-ticker.C:
			t := time.Since(start).Seconds()
			w, h := s.Size()
			s.Clear()
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					if r := fragment(x, y, w, h, t); r != ' ' {
						s.PutStrStyled(x, y, string(r), tcell.StyleDefault)
					}
				}
			}
			s.Show()
		case ev := <-eventCh:
			switch ev.(type) {
			case *tcell.EventKey:
				return
			}
		}
	}
}

func fragment(x, y, width, height int, t float64) rune {
	aspect := float64(width) / float64(height) * 0.5
	px := (float64(x)/float64(width) - 0.5) * 2 * aspect
	py := (float64(y)/float64(height) - 0.5) * -2

	n, hp, ok := doughnut(t, px, py)
	if !ok {
		return ' '
	}

	idx := int(shading(n, hp) * float64(len(chars)-1))
	return rune(chars[idx])
}

func shading(n, hp vec3) float64 {
	light := vec3{0.3, 0.5, -0.8}.norm()
	ambient := 0.25

	brightness := ambient + (1-ambient)*math.Max(0, n.dot(light))
	if brightness > 1 {
		brightness = 1
	}

	viewDir := camera.sub(hp).norm()
	rim := 1 - math.Abs(n.dot(viewDir))
	if rim > 0.65 {
		brightness = math.Min(1, brightness+0.35)
	}

	return brightness
}

func box(t, px, py float64) (n, hp vec3, ok bool) {
	ro := camera
	rd := vec3{px, py, 1}.norm()

	angleX := 0.5 + 0.3*math.Sin(t*0.6)
	angleY := t * 0.7

	rd = rd.rotY(-angleY).rotX(-angleX)
	ro = ro.rotY(-angleY).rotX(-angleX)

	nBound := vec3{-bound, -bound, -bound}
	pBound := vec3{bound, bound, bound}

	t1 := nBound.sub(ro).div(rd)
	t2 := pBound.sub(ro).div(rd)

	tmin := maxVec(minVec(t1, t2), vec3{-1e9, -1e9, -1e9}).maxComp()
	tmax := minVec(maxVec(t1, t2), vec3{1e9, 1e9, 1e9}).minComp()

	if tmax < tmin || tmax < 0 {
		return vec3{}, vec3{}, false
	}

	d := tmin
	if d < 0 {
		d = tmax
	}

	hp = ro.add(rd.mul(d))

	switch {
	case hp.x > bound-eps:
		n = vec3{1, 0, 0}
	case hp.x < -bound+eps:
		n = vec3{-1, 0, 0}
	case hp.y > bound-eps:
		n = vec3{0, 1, 0}
	case hp.y < -bound+eps:
		n = vec3{0, -1, 0}
	case hp.z > bound-eps:
		n = vec3{0, 0, 1}
	case hp.z < -bound+eps:
		n = vec3{0, 0, -1}
	default:
		return vec3{}, vec3{}, false
	}
	n = n.rotX(angleX).rotY(angleY)

	return n, hp, true
}

func doughnut(t, px, py float64) (n, hp vec3, ok bool) {
	ro := camera
	rd := vec3{px, py, 1}.norm()

	angleX := 0.5 + 0.3*math.Sin(t*0.6)
	angleY := t * 0.7

	rd = rd.rotY(-angleY).rotX(-angleX)
	ro = ro.rotY(-angleY).rotX(-angleX)

	R := 1.3
	r := 0.5
	RR := R * R
	rr := r * r

	A := rd.dot(rd)
	B := 2 * ro.dot(rd)
	C := ro.dot(ro)

	Axy := rd.x*rd.x + rd.y*rd.y
	Bxy := 2 * (ro.x*rd.x + ro.y*rd.y)
	Cxy := ro.x*ro.x + ro.y*ro.y

	K := C + RR - rr

	a := A * A
	b := 2 * A * B
	c := B*B + 2*A*K - 4*RR*Axy
	d := 2*B*K - 4*RR*Bxy
	e := K*K - 4*RR*Cxy

	roots := solveQuartic(a, b, c, d, e)
	if len(roots) == 0 {
		return vec3{}, vec3{}, false
	}

	dist := roots[0]
	hp = ro.add(rd.mul(dist))

	dxy := hp.x*hp.x + hp.y*hp.y
	if dxy < 1e-10 {
		return vec3{}, vec3{}, false
	}
	sdxy := math.Sqrt(dxy)
	factor := 1 - R/sdxy

	n = vec3{hp.x * factor, hp.y * factor, hp.z}.norm()
	n = n.rotX(angleX).rotY(angleY)

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
