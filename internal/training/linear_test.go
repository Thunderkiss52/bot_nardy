package training

import (
	"path/filepath"
	"testing"

	"bot_nardy/internal/bot"
)

func TestTrainLinearFromJSONLProducesWeightsFile(t *testing.T) {
	dir := t.TempDir()
	datasetPath := filepath.Join(dir, "train.jsonl")
	weightsPath := filepath.Join(dir, "weights.json")

	writer, err := NewJSONLWriter(datasetPath)
	if err != nil {
		t.Fatalf("new writer: %v", err)
	}
	err = writer.WriteAll([]Example{
		{
			Perspective:   "white",
			Winner:        "white",
			OutcomeValue:  1,
			HasMoveTarget: true,
			MoveTarget:    0.9,
			Features: bot.FeatureVector{
				PipSelf:  40,
				PipOpp:   60,
				MadeSelf: 4,
			},
		},
		{
			Perspective:   "black",
			Winner:        "white",
			OutcomeValue:  -1,
			HasMoveTarget: true,
			MoveTarget:    -0.8,
			Features: bot.FeatureVector{
				PipSelf:  60,
				PipOpp:   40,
				MadeSelf: 2,
			},
		},
	})
	if err != nil {
		t.Fatalf("write all: %v", err)
	}
	_ = writer.Close()

	summary, err := TrainLinearFromJSONL(datasetPath, weightsPath, TrainConfig{
		Epochs:       4,
		LearningRate: 0.00001,
		L2:           0,
		MaxExamples:  100,
	})
	if err != nil {
		t.Fatalf("train: %v", err)
	}
	if summary.Examples != 2 {
		t.Fatalf("expected 2 examples, got %d", summary.Examples)
	}
	if summary.WeightsPath != weightsPath {
		t.Fatalf("expected weights path %s, got %s", weightsPath, summary.WeightsPath)
	}
}
