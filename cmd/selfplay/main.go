package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
)

func main() {
	game := flag.String("game", "short", "short|long")
	n := flag.Int("n", 100, "number of games")
	strongThink := flag.Int("strong-think", 4, "strong bot think time seconds")
	baseThink := flag.Int("base-think", 1, "baseline bot think time seconds")
	seed := flag.Int64("seed", time.Now().UnixNano(), "seed")
	flag.Parse()

	g := parseGameType(*game)
	strongWins := 0
	rng := rand.New(rand.NewSource(*seed))

	for i := 0; i < *n; i++ {
		strongColor := engine.White
		if i%2 == 1 {
			strongColor = engine.Black
		}

		strong := bot.New(bot.Config{
			ThinkTime: time.Duration(*strongThink) * time.Second,
			TopK:      10,
			Workers:   0,
			Seed:      *seed + int64(i)*13 + 1,
		})
		baseline := bot.New(bot.Config{
			ThinkTime: time.Duration(*baseThink) * time.Second,
			TopK:      3,
			Workers:   0,
			Seed:      *seed + int64(i)*13 + 2,
		})

		state := initialState(g, *seed+int64(i))
		winner := playGame(state, strongColor, strong, baseline, rng)
		if winner == strongColor {
			strongWins++
		}
	}

	fmt.Printf("game=%s n=%d strong_winrate=%.2f%%\n", g, *n, 100*float64(strongWins)/float64(*n))
}

func playGame(state engine.GameState, strongColor engine.Color, strong, baseline *bot.Bot, rng *rand.Rand) engine.Color {
	for ply := 0; ply < 2048; ply++ {
		if state.IsTerminal() {
			return state.Winner()
		}
		d1 := rng.Intn(6) + 1
		d2 := rng.Intn(6) + 1

		var thinker *bot.Bot
		if state.Turn == strongColor {
			thinker = strong
		} else {
			thinker = baseline
		}
		decision, err := thinker.ChooseMove(state, d1, d2)
		if err != nil {
			return state.Turn.Opponent()
		}
		if decision.LegalCount == 0 {
			state.Turn = state.Turn.Opponent()
			continue
		}
		next, err := engine.ApplyTurnLine(state, decision.ChosenLine)
		if err != nil {
			return state.Turn.Opponent()
		}
		state = next
	}
	if bot.EvaluatePosition(state, engine.White) >= 0 {
		return engine.White
	}
	return engine.Black
}

func initialState(gt engine.GameType, seed int64) engine.GameState {
	if gt == engine.GameLong {
		return engine.NewLongGame(seed)
	}
	return engine.NewShortGame(seed)
}

func parseGameType(s string) engine.GameType {
	if s == "long" {
		return engine.GameLong
	}
	return engine.GameShort
}
