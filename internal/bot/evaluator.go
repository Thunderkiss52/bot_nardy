package bot

import (
	"encoding/json"
	"os"

	"bot_nardy/internal/engine"
)

type Evaluator interface {
	Name() string
	Evaluate(state engine.GameState, perspective engine.Color) float64
}

type HeuristicEvaluator struct{}

func (HeuristicEvaluator) Name() string {
	return "heuristic-v1"
}

func (HeuristicEvaluator) Evaluate(state engine.GameState, perspective engine.Color) float64 {
	if state.IsTerminal() {
		if state.Winner() == perspective {
			return 1_000_000
		}
		return -1_000_000
	}
	if state.GameType == engine.GameShort {
		s := shortScore(state, perspective)
		o := shortScore(state, perspective.Opponent())
		return s - o
	}
	s := longScore(state, perspective)
	o := longScore(state, perspective.Opponent())
	return s - o
}

var defaultEvaluator Evaluator = HeuristicEvaluator{}

func EvaluatePosition(state engine.GameState, perspective engine.Color) float64 {
	return defaultEvaluator.Evaluate(state, perspective)
}

type LinearEvaluator struct {
	name    string
	bias    float64
	weights map[string]float64
	scales  map[string]float64
	base    Evaluator
}

type linearWeightsFile struct {
	Name    string             `json:"name"`
	Bias    float64            `json:"bias"`
	Weights map[string]float64 `json:"weights"`
	Scales  map[string]float64 `json:"scales,omitempty"`
}

func NewLinearEvaluator(name string, bias float64, weights map[string]float64, scales map[string]float64, base Evaluator) *LinearEvaluator {
	if base == nil {
		base = defaultEvaluator
	}
	return &LinearEvaluator{name: name, bias: bias, weights: weights, scales: scales, base: base}
}

func (e *LinearEvaluator) Name() string {
	if e == nil || e.name == "" {
		return "linear"
	}
	return e.name
}

func (e *LinearEvaluator) Evaluate(state engine.GameState, perspective engine.Color) float64 {
	if e == nil {
		return defaultEvaluator.Evaluate(state, perspective)
	}
	features := ExtractFeatures(state, perspective).Values()
	score := e.bias
	for key, value := range features {
		scale := 1.0
		if e.scales != nil {
			if s := e.scales[key]; s > 0 {
				scale = s
			}
		}
		score += e.weights[key] * (value / scale)
	}
	if e.base != nil {
		score += 0.15 * e.base.Evaluate(state, perspective)
	}
	return score
}

func LoadLinearEvaluator(path string, base Evaluator) (Evaluator, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg linearWeightsFile
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return NewLinearEvaluator(cfg.Name, cfg.Bias, cfg.Weights, cfg.Scales, base), nil
}

func ResolveEvaluatorFromEnv() (Evaluator, error) {
	weightsPath := os.Getenv("BOT_NARDY_WEIGHTS_PATH")
	if weightsPath == "" {
		return defaultEvaluator, nil
	}
	return LoadLinearEvaluator(weightsPath, defaultEvaluator)
}
