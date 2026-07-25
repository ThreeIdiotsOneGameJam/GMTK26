package game

//go:generate stringer -type=TileType -trimprefix=Tile

type TileType uint8

const (
	TileUnknown TileType = iota
	TileVoid
	TileWater
	TilePlains
	TileForest
	TileDesert
	TileJungle
	TileRock
	TileIron
	TileCoal
	TileGold
)
