package game

type BuildingType uint8

const (
	UnknownBuildingType BuildingType = iota
	ForesterBuildingType
	MineBuildingType
	BarracksBuildingType
	FarmBuildingType
)
