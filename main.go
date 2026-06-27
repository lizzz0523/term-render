package main

import (
	"fmt"
	"os"
	"time"

	mdl "test-term/internal/model"
	"test-term/internal/renderer"

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
