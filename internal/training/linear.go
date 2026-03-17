package training

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
)

type TrainConfig struct {
	Epochs       int     `json:"epochs"`
	LearningRate float64 `json:"learning_rate"`
	L2           float64 `json:"l2"`
	MaxExamples  int     `json:"max_examples"`
}

type TrainSummary struct {
	Examples          int     `json:"examples"`
	Epochs            int     `json:"epochs"`
	LearningRate      float64 `json:"learning_rate"`
	L2                float64 `json:"l2"`
	AvgAbsError       float64 `json:"avg_abs_error"`
	WeightsPath       string  `json:"weights_path"`
	ModelName         string  `json:"model_name"`
	Accepted          bool    `json:"accepted"`
	ValidationGames   int     `json:"validation_games"`
	ValidationWinRate float64 `json:"validation_win_rate"`
	LeagueSize        int     `json:"league_size"`
	ChampionModel     string  `json:"champion_model"`
	ChampionScore     float64 `json:"champion_score"`
}

type linearWeightsFile struct {
	Name    string             `json:"name"`
	Bias    float64            `json:"bias"`
	Weights map[string]float64 `json:"weights"`
}

func TrainLinearFromJSONL(datasetPath, weightsPath string, cfg TrainConfig) (TrainSummary, error) {
	if cfg.Epochs <= 0 {
		cfg.Epochs = 12
	}
	if cfg.LearningRate <= 0 {
		cfg.LearningRate = 0.00001
	}
	if cfg.L2 < 0 {
		cfg.L2 = 0
	}
	if cfg.MaxExamples <= 0 {
		cfg.MaxExamples = 50000
	}

	examples, err := loadExamples(datasetPath)
	if err != nil {
		return TrainSummary{}, err
	}
	if len(examples) == 0 {
		return TrainSummary{}, errors.New("no training examples found")
	}
	if len(examples) > cfg.MaxExamples {
		examples = examples[len(examples)-cfg.MaxExamples:]
	}

	model := linearWeightsFile{
		Name:    "linear-human-v3",
		Weights: map[string]float64{},
	}
	if loaded, err := loadWeights(weightsPath); err == nil {
		model = loaded
		if model.Weights == nil {
			model.Weights = map[string]float64{}
		}
		if model.Name == "" {
			model.Name = "linear-human-v3"
		}
	}

	totalAbsErr := 0.0
	steps := 0
	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		for _, ex := range examples {
			values := ex.Features.Values()
			pred := model.Bias
			for key, value := range values {
				pred += model.Weights[key] * value
			}
			target := targetForExample(ex)
			err := pred - target
			totalAbsErr += abs(err)
			steps++

			model.Bias -= cfg.LearningRate * err
			for key, value := range values {
				w := model.Weights[key]
				model.Weights[key] = w - cfg.LearningRate*(err*value+cfg.L2*w)
			}
		}
	}

	if err := saveWeights(weightsPath, model); err != nil {
		return TrainSummary{}, err
	}

	avgAbsErr := 0.0
	if steps > 0 {
		avgAbsErr = totalAbsErr / float64(steps)
	}
	return TrainSummary{
		Examples:     len(examples),
		Epochs:       cfg.Epochs,
		LearningRate: cfg.LearningRate,
		L2:           cfg.L2,
		AvgAbsError:  avgAbsErr,
		WeightsPath:  weightsPath,
		ModelName:    model.Name,
	}, nil
}

func targetForExample(ex Example) float64 {
	target := ex.OutcomeValue
	if ex.HasMoveTarget {
		target = 0.72*ex.OutcomeValue + 0.28*ex.MoveTarget
	}
	return clamp(target, -1, 1)
}

func loadExamples(path string) ([]Example, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	examples := make([]Example, 0, 1024)
	for scanner.Scan() {
		var ex Example
		if err := json.Unmarshal(scanner.Bytes(), &ex); err != nil {
			return nil, err
		}
		examples = append(examples, ex)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return examples, nil
}

func loadWeights(path string) (linearWeightsFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return linearWeightsFile{}, err
	}
	var model linearWeightsFile
	if err := json.Unmarshal(raw, &model); err != nil {
		return linearWeightsFile{}, err
	}
	return model, nil
}

func saveWeights(path string, model linearWeightsFile) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(model)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
