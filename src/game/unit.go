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

type UnitStats struct {
	MaxHP  int8
	Attack int8
}

var unitStatsTable = [5]UnitStats{
	UnitUnknown: {0, 0},
	UnitPeasant: {3, 1},
	UnitArcher:  {3, 2},
	UnitKnight:  {5, 3},
	UnitScout:   {3, 0},
}

func GetUnitStats(t UnitType) UnitStats {
	if int(t) >= len(unitStatsTable) {
		return UnitStats{}
	}
	return unitStatsTable[t]
}
