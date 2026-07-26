package game

import "time"

const (
	MatchDuration       = 15 * time.Minute
	RoundDuration       = 5 * time.Second
	TotalMatchRounds    = int32(MatchDuration / RoundDuration)
	ScoreIntervalRounds = int32(6)
	StartingCoins       = int32(25)

	DefaultMapWidth  = int32(72)
	DefaultMapHeight = int32(72)
)
