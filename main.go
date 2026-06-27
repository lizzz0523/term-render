package main

import (
	"time"

	mdl "test-term/internal/model"
	"test-term/internal/renderer"

	"github.com/gdamore/tcell/v2"
)

func main() {
	model, err := mdl.LoadGLB("models/table_medium.glb")
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
			s.Clear()
			r.Render(s, t, model)
			s.Show()

		case ev := <-eventCh:
			switch ev.(type) {
			case *tcell.EventKey:
				return
			}
		}
	}
}
