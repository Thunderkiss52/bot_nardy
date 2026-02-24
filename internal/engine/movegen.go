package engine

import (
	"fmt"
	"sort"
)

func GenerateLegalLines(state GameState, d1, d2 int) ([]TurnLine, error) {
	if d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6 {
		return nil, fmt.Errorf("invalid dice: %d,%d", d1, d2)
	}
	if err := state.Validate(); err != nil {
		return nil, err
	}

	seqs := diceSequences(d1, d2)
	collected := make(map[string]TurnLine)
	maxUsed := 0

	for _, seq := range seqs {
		collectBySequence(state, seq, 0, TurnLine{}, false, collected, &maxUsed)
	}

	if len(collected) == 0 {
		return []TurnLine{}, nil
	}
	if maxUsed == 0 {
		return []TurnLine{}, nil
	}

	highDie := d1
	if d2 > highDie {
		highDie = d2
	}
	preferHighOnly := d1 != d2 && maxUsed == 1
	hasHigh := false
	if preferHighOnly {
		for _, line := range collected {
			if line.DiceUsed() == 1 && line.Moves[0].Die == highDie {
				hasHigh = true
				break
			}
		}
	}

	lines := make([]TurnLine, 0, len(collected))
	for _, line := range collected {
		if line.DiceUsed() != maxUsed {
			continue
		}
		if preferHighOnly && hasHigh && line.Moves[0].Die != highDie {
			continue
		}
		lines = append(lines, line)
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Key() < lines[j].Key()
	})
	return lines, nil
}

func diceSequences(d1, d2 int) [][]int {
	if d1 == d2 {
		return [][]int{{d1, d1, d1, d1}}
	}
	return [][]int{{d1, d2}, {d2, d1}}
}

func collectBySequence(state GameState, seq []int, index int, prefix TurnLine, headTaken bool, out map[string]TurnLine, maxUsed *int) {
	if index >= len(seq) {
		insertLine(prefix, out, maxUsed)
		return
	}

	die := seq[index]
	moves := legalMovesForDie(state, die, headTaken)
	if len(moves) == 0 {
		insertLine(prefix, out, maxUsed)
		return
	}

	for _, mv := range moves {
		next := state.Clone()
		_ = applySingleMove(&next, mv)
		nextHead := headTaken
		if next.GameType == GameLong && mv.From == HeadPoint(next.Turn) {
			nextHead = true
		}
		newPrefix := TurnLine{Moves: make([]Move, len(prefix.Moves)+1)}
		copy(newPrefix.Moves, prefix.Moves)
		newPrefix.Moves[len(prefix.Moves)] = mv
		collectBySequence(next, seq, index+1, newPrefix, nextHead, out, maxUsed)
	}
}

func insertLine(line TurnLine, out map[string]TurnLine, maxUsed *int) {
	key := line.Key()
	if _, exists := out[key]; !exists {
		out[key] = line
		if line.DiceUsed() > *maxUsed {
			*maxUsed = line.DiceUsed()
		}
	}
}

func legalMovesForDie(state GameState, die int, headTaken bool) []Move {
	player := state.Turn
	opp := player.Opponent()
	moves := make([]Move, 0, 32)

	if state.GameType == GameShort && state.Bar[player.Idx()] > 0 {
		to := EntryPointFromBar(player, die)
		pt := state.Points[to]
		if pt.Owner == opp && pt.Count >= 2 {
			return moves
		}
		mv := Move{From: 0, To: to, Die: die}
		moves = append(moves, mv)
		return moves
	}

	for _, from := range SortedPointsForPlayer(player) {
		pt := state.Points[from]
		if pt.Owner != player || pt.Count == 0 {
			continue
		}
		if state.GameType == GameLong && from == HeadPoint(player) && headTaken {
			continue
		}

		to := from + player.Direction()*die
		if to >= 1 && to <= 24 {
			dst := state.Points[to]
			if state.GameType == GameShort {
				if dst.Owner == opp && dst.Count >= 2 {
					continue
				}
			} else {
				if dst.Owner == opp && dst.Count > 0 {
					continue
				}
			}

			mv := Move{From: from, To: to, Die: die}
			if state.GameType == GameLong {
				next := state.Clone()
				_ = applySingleMove(&next, mv)
				if violatesLongBlockade(next, player) {
					continue
				}
			}
			moves = append(moves, mv)
			continue
		}

		if canBearOff(state, player, from, die) {
			moves = append(moves, Move{From: from, To: 0, Die: die})
		}
	}

	return moves
}

func canBearOff(state GameState, player Color, from int, die int) bool {
	if !allInHome(state, player) {
		return false
	}
	if !isHomePoint(player, from) {
		return false
	}

	distance := DistanceToBearOff(player, from)
	if die == distance {
		return true
	}
	if die < distance {
		return false
	}

	if player == White {
		for p := from + 1; p <= 6; p++ {
			if state.Points[p].Owner == White && state.Points[p].Count > 0 {
				return false
			}
		}
		return true
	}

	for p := from - 1; p >= 19; p-- {
		if state.Points[p].Owner == Black && state.Points[p].Count > 0 {
			return false
		}
	}
	return true
}

func allInHome(state GameState, player Color) bool {
	if state.GameType == GameShort && state.Bar[player.Idx()] > 0 {
		return false
	}
	hFrom, hTo := HomeRange(player)
	for p := 1; p <= 24; p++ {
		pt := state.Points[p]
		if pt.Owner != player || pt.Count == 0 {
			continue
		}
		if p < hFrom || p > hTo {
			return false
		}
	}
	return true
}

func isHomePoint(player Color, point int) bool {
	hFrom, hTo := HomeRange(player)
	return point >= hFrom && point <= hTo
}

func violatesLongBlockade(state GameState, mover Color) bool {
	if state.GameType != GameLong {
		return false
	}
	opp := mover.Opponent()
	for start := 1; start <= 19; start++ {
		if !runOwned(state, mover, start, 6) {
			continue
		}
		if mover == White {
			if hasOpponentBehindWhiteRun(state, opp, start) {
				return true
			}
		} else {
			if hasOpponentBehindBlackRun(state, opp, start+5) {
				return true
			}
		}
	}
	return false
}

func runOwned(state GameState, owner Color, start, length int) bool {
	for p := start; p < start+length; p++ {
		pt := state.Points[p]
		if pt.Owner != owner || pt.Count == 0 {
			return false
		}
	}
	return true
}

func hasOpponentBehindWhiteRun(state GameState, opp Color, runStart int) bool {
	for p := 1; p < runStart; p++ {
		pt := state.Points[p]
		if pt.Owner == opp && pt.Count > 0 {
			return true
		}
	}
	return false
}

func hasOpponentBehindBlackRun(state GameState, opp Color, runEnd int) bool {
	for p := 24; p > runEnd; p-- {
		pt := state.Points[p]
		if pt.Owner == opp && pt.Count > 0 {
			return true
		}
	}
	return false
}

func applySingleMove(state *GameState, mv Move) error {
	player := state.Turn
	opp := player.Opponent()

	if mv.From == 0 {
		if state.GameType != GameShort {
			return fmt.Errorf("bar move is invalid for long game")
		}
		if state.Bar[player.Idx()] <= 0 {
			return fmt.Errorf("no checkers on bar for %s", player)
		}
		state.Bar[player.Idx()]--
	} else {
		if mv.From < 1 || mv.From > 24 {
			return fmt.Errorf("invalid from point: %d", mv.From)
		}
		src := &state.Points[mv.From]
		if src.Owner != player || src.Count <= 0 {
			return fmt.Errorf("cannot move from %d", mv.From)
		}
		src.Count--
		if src.Count == 0 {
			src.Owner = NoColor
		}
	}

	if mv.To == 0 {
		state.Off[player.Idx()]++
		return nil
	}
	if mv.To < 1 || mv.To > 24 {
		return fmt.Errorf("invalid to point: %d", mv.To)
	}

	dst := &state.Points[mv.To]
	if state.GameType == GameShort && dst.Owner == opp && dst.Count == 1 {
		dst.Owner = NoColor
		dst.Count = 0
		state.Bar[opp.Idx()]++
	}
	if state.GameType == GameLong && dst.Owner == opp && dst.Count > 0 {
		return fmt.Errorf("cannot land on opponent point in long")
	}

	if dst.Count == 0 {
		dst.Owner = player
		dst.Count = 1
		return nil
	}
	if dst.Owner != player {
		return fmt.Errorf("destination blocked at %d", mv.To)
	}
	dst.Count++
	return nil
}

func ApplyTurnLine(state GameState, line TurnLine) (GameState, error) {
	next := state.Clone()
	for _, mv := range line.Moves {
		if err := applySingleMove(&next, mv); err != nil {
			return state, err
		}
	}
	next.Turn = state.Turn.Opponent()
	next.Meta.MoveNumber++
	if err := next.Validate(); err != nil {
		return state, err
	}
	return next, nil
}

func IsLegalLine(state GameState, d1, d2 int, line TurnLine) bool {
	lines, err := GenerateLegalLines(state, d1, d2)
	if err != nil {
		return false
	}
	key := line.Key()
	for _, legal := range lines {
		if legal.Key() == key {
			return true
		}
	}
	return false
}
