package renderer

import (
	"math"
	"runtime"
	"sync"

	"term-render/internal/geo"

	"github.com/gdamore/tcell/v2"
)

type Renderer struct {
	buf          [][]float64
	prevW, prevH int
}

func New() *Renderer {
	return &Renderer{}
}

func (r *Renderer) Render(s tcell.Screen, camera Camera, scene Scene) {
	w, h := s.Size()
	bw, bh := w*2, h*4

	if w != r.prevW || h != r.prevH {
		r.buf = make([][]float64, bh)
		for y := range r.buf {
			r.buf[y] = make([]float64, bw)
		}
		r.prevW, r.prevH = w, h
	}

	renderToBuffer(r.buf, bw, bh, camera, scene)
	floydSteinberg(r.buf, bw, bh)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s.PutStrStyled(x, y, string(brailleChar(r.buf, x, y)), tcell.StyleDefault)
		}
	}
}

func renderToBuffer(buf [][]float64, bw, bh int, camera Camera, scene Scene) {
	numCPU := runtime.GOMAXPROCS(0)
	if bh < numCPU*2 {
		renderRowRange(buf, bw, bh, camera, scene, 0, bh)
		return
	}

	rowsPer := (bh + numCPU - 1) / numCPU
	var wg sync.WaitGroup

	for i := 0; i < numCPU; i++ {
		y0 := i * rowsPer
		if y0 >= bh {
			break
		}
		y1 := y0 + rowsPer
		if y1 > bh {
			y1 = bh
		}
		wg.Add(1)
		go func(y0, y1 int) {
			defer wg.Done()
			renderRowRange(buf, bw, bh, camera, scene, y0, y1)
		}(y0, y1)
	}
	wg.Wait()
}

func renderRowRange(buf [][]float64, bw, bh int, camera Camera, scene Scene, y0, y1 int) {
	for y := y0; y < y1; y++ {
		for x := 0; x < bw; x++ {
			buf[y][x] = brightness(x, y, bw, bh, camera, scene)
		}
	}
}

func brightness(bx, by, bw, bh int, camera Camera, scene Scene) float64 {
	aspect := float64(bw) / float64(bh)
	px := (float64(bx)/float64(bw) - 0.5) * 2 * aspect
	py := (float64(by)/float64(bh) - 0.5) * -2

	n, hp, vp, ok := raycast(camera, px, py, scene)
	if !ok {
		return 0
	}
	return shading(n, hp, vp)
}

func raycast(camera Camera, px, py float64, scene Scene) (n, hp, vp geo.Vec3, ok bool) {
	ro := camera.Pos
	rd := geo.NewVec3(px, py, 1).Norm()
	rd = rd.RotY(-camera.Yaw).RotX(-camera.Pitch)

	hit, ok := scene.Intersect(ro, rd)
	if !ok {
		return geo.Vec3{}, geo.Vec3{}, geo.Vec3{}, false
	}
	normal := hit.Normal.RotX(camera.Pitch).RotY(camera.Yaw)

	return normal, hit.Point, ro, true
}

func shading(n, hp, vp geo.Vec3) float64 {
	light := geo.NewVec3(0.3, 0.5, -0.8).Norm()
	ambient := 0.25

	brightness := ambient + (1-ambient)*math.Max(0, n.Dot(light))
	if brightness > 1 {
		brightness = 1
	}

	viewDir := vp.Sub(hp).Norm()
	rim := 1 - math.Abs(n.Dot(viewDir))
	if rim > 0.65 {
		brightness = math.Min(1, brightness+0.35)
	}

	return brightness
}

func floydSteinberg(buf [][]float64, bw, bh int) {
	for y := 0; y < bh; y++ {
		for x := 0; x < bw; x++ {
			old := buf[y][x]
			binary := 0.0
			if old >= 0.5 {
				binary = 1.0
			}
			buf[y][x] = binary
			err := old - binary

			if x+1 < bw {
				buf[y][x+1] += err * 7 / 16
			}
			if y+1 < bh {
				if x > 0 {
					buf[y+1][x-1] += err * 3 / 16
				}
				buf[y+1][x] += err * 5 / 16
				if x+1 < bw {
					buf[y+1][x+1] += err * 1 / 16
				}
			}
		}
	}
}

func brailleChar(buf [][]float64, sx, sy int) rune {
	var code uint16
	for dy := 0; dy < 4; dy++ {
		for dx := 0; dx < 2; dx++ {
			bx := sx*2 + dx
			by := sy*4 + dy
			if buf[by][bx] >= 0.5 {
				var bit uint16
				switch {
				case dx == 0 && dy == 0:
					bit = 0x01
				case dx == 1 && dy == 0:
					bit = 0x08
				case dx == 0 && dy == 1:
					bit = 0x02
				case dx == 1 && dy == 1:
					bit = 0x10
				case dx == 0 && dy == 2:
					bit = 0x04
				case dx == 1 && dy == 2:
					bit = 0x20
				case dx == 0 && dy == 3:
					bit = 0x40
				case dx == 1 && dy == 3:
					bit = 0x80
				}
				code |= bit
			}
		}
	}
	return rune(0x2800 + code)
}
