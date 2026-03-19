package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"bot_nardy/internal/bot"
	"bot_nardy/internal/gnubg"
)

type summary struct {
	InputPath      string  `json:"input_path"`
	Positions      int     `json:"positions"`
	ExactRate      float64 `json:"exact_rate"`
	TopKRate       float64 `json:"topk_rate"`
	MissedTopKRate float64 `json:"missed_topk_rate"`
	AvgTeacherLoss float64 `json:"avg_teacher_loss"`
	ThinkTime      string  `json:"think_time"`
	GeneratedAtUTC string  `json:"generated_at_utc"`
}

func main() {
	in := flag.String("in", "gnubg_teacher.jsonl", "teacher-labeled JSONL path")
	think := flag.Duration("think", 2*time.Second, "bot think time")
	topK := flag.Int("topk", 3, "teacher top-k window for agreement")
	out := flag.String("out", "", "optional output JSON path")
	flag.Parse()

	s, err := run(*in, *think, *topK)
	if err != nil {
		fmt.Fprintf(os.Stderr, "teacher quality failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(
		"teacher_quality positions=%d exact=%.2f%% topk=%.2f%% miss=%.2f%% avg_loss=%.4f think=%s\n",
		s.Positions,
		s.ExactRate*100,
		s.TopKRate*100,
		s.MissedTopKRate*100,
		s.AvgTeacherLoss,
		s.ThinkTime,
	)
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create %s: %v\n", *out, err)
			os.Exit(1)
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(s); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", *out, err)
			os.Exit(1)
		}
		fmt.Printf("saved teacher quality summary to %s\n", *out)
	}
}

func run(path string, think time.Duration, topK int) (summary, error) {
	f, err := os.Open(path)
	if err != nil {
		return summary{}, err
	}
	defer f.Close()

	evaluator, err := bot.ResolveEvaluatorFromEnv()
	if err != nil {
		return summary{}, err
	}
	player := bot.New(bot.Config{
		ThinkTime: think,
		TopK:      12,
		MaxPlies:  640,
		Evaluator: evaluator,
	})

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	var positions, exact, topkHits, topkMisses int
	var totalLoss float64
	for scanner.Scan() {
		var rec gnubg.TeacherRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			return summary{}, err
		}
		if rec.StateBefore.Validate() != nil || len(rec.Top3) == 0 {
			continue
		}
		decision, err := player.ChooseMove(rec.StateBefore, rec.Dice[0], rec.Dice[1])
		if err != nil || decision.LegalCount == 0 {
			continue
		}
		positions++
		best := rec.Top3[0]
		if decision.ChosenLine.Key() == best.Line.Key() {
			exact++
		}
		loss, hit := teacherLoss(decision.ChosenLine, rec.Top3, topK)
		if hit {
			topkHits++
		} else {
			topkMisses++
		}
		totalLoss += loss
	}
	if err := scanner.Err(); err != nil {
		return summary{}, err
	}

	s := summary{
		InputPath:      path,
		Positions:      positions,
		ThinkTime:      think.String(),
		GeneratedAtUTC: time.Now().UTC().Format(time.RFC3339),
	}
	if positions > 0 {
		s.ExactRate = float64(exact) / float64(positions)
		s.TopKRate = float64(topkHits) / float64(positions)
		s.MissedTopKRate = float64(topkMisses) / float64(positions)
		s.AvgTeacherLoss = totalLoss / float64(positions)
	}
	return s, nil
}

func teacherLoss(chosenLine interface{ Key() string }, top []bot.MoveEvaluation, topK int) (float64, bool) {
	if len(top) == 0 {
		return 0, false
	}
	if topK <= 0 || topK > len(top) {
		topK = len(top)
	}
	best := top[0].WinProb
	worstTopK := top[topK-1].WinProb
	chosenKey := chosenLine.Key()
	for idx, move := range top {
		if idx >= topK {
			break
		}
		if move.Line.Key() == chosenKey {
			loss := best - move.WinProb
			if loss < 0 {
				loss = 0
			}
			return loss, true
		}
	}
	loss := best - worstTopK + 0.02
	if loss < 0 {
		loss = 0
	}
	return loss, false
}
