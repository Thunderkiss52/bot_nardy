package main

import (
	"flag"
	"fmt"
	"os"

	"bot_nardy/internal/app"
	"bot_nardy/internal/gnubg"
)

func main() {
	in := flag.String("in", "moves.jsonl", "input JSONL move log path")
	out := flag.String("out", "gnubg_teacher.jsonl", "output JSONL path compatible with ImportMoveLog")
	binary := flag.String("gnubg", "gnubg", "path to GNU Backgammon CLI binary")
	limit := flag.Int("limit", 0, "max number of positions to analyze (0 = all)")
	top := flag.Int("top", 3, "number of top GNU BG moves to export")
	selflearn := flag.Int("selflearn", 0, "run self-learning after importing teacher labels")
	flag.Parse()

	summary, err := gnubg.LabelFile(*in, *out, gnubg.DefaultTeacher(*binary), *limit, *top)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gnubg import failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("gnubg import successful: wrote %d teacher-labeled examples to %s (skipped %d)\n", summary.Labeled, summary.OutputPath, summary.Skipped)

	if *selflearn > 0 {
		svc, err := app.NewService("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "service init failed: %v\n", err)
			os.Exit(1)
		}
		defer svc.Close()
		result, err := svc.ImportMoveLog(*out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "import failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("import successful: imported %d examples from %s\n", result.Imported, result.Path)
		summary, err := svc.SelfLearn(*selflearn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "selflearn failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("selflearn successful: examples=%d err=%.4f accepted=%v validation=%.3f games=%d champion=%s score=%.3f\n",
			summary.Examples,
			summary.AvgAbsError,
			summary.Accepted,
			summary.ValidationWinRate,
			summary.ValidationGames,
			summary.ChampionModel,
			summary.ChampionScore,
		)
	}
}
