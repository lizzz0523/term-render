package main

import (
	"math"
	"time"

	"term-render/internal/game"
	mdl "term-render/internal/model"
	"term-render/internal/renderer"

	"github.com/gdamore/tcell/v2"
)

func main() {
	gun, err := mdl.LoadGLB("./models/mac10.glb", 5.0)
	if err != nil {
		panic(err)
	}
	gun.RotateY(-math.Pi / 2)
	gun.Scale(0.12)
	enemy := mdl.NewCube(0.8, 1.8, 0.8)

	s, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := s.Init(); err != nil {
		panic(err)
	}
	defer s.Fini()

	g := game.NewGame(gun, enemy)
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
			g.GameTime = time.Since(start).Seconds()
			s.Clear()
			r.Render(s, renderer.Camera{Pos: g.Player.Pos, Yaw: g.Player.Angle}, g)
			w, h := s.Size()
			s.PutStrStyled(w/2, h/2, "+", tcell.StyleDefault.Foreground(tcell.ColorRed))
			s.Show()

		case ev := <-eventCh:
			if !g.HandleInput(ev) {
				return
			}
		}
	}
}
