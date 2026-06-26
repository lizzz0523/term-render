package main

import (
	"math"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	bound   = 1.2
	eps     = 0.01
	cameraZ = -4.0
)

var model *Model
var camera = vec3{0, 0, cameraZ}

func main() {
	var err error
	model, err = LoadGLB("models/table_medium.glb")
	if err != nil {
		panic(err)
	}
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

	var buf [][]float64
	prevW, prevH := -1, -1

	for {
		select {
		case <-ticker.C:
			t := time.Since(start).Seconds()
			w, h := s.Size()
			bw, bh := w*2, h*4

			if w != prevW || h != prevH {
				buf = make([][]float64, bh)
				for y := range buf {
					buf[y] = make([]float64, bw)
				}
				prevW, prevH = w, h
			}

			renderToBuffer(buf, bw, bh, t)
			floydSteinberg(buf, bw, bh)

			s.Clear()
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					s.PutStrStyled(x, y, string(brailleChar(buf, x, y)), tcell.StyleDefault)
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

func renderToBuffer(buf [][]float64, bw, bh int, t float64) {
	for y := 0; y < bh; y++ {
		for x := 0; x < bw; x++ {
			buf[y][x] = brightness(x, y, bw, bh, t)
		}
	}
}

func brightness(bx, by, bw, bh int, t float64) float64 {
	aspect := float64(bw) / float64(bh)
	px := (float64(bx)/float64(bw) - 0.5) * 2 * aspect
	py := (float64(by)/float64(bh) - 0.5) * -2

	n, hp, ok := intersect(t, px, py)
	if !ok {
		return 0
	}
	return shading(n, hp)
}

func intersect(t, px, py float64) (n, hp vec3, ok bool) {
	ro := camera
	rd := vec3{px, py, 1}.norm()

	angleX := 0.5 + 0.3*math.Sin(t*0.6)
	angleY := t * 0.7

	rd = rd.rotY(-angleY).rotX(-angleX)
	ro = ro.rotY(-angleY).rotX(-angleX)

	hit, ok := model.Root.intersect(ro, rd)
	if !ok {
		return vec3{}, vec3{}, false
	}
	return hit.Normal, hit.Point, true
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
