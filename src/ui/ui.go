package ui

type Element interface {
	update(deltaNano int64)
	draw()
}
