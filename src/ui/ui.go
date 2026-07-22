package ui

type Element interface {
	update(delta float32)
	draw()
}
