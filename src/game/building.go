package game

//go:generate stringer -type=BuildingType -trimprefix=Building

type BuildingType uint8

const (
	BuildingUnknown BuildingType = iota
	BuildingForester
	BuildingMine
	BuildingBarracks
	BuildingFarm
)
