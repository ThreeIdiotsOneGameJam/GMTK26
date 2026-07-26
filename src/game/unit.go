package game

//go:generate stringer -type=UnitType -trimprefix=Unit

type UnitType uint8

const (
	UnitUnknown UnitType = iota
	UnitPeasant
	UnitArcher
	UnitKnight
	UnitScout
)
