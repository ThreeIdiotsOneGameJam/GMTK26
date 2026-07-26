package ai

import (
	"math"
	"sort"
)

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func ramp(value, low, high float64) float64 {
	if high <= low {
		if value >= high {
			return 1
		}
		return 0
	}
	return clamp01((value - low) / (high - low))
}

func stableMix(value uint64) uint64 {
	value += 0x9e3779b97f4a7c15
	value = (value ^ value>>30) * 0xbf58476d1ce4e5b9
	value = (value ^ value>>27) * 0x94d049bb133111eb
	return value ^ value>>31
}

func stableUnit(seed uint64) float64 {
	return float64(stableMix(seed)>>11) / float64(uint64(1)<<53)
}

func seededPersonality(seed int64, owner int8) Personality {
	base := uint64(seed) ^ uint64(uint8(owner)+1)*0x9e3779b97f4a7c15
	deviations := [6]float64{}
	var deviationMean float64
	for i := range deviations {
		deviations[i] = -0.15 + 0.30*stableUnit(base+uint64(i)*0x517cc1b727220a95)
		deviationMean += deviations[i]
	}
	deviationMean /= float64(len(deviations))
	maxDeviation := 0.0
	for i := range deviations {
		deviations[i] -= deviationMean
		maxDeviation = math.Max(maxDeviation, math.Abs(deviations[i]))
	}
	scale := 1.0
	if maxDeviation > 0.15 {
		scale = 0.15 / maxDeviation
	}
	values := [6]float64{}
	for i := range values {
		values[i] = 1 + deviations[i]*scale
	}
	return Personality{
		Economy:     values[0],
		Expansion:   values[1],
		Defense:     values[2],
		Aggression:  values[3],
		Risk:        values[4],
		Opportunism: values[5],
	}
}

type candidate struct {
	key         string
	description string
	manual      *ManualAction
	utility     float64
}

func chooseCandidate(
	candidates []candidate,
	seed uint64,
	band float64,
	temperature float64,
) (candidate, []ScoredAlternative) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].utility == candidates[j].utility {
			return candidates[i].key < candidates[j].key
		}
		return candidates[i].utility > candidates[j].utility
	})
	if len(candidates) == 0 {
		return candidate{description: "No legal action"}, nil
	}

	best := candidates[0].utility
	eligible := 1
	for eligible < len(candidates) && candidates[eligible].utility >= best-band {
		eligible++
	}

	chosen := 0
	if eligible > 1 && temperature > 0 {
		weights := make([]float64, eligible)
		var total float64
		for i := range weights {
			weights[i] = math.Exp((candidates[i].utility - best) / temperature)
			total += weights[i]
		}
		draw := stableUnit(seed) * total
		for i, weight := range weights {
			draw -= weight
			if draw <= 0 {
				chosen = i
				break
			}
		}
	}

	alternatives := make([]ScoredAlternative, 0, min(3, len(candidates)))
	for _, item := range candidates {
		if item.key == candidates[chosen].key {
			continue
		}
		alternatives = append(alternatives, ScoredAlternative{
			Description: item.description,
			Utility:     item.utility,
		})
		if len(alternatives) == 3 {
			break
		}
	}
	return candidates[chosen], alternatives
}
