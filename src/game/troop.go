package game

type TroopType uint8

const (
	UnknownTroopType TroopType = iota
	PeasantTroopType
	ArcherTroopType
	KnightTroopType
)
