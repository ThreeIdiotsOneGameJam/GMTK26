package util

type RuneSet map[rune]struct{}

func NewRuneSetFromString(s string) RuneSet {
	return NewRuneSet([]rune(s)...)
}

func NewRuneSet(runes ...rune) RuneSet {
	set := make(RuneSet, len(runes))
	for _, r := range runes {
		set[r] = struct{}{}
	}
	return set
}

func (s RuneSet) Contains(r rune) bool {
	_, ok := s[r]
	return ok
}

func (s RuneSet) Add(r rune) {
	s[r] = struct{}{}
}

func (s RuneSet) Remove(r rune) {
	delete(s, r)
}
