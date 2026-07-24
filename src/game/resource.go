package game

type ResourceType uint8

const (
	UnknownResource ResourceType = iota
	WoodResource
	StoneResource
	CoalResource
	IronResource
	SteelResource
	GoldResource
)
