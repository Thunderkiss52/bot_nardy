package gnubg

import (
	"encoding/base64"
	"strings"

	"bot_nardy/internal/engine"
)

// EncodePositionID converts a checker-only position into GNU Backgammon's 14-char Position ID.
// The encoding follows the official format relative to the player on roll.
func EncodePositionID(state engine.GameState) string {
	bits := make([]byte, 0, 80)
	appendPlayer := func(player engine.Color) {
		for rel := 1; rel <= 24; rel++ {
			abs := absolutePointFromRelative(player, rel)
			count := 0
			pt := state.Points[abs]
			if pt.Owner == player {
				count = pt.Count
			}
			for i := 0; i < count; i++ {
				bits = append(bits, 1)
			}
			bits = append(bits, 0)
		}
		for i := 0; i < state.Bar[player.Idx()]; i++ {
			bits = append(bits, 1)
		}
		bits = append(bits, 0)
	}

	appendPlayer(state.Turn)
	appendPlayer(state.Turn.Opponent())
	for len(bits) < 80 {
		bits = append(bits, 0)
	}

	raw := make([]byte, 10)
	for idx, bit := range bits[:80] {
		if bit == 0 {
			continue
		}
		raw[idx/8] |= 1 << (idx % 8)
	}

	encoded := base64.StdEncoding.EncodeToString(raw)
	return strings.TrimRight(encoded, "=")
}

func absolutePointFromRelative(player engine.Color, rel int) int {
	if player == engine.Black {
		return 25 - rel
	}
	return rel
}
