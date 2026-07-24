package game

type TroopType uint8

const (
	TroopUnknown TroopType = iota
	TroopPeasant
	TroopArcher
	TroopKnight
)
