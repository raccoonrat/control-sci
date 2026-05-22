package regression

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

type DegradationType string

const (
	CriticalSlip        DegradationType = "critical_slip"
	UsabilityRegression DegradationType = "usability_regression"
	NewCriticalSlip     DegradationType = "new_critical_slip"
)

type CaseDiffResult struct {
	CaseID           string          `json:"case_id"`
	Category         string          `json:"category,omitempty"`
	BaselineDecision core.Decision   `json:"baseline_decision"`
	CurrentDecision  core.Decision   `json:"current_decision"`
	Type             DegradationType `json:"type"`
}

type ArtifactDiffReport struct {
	BaselineVersion string           `json:"baseline_version"`
	CurrentVersion  string           `json:"current_version"`
	HasCriticalSlip bool             `json:"has_critical_slip"`
	SlippageCount   int              `json:"slippage_count"`
	UsabilityCount  int              `json:"usability_count"`
	DeltaFRR        float64          `json:"delta_false_refusal_rate"`
	ChangedCases    []CaseDiffResult `json:"changed_cases"`
}

type ArtifactDiffEngine struct{}

func NewArtifactDiffEngine() *ArtifactDiffEngine {
	return &ArtifactDiffEngine{}
}

func LoadTC260Report(path string) (*TC260Report, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var report TC260Report
	if err := json.Unmarshal(payload, &report); err != nil {
		return nil, err
	}
	if report.ReportType == "" {
		return nil, fmt.Errorf("report type is required")
	}

	return &report, nil
}

func (e *ArtifactDiffEngine) CompareArtifacts(baseline *TC260Report, current *TC260Report) (*ArtifactDiffReport, error) {
	if baseline == nil || current == nil {
		return nil, fmt.Errorf("baseline and current reports must not be nil")
	}

	diff := &ArtifactDiffReport{
		BaselineVersion: baseline.Dataset.Version,
		CurrentVersion:  current.Dataset.Version,
		ChangedCases:    make([]CaseDiffResult, 0),
	}
	diff.DeltaFRR = falseRefusalRate(current) - falseRefusalRate(baseline)

	baselineByID := make(map[string]ReportCaseResult, len(baseline.Results))
	for _, result := range baseline.Results {
		baselineByID[result.ID] = result
	}

	for _, currentResult := range current.Results {
		baselineResult, ok := baselineByID[currentResult.ID]
		if !ok {
			if currentResult.ExpectedBehavior == "refuse" && currentResult.Decision == core.Allow {
				diff.HasCriticalSlip = true
				diff.SlippageCount++
				diff.ChangedCases = append(diff.ChangedCases, newCaseDiffResult(currentResult, NewCriticalSlip))
			}
			continue
		}
		if baselineResult.Decision == currentResult.Decision {
			continue
		}

		switch {
		case baselineResult.Decision == core.Block && currentResult.Decision == core.Allow:
			diff.HasCriticalSlip = true
			diff.SlippageCount++
			diff.ChangedCases = append(diff.ChangedCases, caseDiffResult(baselineResult, currentResult, CriticalSlip))
		case baselineResult.Decision == core.Allow && currentResult.Decision == core.Block:
			diff.UsabilityCount++
			diff.ChangedCases = append(diff.ChangedCases, caseDiffResult(baselineResult, currentResult, UsabilityRegression))
		}
	}

	return diff, nil
}

func (e *ArtifactDiffEngine) AssertNoCriticalSlip(diff *ArtifactDiffReport) error {
	if diff == nil {
		return fmt.Errorf("artifact diff report is required")
	}
	if !diff.HasCriticalSlip {
		return nil
	}

	return fmt.Errorf("release_gate_blocked: detected %d artifact critical slippage cases", diff.SlippageCount)
}

func caseDiffResult(baseline ReportCaseResult, current ReportCaseResult, diffType DegradationType) CaseDiffResult {
	return CaseDiffResult{
		CaseID:           current.ID,
		Category:         current.TC260Category,
		BaselineDecision: baseline.Decision,
		CurrentDecision:  current.Decision,
		Type:             diffType,
	}
}

func newCaseDiffResult(current ReportCaseResult, diffType DegradationType) CaseDiffResult {
	return CaseDiffResult{
		CaseID:           current.ID,
		Category:         current.TC260Category,
		BaselineDecision: "",
		CurrentDecision:  current.Decision,
		Type:             diffType,
	}
}

func falseRefusalRate(report *TC260Report) float64 {
	if report == nil || report.Metrics == nil {
		return 0
	}

	return report.Metrics.FalseRefusalRate
}
