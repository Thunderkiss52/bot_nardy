package engine

import "math"

func PipCount(state GameState, player Color) int {
	total := 0
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner != player || pt.Count == 0 {
			continue
		}
		total += DistanceToBearOff(player, p) * pt.Count
	}
	if state.GameType == GameShort {
		total += state.Bar[player.Idx()] * 25
	}
	return total
}

func CountBlots(state GameState, player Color) int {
	count := 0
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner == player && pt.Count == 1 {
			count++
		}
	}
	return count
}

func CountMadePoints(state GameState, player Color) int {
	count := 0
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner == player && pt.Count >= 2 {
			count++
		}
	}
	return count
}

func LongestPrime(state GameState, player Color) int {
	best := 0
	run := 0
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner == player && pt.Count >= 2 {
			run++
			if run > best {
				best = run
			}
		} else {
			run = 0
		}
	}
	return best
}

func HomeMadePoints(state GameState, player Color) int {
	hFrom, hTo := HomeRange(player)
	count := 0
	for p := hFrom; p <= hTo; p++ {
		pt := state.Points[p]
		if pt.Owner == player && pt.Count >= 2 {
			count++
		}
	}
	return count
}

func BearoffReady(state GameState, player Color) float64 {
	if allInHome(state, player) {
		return 1
	}
	return 0
}

func ContactExists(state GameState) bool {
	whiteMin := 25
	blackMax := 0
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner == White && pt.Count > 0 && p < whiteMin {
			whiteMin = p
		}
		if pt.Owner == Black && pt.Count > 0 && p > blackMax {
			blackMax = p
		}
	}
	if whiteMin == 25 || blackMax == 0 {
		return false
	}
	return blackMax >= whiteMin
}

func CountBlocksLong(state GameState, player Color) int {
	return CountMadePoints(state, player)
}

func MaxBlockRunLong(state GameState, player Color) int {
	best := 0
	run := 0
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner == player && pt.Count >= 2 {
			run++
			if run > best {
				best = run
			}
		} else {
			run = 0
		}
	}
	return best
}

func Spread(state GameState, player Color) float64 {
	positions := make([]float64, 0, 15)
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner != player || pt.Count == 0 {
			continue
		}
		for i := 0; i < pt.Count; i++ {
			positions = append(positions, float64(p))
		}
	}
	if len(positions) == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range positions {
		mean += v
	}
	mean /= float64(len(positions))
	varSum := 0.0
	for _, v := range positions {
		d := v - mean
		varSum += d * d
	}
	return math.Sqrt(varSum / float64(len(positions)))
}

func CountSingletons(state GameState, player Color) int {
	count := 0
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner == player && pt.Count == 1 {
			count++
		}
	}
	return count
}

func HomeMass(state GameState, player Color) int {
	hFrom, hTo := HomeRange(player)
	total := 0
	for p := hFrom; p <= hTo; p++ {
		pt := state.Points[p]
		if pt.Owner == player {
			total += pt.Count
		}
	}
	return total
}
