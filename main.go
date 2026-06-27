package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"term-render/internal/geo"
	mdl "term-render/internal/model"
	"term-render/internal/renderer"

	"github.com/gdamore/tcell/v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <model.glb>\n", os.Args[0])
		os.Exit(1)
	}
	model, err := mdl.LoadGLB(os.Args[1])
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

	r := renderer.New()
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
			yaw := t * 0.7
			pitch := 0.5 + 0.3*math.Sin(t*0.6)
			camPos := geo.NewVec3(0, 0, -7.0).RotY(-yaw).RotX(-pitch)
			camera := renderer.Camera{Pos: camPos, Yaw: yaw, Pitch: pitch}
			s.Clear()
			r.Render(s, camera, model)
			s.Show()

		case ev := <-eventCh:
			switch ev.(type) {
			case *tcell.EventKey:
				return
			}
		}
	}
}
