package game

//go:generate stringer -type=TroopType -trimprefix=Troop

type TroopType uint8

const (
	TroopUnknown TroopType = iota
	TroopPeasant
	TroopArcher
	TroopKnight
	TroopScout
)
