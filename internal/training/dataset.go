package training

import (
	"encoding/json"
	"os"
	"sync"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
)

type Example struct {
	StateBefore   engine.GameState  `json:"state_before,omitempty"`
	GameType      string            `json:"game_type"`
	StateKey      string            `json:"state_key"`
	Perspective   string            `json:"perspective"`
	Turn          string            `json:"turn"`
	MoveNumber    int               `json:"move_number"`
	Dice          [2]int            `json:"dice"`
	LegalCount    int               `json:"legal_count"`
	ChosenLine    engine.TurnLine   `json:"chosen_line"`
	ChosenProb    float64           `json:"chosen_prob"`
	HasMoveTarget bool              `json:"has_move_target"`
	MoveTarget    float64           `json:"move_target"`
	Winner        string            `json:"winner"`
	OutcomeValue  float64           `json:"outcome_value"`
	Features      bot.FeatureVector `json:"features"`
}

type JSONLWriter struct {
	mu sync.Mutex
	f  *os.File
}

func NewJSONLWriter(path string) (*JSONLWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &JSONLWriter{f: f}, nil
}

func (w *JSONLWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	return w.f.Close()
}

func (w *JSONLWriter) WriteAll(examples []Example) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return os.ErrClosed
	}
	enc := json.NewEncoder(w.f)
	for _, ex := range examples {
		if err := enc.Encode(ex); err != nil {
			return err
		}
	}
	return nil
}
