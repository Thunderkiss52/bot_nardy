package gnubg

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/engine"
)

type SourceRecord struct {
	StateBefore engine.GameState `json:"state_before"`
	Dice        [2]int           `json:"dice"`
	LegalCount  int              `json:"legal_count"`
	ChosenLine  engine.TurnLine  `json:"chosen_line"`
	ChosenProb  float64          `json:"chosen_prob"`
	WinProb     float64          `json:"winprob"`
}

type TeacherRecord struct {
	StateBefore engine.GameState     `json:"state_before"`
	Dice        [2]int               `json:"dice"`
	LegalCount  int                  `json:"legal_count"`
	Top3        []bot.MoveEvaluation `json:"top3"`
	ChosenLine  engine.TurnLine      `json:"chosen_line"`
	WinProb     float64              `json:"winprob"`
}

type LabelSummary struct {
	InputPath  string `json:"input_path"`
	OutputPath string `json:"output_path"`
	Labeled    int    `json:"labeled"`
	Skipped    int    `json:"skipped"`
}

func LabelFile(inPath, outPath string, teacher Teacher, limit, top int) (LabelSummary, error) {
	in, err := os.Open(inPath)
	if err != nil {
		return LabelSummary{}, err
	}
	defer in.Close()

	out, err := os.Create(outPath)
	if err != nil {
		return LabelSummary{}, err
	}
	defer out.Close()

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	enc := json.NewEncoder(out)

	summary := LabelSummary{InputPath: inPath, OutputPath: outPath}
	for scanner.Scan() {
		if limit > 0 && summary.Labeled >= limit {
			break
		}
		record, ok, err := parseSourceRecord(scanner.Bytes())
		if err != nil {
			return summary, err
		}
		if !ok {
			summary.Skipped++
			continue
		}
		labeled, ok, err := teacherLabelRecord(teacher, record, top)
		if err != nil {
			return summary, err
		}
		if !ok {
			summary.Skipped++
			continue
		}
		if err := enc.Encode(labeled); err != nil {
			return summary, err
		}
		summary.Labeled++
	}
	if err := scanner.Err(); err != nil {
		return summary, err
	}
	return summary, nil
}

func teacherLabelRecord(teacher Teacher, rec SourceRecord, top int) (TeacherRecord, bool, error) {
	if rec.StateBefore.Validate() != nil {
		return TeacherRecord{}, false, nil
	}
	if rec.StateBefore.GameType != engine.GameShort {
		return TeacherRecord{}, false, nil
	}
	if rec.Dice[0] < 1 || rec.Dice[0] > 6 || rec.Dice[1] < 1 || rec.Dice[1] > 6 {
		return TeacherRecord{}, false, nil
	}
	hints, err := teacher.Analyze(rec.StateBefore, rec.Dice[0], rec.Dice[1], top)
	if err != nil {
		return TeacherRecord{}, false, err
	}
	chosenProb := rec.WinProb
	if chosenProb == 0 {
		chosenProb = rec.ChosenProb
	}
	for _, hint := range hints {
		if hint.Line.Key() == rec.ChosenLine.Key() {
			chosenProb = hint.WinProb
			break
		}
	}
	return TeacherRecord{
		StateBefore: rec.StateBefore,
		Dice:        rec.Dice,
		LegalCount:  rec.LegalCount,
		Top3:        HintMovesToEvaluations(hints),
		ChosenLine:  rec.ChosenLine,
		WinProb:     chosenProb,
	}, true, nil
}

func parseSourceRecord(raw []byte) (SourceRecord, bool, error) {
	var rec SourceRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return SourceRecord{}, false, err
	}
	if rec.StateBefore.GameType == 0 {
		return SourceRecord{}, false, nil
	}
	return rec, true, nil
}

func DefaultTeacher(binary string) Teacher {
	return Teacher{
		Binary:  strings.TrimSpace(binary),
		Timeout: 20 * time.Second,
	}
}
