package main

import (
	"github.com/gdamore/tcell/v2"
)

func fragment(x, y, width, height int) rune {
	text := "hello world"
	cx := width/2 - len(text)/2
	cy := height / 2

	if y == cy && x >= cx && x < cx+len(text) {
		return rune(text[x-cx])
	}
	return ' '
}

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := s.Init(); err != nil {
		panic(err)
	}
	defer s.Fini()

	s.Clear()

	w, h := s.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if r := fragment(x, y, w, h); r != ' ' {
				s.PutStrStyled(x, y, string(r), tcell.StyleDefault)
			}
		}
	}

	s.Show()

	for {
		ev := s.PollEvent()
		switch ev.(type) {
		case *tcell.EventKey:
			return
		}
	}
}
