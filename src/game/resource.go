package game

//go:generate stringer -type=ResourceType -trimprefix=Resource

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

type Resources map[ResourceType]uint32
