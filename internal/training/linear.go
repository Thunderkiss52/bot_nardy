package training

import (
	"bufio"
	"encoding/json"
	"errors"
	"hash/fnv"
	"math"
	"math/rand"
	"os"
)

type TrainConfig struct {
	Epochs          int     `json:"epochs"`
	LearningRate    float64 `json:"learning_rate"`
	L2              float64 `json:"l2"`
	MaxExamples     int     `json:"max_examples"`
	ValidationSplit float64 `json:"validation_split"`
	Patience        int     `json:"patience"`
	ShuffleSeed     int64   `json:"shuffle_seed"`
}

type TrainSummary struct {
	Examples           int     `json:"examples"`
	TrainExamples      int     `json:"train_examples"`
	ValidationExamples int     `json:"validation_examples"`
	Epochs             int     `json:"epochs"`
	LearningRate       float64 `json:"learning_rate"`
	L2                 float64 `json:"l2"`
	AvgAbsError        float64 `json:"avg_abs_error"`
	ValidationAbsError float64 `json:"validation_abs_error"`
	BestEpoch          int     `json:"best_epoch"`
	WeightsPath        string  `json:"weights_path"`
	ModelName          string  `json:"model_name"`
	Accepted           bool    `json:"accepted"`
	ValidationGames    int     `json:"validation_games"`
	ValidationWinRate  float64 `json:"validation_win_rate"`
	LeagueSize         int     `json:"league_size"`
	ChampionModel      string  `json:"champion_model"`
	ChampionScore      float64 `json:"champion_score"`
}

type linearWeightsFile struct {
	Name    string             `json:"name"`
	Bias    float64            `json:"bias"`
	Weights map[string]float64 `json:"weights"`
	Scales  map[string]float64 `json:"scales,omitempty"`
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
	if cfg.ValidationSplit <= 0 || cfg.ValidationSplit >= 0.5 {
		cfg.ValidationSplit = 0.12
	}
	if cfg.Patience <= 0 {
		cfg.Patience = 4
	}
	if cfg.ShuffleSeed == 0 {
		cfg.ShuffleSeed = 1
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
	trainExamples, validationExamples := splitExamples(examples, cfg.ValidationSplit)
	if len(trainExamples) == 0 {
		trainExamples = examples
		validationExamples = nil
	}
	stats := fitFeatureScales(trainExamples)
	trainRows := buildRows(trainExamples, stats)
	validationRows := buildRows(validationExamples, stats)

	model := linearWeightsFile{
		Name:    "linear-human-v4",
		Weights: map[string]float64{},
		Scales:  cloneFloatMap(stats),
	}
	if loaded, err := loadWeights(weightsPath); err == nil {
		model = adaptModelToScales(loaded, stats)
	}
	if model.Weights == nil {
		model.Weights = map[string]float64{}
	}
	if model.Name == "" {
		model.Name = "linear-human-v4"
	}
	if model.Scales == nil {
		model.Scales = cloneFloatMap(stats)
	}

	rng := rand.New(rand.NewSource(cfg.ShuffleSeed))
	bestModel := cloneLinearModel(model)
	bestEpoch := 1
	bestMetric := math.Inf(1)
	staleEpochs := 0
	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		rng.Shuffle(len(trainRows), func(i, j int) {
			trainRows[i], trainRows[j] = trainRows[j], trainRows[i]
		})
		for _, row := range trainRows {
			pred := predictRow(model, row.values)
			err := pred - row.target

			model.Bias -= cfg.LearningRate * err
			for key, value := range row.values {
				w := model.Weights[key]
				model.Weights[key] = w - cfg.LearningRate*(err*value+cfg.L2*w)
			}
		}

		trainErr := meanAbsError(model, trainRows)
		validationErr := meanAbsError(model, validationRows)
		metric := validationErr
		if len(validationRows) == 0 {
			metric = trainErr
		}
		if metric < bestMetric-1e-6 {
			bestMetric = metric
			bestModel = cloneLinearModel(model)
			bestEpoch = epoch + 1
			staleEpochs = 0
			continue
		}
		staleEpochs++
		if staleEpochs >= cfg.Patience {
			break
		}
	}
	model = bestModel

	if err := saveWeights(weightsPath, model); err != nil {
		return TrainSummary{}, err
	}

	avgAbsErr := meanAbsError(model, trainRows)
	validationAbsErr := meanAbsError(model, validationRows)
	return TrainSummary{
		Examples:           len(examples),
		TrainExamples:      len(trainRows),
		ValidationExamples: len(validationRows),
		Epochs:             cfg.Epochs,
		LearningRate:       cfg.LearningRate,
		L2:                 cfg.L2,
		AvgAbsError:        avgAbsErr,
		ValidationAbsError: validationAbsErr,
		BestEpoch:          bestEpoch,
		WeightsPath:        weightsPath,
		ModelName:          model.Name,
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

type trainingRow struct {
	values map[string]float64
	target float64
}

func splitExamples(examples []Example, validationSplit float64) ([]Example, []Example) {
	if len(examples) < 20 || validationSplit <= 0 {
		return examples, nil
	}
	train := make([]Example, 0, len(examples))
	validation := make([]Example, 0, len(examples)/8)
	threshold := uint64(validationSplit * 10_000)
	for idx, ex := range examples {
		if hashExampleForSplit(ex, idx)%10_000 < threshold {
			validation = append(validation, ex)
			continue
		}
		train = append(train, ex)
	}
	if len(train) == 0 || len(validation) == 0 {
		return examples, nil
	}
	return train, validation
}

func hashExampleForSplit(ex Example, idx int) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(ex.StateKey))
	_, _ = h.Write([]byte(ex.Turn))
	_, _ = h.Write([]byte{byte(ex.MoveNumber >> 8), byte(ex.MoveNumber)})
	if ex.StateKey == "" && ex.Turn == "" && ex.MoveNumber == 0 {
		_, _ = h.Write([]byte{byte(idx >> 24), byte(idx >> 16), byte(idx >> 8), byte(idx)})
	}
	return h.Sum64()
}

func fitFeatureScales(examples []Example) map[string]float64 {
	type agg struct {
		sumSq float64
	}
	acc := map[string]agg{}
	n := float64(len(examples))
	if n == 0 {
		return map[string]float64{}
	}
	for _, ex := range examples {
		for key, value := range ex.Features.Values() {
			a := acc[key]
			a.sumSq += value * value
			acc[key] = a
		}
	}
	scales := make(map[string]float64, len(acc))
	for key, a := range acc {
		scale := math.Sqrt(a.sumSq / n)
		if scale < 1e-6 {
			scale = 1.0
		}
		scales[key] = scale
	}
	return scales
}

func buildRows(examples []Example, scales map[string]float64) []trainingRow {
	rows := make([]trainingRow, 0, len(examples))
	for _, ex := range examples {
		values := ex.Features.Values()
		norm := make(map[string]float64, len(values))
		for key, value := range values {
			scale := scales[key]
			if scale == 0 {
				scale = 1.0
			}
			norm[key] = value / scale
		}
		rows = append(rows, trainingRow{
			values: norm,
			target: targetForExample(ex),
		})
	}
	return rows
}

func adaptModelToScales(model linearWeightsFile, newScales map[string]float64) linearWeightsFile {
	if model.Weights == nil {
		model.Weights = map[string]float64{}
	}
	oldScales := model.Scales
	if oldScales == nil {
		oldScales = map[string]float64{}
	}
	adjusted := linearWeightsFile{
		Name:    model.Name,
		Bias:    model.Bias,
		Weights: make(map[string]float64, len(model.Weights)),
		Scales:  cloneFloatMap(newScales),
	}
	for key, weight := range model.Weights {
		oldScale := oldScales[key]
		if oldScale == 0 {
			oldScale = 1.0
		}
		newScale := newScales[key]
		if newScale == 0 {
			newScale = 1.0
		}
		adjusted.Weights[key] = weight * newScale / oldScale
	}
	return adjusted
}

func predictRow(model linearWeightsFile, values map[string]float64) float64 {
	pred := model.Bias
	for key, value := range values {
		pred += model.Weights[key] * value
	}
	return pred
}

func meanAbsError(model linearWeightsFile, rows []trainingRow) float64 {
	if len(rows) == 0 {
		return 0
	}
	total := 0.0
	for _, row := range rows {
		total += abs(predictRow(model, row.values) - row.target)
	}
	return total / float64(len(rows))
}

func cloneLinearModel(model linearWeightsFile) linearWeightsFile {
	return linearWeightsFile{
		Name:    model.Name,
		Bias:    model.Bias,
		Weights: cloneFloatMap(model.Weights),
		Scales:  cloneFloatMap(model.Scales),
	}
}

func cloneFloatMap(src map[string]float64) map[string]float64 {
	if src == nil {
		return nil
	}
	dst := make(map[string]float64, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
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
