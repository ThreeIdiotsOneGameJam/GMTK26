package ui

// Group is a non-rendering element for grouping children and running
// screen-local update logic such as animations.
func Group() *GroupElement {
	el := &GroupElement{}
	el.BaseElement = NewBaseElement(el)
	return el
}

func (el *GroupElement) WithUpdate(update func(deltaNano int64)) *GroupElement {
	el.Update = update
	return el
}

type GroupElement struct {
	BaseElement[*GroupElement]
	Update func(deltaNano int64)
}

func (el *GroupElement) update(deltaNano int64) {
	if el.Update != nil {
		el.Update(deltaNano)
	}
}

func (el *GroupElement) draw() {}
