package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
	"bot_nardy/internal/training"
)

func main() {
	game := flag.String("game", "short", "short|long")
	n := flag.Int("n", 100, "number of games")
	strongThink := flag.Int("strong-think", 4, "strong bot think time seconds")
	baseThink := flag.Int("base-think", 1, "baseline bot think time seconds")
	datasetOut := flag.String("dataset-out", "", "optional JSONL path for self-play training examples")
	seed := flag.Int64("seed", time.Now().UnixNano(), "seed")
	flag.Parse()

	g := parseGameType(*game)
	strongWins := 0
	rng := rand.New(rand.NewSource(*seed))
	evaluator, err := bot.ResolveEvaluatorFromEnv()
	if err != nil {
		panic(err)
	}
	var datasetWriter *training.JSONLWriter
	if *datasetOut != "" {
		datasetWriter, err = training.NewJSONLWriter(*datasetOut)
		if err != nil {
			panic(err)
		}
		defer datasetWriter.Close()
	}

	for i := 0; i < *n; i++ {
		strongColor := engine.White
		if i%2 == 1 {
			strongColor = engine.Black
		}

		strong := bot.New(bot.Config{
			ThinkTime: time.Duration(*strongThink) * time.Second,
			TopK:      12,
			Workers:   0,
			MaxPlies:  640,
			Seed:      *seed + int64(i)*13 + 1,
			Evaluator: evaluator,
		})
		baseline := bot.New(bot.Config{
			ThinkTime: time.Duration(*baseThink) * time.Second,
			TopK:      3,
			Workers:   0,
			MaxPlies:  320,
			Seed:      *seed + int64(i)*13 + 2,
			Evaluator: evaluator,
		})

		state := initialState(g, *seed+int64(i))
		winner, examples := playGame(state, strongColor, strong, baseline, rng)
		if winner == strongColor {
			strongWins++
		}
		if datasetWriter != nil && len(examples) > 0 {
			for idx := range examples {
				examples[idx].Winner = winner.String()
				if examples[idx].Perspective == winner.String() {
					examples[idx].OutcomeValue = 1
				} else {
					examples[idx].OutcomeValue = -1
				}
			}
			if err := datasetWriter.WriteAll(examples); err != nil {
				panic(err)
			}
		}
	}

	fmt.Printf("game=%s n=%d strong_winrate=%.2f%%\n", g, *n, 100*float64(strongWins)/float64(*n))
}

func playGame(state engine.GameState, strongColor engine.Color, strong, baseline *bot.Bot, rng *rand.Rand) (engine.Color, []training.Example) {
	examples := make([]training.Example, 0, 256)
	for ply := 0; ply < 2048; ply++ {
		if state.IsTerminal() {
			return state.Winner(), examples
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
			return state.Turn.Opponent(), examples
		}
		if decision.LegalCount == 0 {
			state.Turn = state.Turn.Opponent()
			continue
		}
		examples = append(examples, training.Example{
			StateBefore:   state,
			GameType:      state.GameType.String(),
			StateKey:      state.NormalizeKey(),
			Perspective:   state.Turn.String(),
			Turn:          state.Turn.String(),
			MoveNumber:    state.Meta.MoveNumber,
			Dice:          [2]int{d1, d2},
			LegalCount:    decision.LegalCount,
			ChosenLine:    decision.ChosenLine,
			ChosenProb:    decision.ChosenProb,
			HasMoveTarget: true,
			MoveTarget:    clamp(2*decision.ChosenProb-1, -1, 1),
			Features:      bot.ExtractFeatures(state, state.Turn),
		})
		next, err := engine.ApplyTurnLine(state, decision.ChosenLine)
		if err != nil {
			return state.Turn.Opponent(), examples
		}
		state = next
	}
	if bot.EvaluatePosition(state, engine.White) >= 0 {
		return engine.White, examples
	}
	return engine.Black, examples
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
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
