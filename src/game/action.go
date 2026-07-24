package game

type ActionType uint8

const (
	ActionUnknown ActionType = iota
	ActionPass
	ActionClaim
	ActionBuild
	ActionDispatch
)
