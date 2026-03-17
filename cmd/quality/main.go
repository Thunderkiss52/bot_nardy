package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
)

type qualitySummary struct {
	Game           string  `json:"game"`
	Games          int     `json:"games"`
	StrongWinRate  float64 `json:"strong_win_rate"`
	MovesEvaluated int     `json:"moves_evaluated"`
	AvgEquityLoss  float64 `json:"avg_equity_loss"`
	Blunder03Rate  float64 `json:"blunder_rate_ge_003"`
	Blunder05Rate  float64 `json:"blunder_rate_ge_005"`
	StrongThink    string  `json:"strong_think"`
	BaselineThink  string  `json:"baseline_think"`
	GeneratedAtUTC string  `json:"generated_at_utc"`
}

type moveQuality struct {
	delta float64
}

func main() {
	game := flag.String("game", "short", "short|long|both")
	n := flag.Int("n", 1000, "number of games")
	strongThink := flag.Duration("strong-think", 800*time.Millisecond, "strong bot think time")
	baselineThink := flag.Duration("baseline-think", 120*time.Millisecond, "baseline bot think time")
	seed := flag.Int64("seed", time.Now().UnixNano(), "seed")
	out := flag.String("out", "", "optional output JSON file")
	flag.Parse()

	games := []engine.GameType{parseGameType(*game)}
	if strings.EqualFold(strings.TrimSpace(*game), "both") {
		games = []engine.GameType{engine.GameShort, engine.GameLong}
	}

	summaries := make([]qualitySummary, 0, len(games))
	for i, gt := range games {
		s := runQuality(gt, *n, *strongThink, *baselineThink, *seed+int64(i)*10000019)
		summaries = append(summaries, s)
		printSummary(s)
	}

	if strings.TrimSpace(*out) != "" {
		if err := writeJSON(*out, summaries); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("saved summary to %s\n", *out)
	}
}

func runQuality(gt engine.GameType, n int, strongThink, baselineThink time.Duration, seed int64) qualitySummary {
	rng := rand.New(rand.NewSource(seed))
	wins := 0
	quality := make([]moveQuality, 0, n*120)
	evaluator, err := bot.ResolveEvaluatorFromEnv()
	if err != nil {
		panic(err)
	}

	for i := 0; i < n; i++ {
		strongColor := engine.White
		if i%2 == 1 {
			strongColor = engine.Black
		}
		strong := bot.New(bot.Config{ThinkTime: strongThink, TopK: 12, MaxPlies: 640, Seed: seed + int64(i)*17 + 1, Evaluator: evaluator})
		baseline := bot.New(bot.Config{ThinkTime: baselineThink, TopK: 4, MaxPlies: 320, Seed: seed + int64(i)*17 + 2, Evaluator: evaluator})

		state := initialState(gt, seed+int64(i)*37)
		winner, evals := playAndMeasure(state, strongColor, strong, baseline, rng)
		if winner == strongColor {
			wins++
		}
		quality = append(quality, evals...)
	}

	avgLoss := 0.0
	b03 := 0
	b05 := 0
	for _, q := range quality {
		avgLoss += q.delta
		if q.delta >= 0.03 {
			b03++
		}
		if q.delta >= 0.05 {
			b05++
		}
	}
	moveN := len(quality)
	if moveN > 0 {
		avgLoss /= float64(moveN)
	}

	return qualitySummary{
		Game:           gt.String(),
		Games:          n,
		StrongWinRate:  float64(wins) / float64(n),
		MovesEvaluated: moveN,
		AvgEquityLoss:  avgLoss,
		Blunder03Rate:  ratio(b03, moveN),
		Blunder05Rate:  ratio(b05, moveN),
		StrongThink:    strongThink.String(),
		BaselineThink:  baselineThink.String(),
		GeneratedAtUTC: time.Now().UTC().Format(time.RFC3339),
	}
}

func playAndMeasure(state engine.GameState, strongColor engine.Color, strong, baseline *bot.Bot, rng *rand.Rand) (engine.Color, []moveQuality) {
	evals := make([]moveQuality, 0, 200)

	for ply := 0; ply < 2048; ply++ {
		if state.IsTerminal() {
			return state.Winner(), evals
		}
		d1 := rng.Intn(6) + 1
		d2 := rng.Intn(6) + 1
		lines, err := engine.GenerateLegalLines(state, d1, d2)
		if err != nil {
			return state.Turn.Opponent(), evals
		}
		if len(lines) == 0 {
			state.Turn = state.Turn.Opponent()
			state.Meta.MoveNumber++
			continue
		}

		thinker := baseline
		isStrongMove := state.Turn == strongColor
		if isStrongMove {
			thinker = strong
		}
		dec, err := thinker.ChooseMove(state, d1, d2)
		if err != nil {
			return state.Turn.Opponent(), evals
		}
		next, err := engine.ApplyTurnLine(state, dec.ChosenLine)
		if err != nil {
			return state.Turn.Opponent(), evals
		}

		if isStrongMove {
			delta := heuristicEquityLoss(state, lines, dec.ChosenLine)
			evals = append(evals, moveQuality{delta: delta})
		}
		state = next
	}

	if bot.EvaluatePosition(state, engine.White) >= 0 {
		return engine.White, evals
	}
	return engine.Black, evals
}

func heuristicEquityLoss(state engine.GameState, lines []engine.TurnLine, chosen engine.TurnLine) float64 {
	best := -1e18
	chosenScore := -1e18
	for _, line := range lines {
		next, err := engine.ApplyTurnLine(state, line)
		if err != nil {
			continue
		}
		s := bot.EvaluatePosition(next, state.Turn)
		if s > best {
			best = s
		}
		if line.Key() == chosen.Key() {
			chosenScore = s
		}
	}
	if chosenScore > best {
		chosenScore = best
	}
	if best < -1e17 || chosenScore < -1e17 {
		return 0
	}
	delta := (best - chosenScore) / 100.0
	if delta < 0 {
		return 0
	}
	if delta > 1 {
		return 1
	}
	return delta
}

func initialState(gt engine.GameType, seed int64) engine.GameState {
	if gt == engine.GameLong {
		return engine.NewLongGame(seed)
	}
	return engine.NewShortGame(seed)
}

func parseGameType(v string) engine.GameType {
	if strings.EqualFold(strings.TrimSpace(v), "long") {
		return engine.GameLong
	}
	return engine.GameShort
}

func ratio(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}

func printSummary(s qualitySummary) {
	fmt.Printf(
		"game=%s games=%d strong_winrate=%.2f%% ael=%.5f b03=%.3f%% b05=%.3f%% moves=%d\n",
		s.Game,
		s.Games,
		s.StrongWinRate*100,
		s.AvgEquityLoss,
		s.Blunder03Rate*100,
		s.Blunder05Rate*100,
		s.MovesEvaluated,
	)
}

func writeJSON(path string, summaries []qualitySummary) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(summaries)
}
