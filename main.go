package main

import (
	"fmt"
	"os"
	"time"

	"term-render/internal/geo"
	mdl "term-render/internal/model"
	"term-render/internal/renderer"

	"github.com/gdamore/tcell/v2"
)

type ViewScene struct {
	*mdl.Model
}

func (v *ViewScene) Intersect(ro, rd geo.Vec3) (renderer.Hit, bool) {
	hit, ok := v.Model.Intersect(ro, rd)
	if !ok {
		return renderer.Hit{}, false
	}
	return renderer.Hit{Point: hit.Point, Normal: hit.Normal}, true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <model.glb>\n", os.Args[0])
		os.Exit(1)
	}
	model, err := mdl.LoadGLB(os.Args[1])
	if err != nil {
		panic(err)
	}
	scene := &ViewScene{Model: model}
	s, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := s.Init(); err != nil {
		panic(err)
	}
	defer s.Fini()

	r := renderer.New()
	prev := time.Now()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	eventCh := make(chan tcell.Event)
	go func() {
		for {
			eventCh <- s.PollEvent()
		}
	}()

	camera := renderer.Camera{Pos: geo.NewVec3(0, 0, -7.0), Yaw: 0, Pitch: 0}

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			dt := now.Sub(prev).Seconds()
			prev = now

			model.RotateY(0.7 * dt)
			s.Clear()
			r.Render(s, camera, scene)
			s.Show()

		case ev := <-eventCh:
			switch ev.(type) {
			case *tcell.EventKey:
				return
			}
		}
	}
}
