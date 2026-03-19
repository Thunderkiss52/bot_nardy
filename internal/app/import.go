package app

import (
	"bufio"
	"encoding/json"
	"os"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
	"bot_nardy/internal/training"
)

type importedMoveRecord struct {
	StateBefore engine.GameState     `json:"state_before"`
	Dice        [2]int               `json:"dice"`
	LegalCount  int                  `json:"legal_count"`
	Top3        []bot.MoveEvaluation `json:"top3"`
	ChosenLine  engine.TurnLine      `json:"chosen_line"`
	WinProb     float64              `json:"winprob"`
}

type ImportLogResult struct {
	Path     string `json:"path"`
	Imported int    `json:"imported"`
}

func (s *Service) ImportMoveLog(path string) (ImportLogResult, error) {
	examples, err := loadImportedExamples(path)
	if err != nil {
		return ImportLogResult{}, err
	}
	if len(examples) == 0 {
		return ImportLogResult{Path: path}, nil
	}

	s.trainingMu.Lock()
	defer s.trainingMu.Unlock()
	if s.experienceWriter == nil {
		return ImportLogResult{}, os.ErrClosed
	}
	if err := s.experienceWriter.WriteAll(examples); err != nil {
		return ImportLogResult{}, err
	}
	return ImportLogResult{Path: path, Imported: len(examples)}, nil
}

func loadImportedExamples(path string) ([]training.Example, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	examples := make([]training.Example, 0, 1024)
	seen := make(map[string]struct{}, 1024)
	for scanner.Scan() {
		var rec importedMoveRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			return nil, err
		}
		if rec.StateBefore.Validate() != nil {
			continue
		}
		if rec.Dice[0] < 1 || rec.Dice[0] > 6 || rec.Dice[1] < 1 || rec.Dice[1] > 6 {
			continue
		}
		key := rec.StateBefore.NormalizeKey() + "|" + rec.ChosenLine.Key()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		moveTarget, hasMoveTarget := importedMoveTarget(rec)
		outcome := probabilityTarget(rec.WinProb)
		examples = append(examples, training.Example{
			StateBefore:   rec.StateBefore,
			GameType:      rec.StateBefore.GameType.String(),
			StateKey:      rec.StateBefore.NormalizeKey(),
			Perspective:   rec.StateBefore.Turn.String(),
			Turn:          rec.StateBefore.Turn.String(),
			MoveNumber:    rec.StateBefore.Meta.MoveNumber,
			Dice:          rec.Dice,
			LegalCount:    rec.LegalCount,
			ChosenLine:    rec.ChosenLine,
			ChosenProb:    rec.WinProb,
			HasMoveTarget: hasMoveTarget,
			MoveTarget:    moveTarget,
			Winner:        "",
			OutcomeValue:  outcome,
			Features:      bot.ExtractFeatures(rec.StateBefore, rec.StateBefore.Turn),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return examples, nil
}

func importedMoveTarget(rec importedMoveRecord) (float64, bool) {
	if len(rec.Top3) == 0 {
		if rec.WinProb > 0 {
			return probabilityTarget(rec.WinProb), true
		}
		return 0, false
	}
	best := rec.Top3[0].WinProb
	chosen := rec.WinProb
	if chosen <= 0 {
		for _, cand := range rec.Top3 {
			if cand.Line.Key() == rec.ChosenLine.Key() {
				chosen = cand.WinProb
				break
			}
		}
	}
	if chosen <= 0 {
		chosen = best
	}
	gap := best - chosen
	if gap <= 0 {
		return 1.0, true
	}
	return 1.0 - 2.0*clampFloat(gap/0.35, 0, 1), true
}
