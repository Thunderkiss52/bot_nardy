package main

import (
	"flag"
	"fmt"
	"os"

	"bot_nardy/internal/app"
)

func main() {
	file := flag.String("file", "", "path to moves.jsonl")
	epochs := flag.Int("selflearn", 0, "run self-learning after import with N epochs")
	flag.Parse()

	if *file == "" {
		fmt.Fprintln(os.Stderr, "file path is required")
		os.Exit(1)
	}

	svc, err := app.NewService("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "service init failed: %v\n", err)
		os.Exit(1)
	}
	defer svc.Close()

	result, err := svc.ImportMoveLog(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "import failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("import successful: imported %d examples from %s\n", result.Imported, result.Path)

	if *epochs > 0 {
		train, err := svc.SelfLearn(*epochs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "selflearn failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("selflearn: examples=%d err=%.4f accepted=%v champion=%s score=%.3f\n",
			train.Examples, train.AvgAbsError, train.Accepted, train.ChampionModel, train.ChampionScore)
	}
}
