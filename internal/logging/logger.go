package logging

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
)

type MoveLog struct {
	Timestamp   time.Time            `json:"ts"`
	StateBefore engine.GameState     `json:"state_before"`
	Dice        [2]int               `json:"dice"`
	LegalCount  int                  `json:"legal_count"`
	Top3        []bot.MoveEvaluation `json:"top3"`
	ChosenLine  engine.TurnLine      `json:"chosen_line"`
	WinProb     float64              `json:"winprob"`
	ThinkMillis int64                `json:"think_ms"`
	Seed        int64                `json:"seed"`
}

type JSONLLogger struct {
	mu sync.Mutex
	f  *os.File
}

func NewJSONLLogger(path string) (*JSONLLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &JSONLLogger{f: f}, nil
}

func (l *JSONLLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		return nil
	}
	return l.f.Close()
}

func (l *JSONLLogger) Write(entry MoveLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		return os.ErrClosed
	}
	enc := json.NewEncoder(l.f)
	return enc.Encode(entry)
}
