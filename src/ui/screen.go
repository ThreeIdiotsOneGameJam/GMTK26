package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Screen struct {
	BackgroundColor color.RGBA
	Elements        []Element
}

func (s *Screen) Update(delta float32) {
	for _, el := range s.Elements {
		el.update(delta)
	}
}

func (s *Screen) Draw() {
	rl.ClearBackground(s.BackgroundColor)

	for _, el := range s.Elements {
		el.draw()
	}
}

func (s *Screen) AddElement(el Element) {
	s.Elements = append(s.Elements, el)
}
