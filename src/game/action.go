package game

//go:generate stringer -type=ActionType -trimprefix=Action

type ActionType uint8

const (
	ActionUnknown ActionType = iota
	ActionPass
	ActionBuild
	ActionDispatch
)

type BuildActionPayload struct {
	Hex      Hex          `json:"hex"`
	Building BuildingType `json:"building"`
}

type DispatchActionPayload struct {
	Hex   Hex       `json:"hex"`
	To    Hex       `json:"to"`
	Troop TroopType `json:"troop"`
}
