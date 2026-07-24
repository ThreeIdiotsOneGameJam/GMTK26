package game

//go:generate stringer -type=ActionType -trimprefix=Action

type ActionType uint8

const (
	ActionUnknown ActionType = iota
	ActionPass
	ActionClaim
	ActionBuild
	ActionDispatch
)
