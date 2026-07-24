package game

type BuildingType uint8

const (
	BuildingUnknown BuildingType = iota
	BuildingForester
	BuildingMine
	BuildingBarracks
	BuildingFarm
)
