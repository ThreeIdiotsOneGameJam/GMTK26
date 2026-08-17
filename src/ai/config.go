package ai

type Config struct {
	MaxCandidates          int
	MaxDetailedCandidates  int
	MaxPathQueries         int
	ForecastRounds         int32
	GoalCommitRounds       int32
	TargetCommitRounds     int32
	EmergencyThreat        float64
	GoalSwitchMargin       float64
	TargetSwitchMultiplier float64
	ChoiceBand             float64
	ChoiceTemperature      float64
}

func StandardConfig() Config {
	return Config{
		MaxCandidates:          128,
		MaxDetailedCandidates:  24,
		MaxPathQueries:         32,
		ForecastRounds:         6,
		GoalCommitRounds:       3,
		TargetCommitRounds:     4,
		EmergencyThreat:        0.70,
		GoalSwitchMargin:       0.12,
		TargetSwitchMultiplier: 1.25,
		ChoiceBand:             0.06,
		ChoiceTemperature:      0.025,
	}
}
