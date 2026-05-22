package regression

import (
	"path/filepath"
	"time"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

type TC260Report struct {
	ReportType      string                `json:"report_type"`
	GeneratedAt     time.Time             `json:"generated_at"`
	Dataset         TC260DatasetEvidence  `json:"dataset"`
	Summary         TC260Summary          `json:"summary"`
	Expected        map[string]int        `json:"expected"`
	ByCategory      map[string]int        `json:"by_category"`
	ByDifficulty    map[string]int        `json:"by_difficulty"`
	BySource        map[string]int        `json:"by_source"`
	DecisionCounts  map[core.Decision]int `json:"decision_counts"`
	Matrix          *ConfusionMatrix      `json:"confusion_matrix,omitempty"`
	Metrics         *DerivedMetrics       `json:"metrics,omitempty"`
	Profiler        *Profiler             `json:"profiler,omitempty"`
	FailureExamples []TC260Failure        `json:"failure_examples,omitempty"`
}

type TC260DatasetEvidence struct {
	Name                  string        `json:"name"`
	Version               string        `json:"version"`
	FileName              string        `json:"file_name"`
	ManifestFormatVersion string        `json:"manifest_format_version,omitempty"`
	DatasetLayoutVersion  string        `json:"dataset_layout_version,omitempty"`
	CreatedAt             string        `json:"created_at,omitempty"`
	File                  *ManifestFile `json:"file,omitempty"`
}

type TC260Failure struct {
	ID               string        `json:"id"`
	TC260Category    string        `json:"tc260_category,omitempty"`
	ExpectedBehavior string        `json:"expected_behavior"`
	Decision         core.Decision `json:"decision"`
	ReasonCode       string        `json:"reason_code"`
	Source           string        `json:"source,omitempty"`
	Difficulty       string        `json:"difficulty,omitempty"`
	Failure          string        `json:"failure"`
}

func BuildTC260Report(datasetPath string, manifest *DatasetManifest, results []TC260Result, summary TC260Summary) TC260Report {
	fileName := filepath.Base(datasetPath)
	report := TC260Report{
		ReportType:      "tc260_release_evidence",
		GeneratedAt:     time.Now().UTC(),
		Dataset:         buildDatasetEvidence(datasetPath, fileName, manifest),
		Summary:         summary,
		Expected:        map[string]int{},
		ByCategory:      map[string]int{},
		ByDifficulty:    map[string]int{},
		BySource:        map[string]int{},
		DecisionCounts:  map[core.Decision]int{},
		FailureExamples: make([]TC260Failure, 0),
	}

	for _, result := range results {
		report.Expected[result.Case.ExpectedBehavior]++
		report.ByCategory[categoryOrDefault(result.Case.TC260Category)]++
		report.ByDifficulty[difficultyOrDefault(result.Case)]++
		report.BySource[sourceOrDefault(result.Case.Source)]++
		report.DecisionCounts[result.Decision]++

		if !result.Passed {
			report.FailureExamples = append(report.FailureExamples, TC260Failure{
				ID:               result.Case.ID,
				TC260Category:    result.Case.TC260Category,
				ExpectedBehavior: result.Case.ExpectedBehavior,
				Decision:         result.Decision,
				ReasonCode:       result.ReasonCode,
				Source:           result.Case.Source,
				Difficulty:       difficultyOrDefault(result.Case),
				Failure:          result.Failure,
			})
		}
	}

	return report
}

func BuildTC260QualityReport(
	datasetPath string,
	manifest *DatasetManifest,
	results []TC260Result,
	summary TC260Summary,
	matrix ConfusionMatrix,
	profiler *Profiler,
) TC260Report {
	report := BuildTC260Report(datasetPath, manifest, results, summary)
	metrics := matrix.CalculateMetrics()
	report.Matrix = &matrix
	report.Metrics = &metrics
	report.Profiler = profiler
	return report
}

func buildDatasetEvidence(datasetPath string, fileName string, manifest *DatasetManifest) TC260DatasetEvidence {
	evidence := TC260DatasetEvidence{
		Name:     "tc260",
		Version:  filepath.Base(filepath.Dir(datasetPath)),
		FileName: fileName,
	}
	if manifest == nil {
		return evidence
	}

	evidence.ManifestFormatVersion = manifest.ManifestFormatVersion
	evidence.DatasetLayoutVersion = manifest.DatasetLayoutVersion
	evidence.CreatedAt = manifest.CreatedAt
	if file, ok := manifest.FindFile(fileName); ok {
		evidence.File = &file
	}

	return evidence
}

func categoryOrDefault(category string) string {
	if category == "" {
		return "uncategorized"
	}

	return category
}

func difficultyOrDefault(testCase TC260Case) string {
	difficulty, _ := testCase.Attributes["difficulty"].(string)
	if difficulty == "" {
		return "unknown"
	}

	return difficulty
}

func sourceOrDefault(source string) string {
	if source == "" {
		return "unknown"
	}

	return source
}
