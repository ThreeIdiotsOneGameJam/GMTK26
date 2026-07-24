package game

type ResourceType uint8

const (
	ResourceUnknown ResourceType = iota
	ResourceWood
	ResourceStone
	ResourceCoal
	ResourceIron
	ResourceSteel
	ResourceGold
)
