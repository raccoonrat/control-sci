package regression

import "github.com/raccoonrat/control-sci/tianmu/core"

type CategoryProfile struct {
	Category      string  `json:"category"`
	TotalCount    int     `json:"total_count"`
	FailureCount  int     `json:"failure_count"`
	TruePositive  int     `json:"true_positive"`
	FalsePositive int     `json:"false_positive"`
	TrueNegative  int     `json:"true_negative"`
	FalseNegative int     `json:"false_negative"`
	Recall        float64 `json:"recall"`
}

type DifficultyProfile struct {
	Difficulty string `json:"difficulty"`
	Passed     int    `json:"passed"`
	Failed     int    `json:"failed"`
}

type Profiler struct {
	CategoryStats   map[string]*CategoryProfile   `json:"category_stats"`
	DifficultyStats map[string]*DifficultyProfile `json:"difficulty_stats"`
}

func NewProfiler() *Profiler {
	return &Profiler{
		CategoryStats:   make(map[string]*CategoryProfile),
		DifficultyStats: make(map[string]*DifficultyProfile),
	}
}

func (p *Profiler) ProfileCase(testCase TC260Case, decision core.Decision, passed bool) {
	category := categoryOrDefault(testCase.TC260Category)
	if category == "uncategorized" && testCase.ExpectedBehavior != "refuse" {
		category = "positive_control"
	}
	if _, ok := p.CategoryStats[category]; !ok {
		p.CategoryStats[category] = &CategoryProfile{Category: category}
	}
	categoryProfile := p.CategoryStats[category]
	categoryProfile.TotalCount++
	if !passed {
		categoryProfile.FailureCount++
	}
	_, quadrant := ClassifyDecision(testCase.ExpectedBehavior, decision)
	switch quadrant {
	case "TP":
		categoryProfile.TruePositive++
	case "FP":
		categoryProfile.FalsePositive++
	case "TN":
		categoryProfile.TrueNegative++
	case "FN":
		categoryProfile.FalseNegative++
	}
	if positives := categoryProfile.TruePositive + categoryProfile.FalseNegative; positives > 0 {
		categoryProfile.Recall = float64(categoryProfile.TruePositive) / float64(positives)
	}

	difficulty := difficultyOrDefault(testCase)
	if _, ok := p.DifficultyStats[difficulty]; !ok {
		p.DifficultyStats[difficulty] = &DifficultyProfile{Difficulty: difficulty}
	}
	if passed {
		p.DifficultyStats[difficulty].Passed++
	} else {
		p.DifficultyStats[difficulty].Failed++
	}
}
