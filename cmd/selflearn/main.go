package main

import (
	"flag"
	"fmt"
	"os"

	"bot_nardy/internal/app"
)

func main() {
	epochs := flag.Int("epochs", 12, "number of training epochs")
	flag.Parse()

	svc, err := app.NewService("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "service init failed: %v\n", err)
		os.Exit(1)
	}
	defer svc.Close()

	result, err := svc.SelfLearn(*epochs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "selflearn failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"selflearn successful: examples=%d err=%.4f accepted=%v validation=%.3f games=%d champion=%s score=%.3f\n",
		result.Examples,
		result.AvgAbsError,
		result.Accepted,
		result.ValidationWinRate,
		result.ValidationGames,
		result.ChampionModel,
		result.ChampionScore,
	)
}
