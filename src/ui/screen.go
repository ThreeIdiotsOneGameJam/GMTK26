package ui

type Screen struct {
	Elements []Element
}

func (s *Screen) Update(delta float32) {
	for _, el := range s.Elements {
		el.update(delta)
	}
}

func (s *Screen) Draw() {
	for _, el := range s.Elements {
		el.draw()
	}
}

func (s *Screen) AddElement(el Element) {
	s.Elements = append(s.Elements, el)
}
