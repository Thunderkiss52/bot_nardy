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

func RaceLead(state GameState, player Color) int {
	return PipCount(state, player.Opponent()) - PipCount(state, player)
}

func EntryBlockPoints(state GameState, player Color) int {
	count := 0
	hFrom, hTo := HomeRange(player)
	for p := hFrom; p <= hTo; p++ {
		pt := state.Points[p]
		if pt.Owner == player && pt.Count >= 2 {
			count++
		}
	}
	return count
}

func CountAnchorsShort(state GameState, player Color) int {
	oppFrom, oppTo := HomeRange(player.Opponent())
	count := 0
	for p := oppFrom; p <= oppTo; p++ {
		pt := state.Points[p]
		if pt.Owner == player && pt.Count >= 2 {
			count++
		}
	}
	return count
}

func CountEscapedCheckersShort(state GameState, player Color) int {
	oppExtreme := 0
	if player == White {
		oppExtreme = 25
		for p := 1; p <= 24; p++ {
			pt := state.Points[p]
			if pt.Owner == Black && pt.Count > 0 && p < oppExtreme {
				oppExtreme = p
			}
		}
		if oppExtreme == 25 {
			return 15
		}
		count := 0
		for p := 1; p < oppExtreme; p++ {
			pt := state.Points[p]
			if pt.Owner == White {
				count += pt.Count
			}
		}
		return count
	}

	for p := 24; p >= 1; p-- {
		pt := state.Points[p]
		if pt.Owner == White && pt.Count > 0 {
			oppExtreme = p
			break
		}
	}
	if oppExtreme == 0 {
		return 15
	}
	count := 0
	for p := oppExtreme + 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner == Black {
			count += pt.Count
		}
	}
	return count
}

func BackmostPoint(state GameState, player Color) int {
	if player == White {
		for p := 24; p >= 1; p-- {
			pt := state.Points[p]
			if pt.Owner == White && pt.Count > 0 {
				return p
			}
		}
		return 0
	}
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner == Black && pt.Count > 0 {
			return p
		}
	}
	return 25
}

func CheckersOnHead(state GameState, player Color) int {
	head := HeadPoint(player)
	pt := state.Points[head]
	if pt.Owner != player {
		return 0
	}
	return pt.Count
}

func CountPointsInOuterBoard(state GameState, player Color) int {
	count := 0
	if player == White {
		for p := 7; p <= 18; p++ {
			pt := state.Points[p]
			if pt.Owner == White && pt.Count >= 2 {
				count++
			}
		}
		return count
	}
	for p := 7; p <= 18; p++ {
		pt := state.Points[p]
		if pt.Owner == Black && pt.Count >= 2 {
			count++
		}
	}
	return count
}

func CountExposedBlotsShort(state GameState, player Color) int {
	count := 0
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner != player || pt.Count != 1 {
			continue
		}
		if directShotCountOnPointShort(state, player.Opponent(), p) > 0 {
			count++
		}
	}
	return count
}

func CountDirectShotsShort(state GameState, attacker Color) int {
	total := 0
	defender := attacker.Opponent()
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner != defender || pt.Count != 1 {
			continue
		}
		total += directShotCountOnPointShort(state, attacker, p)
	}
	return total
}

func directShotCountOnPointShort(state GameState, attacker Color, target int) int {
	if state.GameType != GameShort {
		return 0
	}

	shots := 0
	if state.Bar[attacker.Idx()] > 0 {
		die := 0
		if attacker == White {
			die = 25 - target
		} else {
			die = target
		}
		if die >= 1 && die <= 6 {
			shots++
		}
		return shots
	}

	for die := 1; die <= 6; die++ {
		from := target - attacker.Direction()*die
		if from < 1 || from > 24 {
			continue
		}
		pt := state.Points[from]
		if pt.Owner == attacker && pt.Count > 0 {
			shots++
		}
	}
	return shots
}
