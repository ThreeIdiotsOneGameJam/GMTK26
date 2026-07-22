package ui

type Element interface {
	Update(delta float32)
	Draw()
}
