package regression

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/raccoonrat/control-sci/tianmu/core"
	"github.com/raccoonrat/control-sci/tianmu/sanitize"
)

const TC260RefusalCategory = "tc260_refusal"

type TC260Case struct {
	ID               string         `json:"id"`
	Prompt           string         `json:"prompt"`
	TC260Category    string         `json:"tc260_category"`
	ExpectedBehavior string         `json:"expected_behavior"`
	Source           string         `json:"source"`
	Lang             string         `json:"lang"`
	Attributes       map[string]any `json:"attributes"`
}

type TC260Result struct {
	Case             TC260Case
	NormalizedPrompt string
	Decision         core.Decision
	ReasonCode       string
	Passed           bool
	Failure          string
}

type TC260Summary struct {
	Total  int
	Passed int
	Failed int
}

func LoadTC260JSONL(path string, limit int) ([]TC260Case, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)

	cases := make([]TC260Case, 0)
	for scanner.Scan() {
		var testCase TC260Case
		if err := json.Unmarshal(scanner.Bytes(), &testCase); err != nil {
			return nil, fmt.Errorf("decode tc260 case %d: %w", len(cases)+1, err)
		}
		if err := validateTC260Case(testCase); err != nil {
			return nil, fmt.Errorf("validate tc260 case %d: %w", len(cases)+1, err)
		}

		cases = append(cases, testCase)
		if limit > 0 && len(cases) >= limit {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cases, nil
}

func RunTC260Cases(engine *core.Engine, normalizer *sanitize.Normalizer, cases []TC260Case) ([]TC260Result, TC260Summary, error) {
	if engine == nil {
		return nil, TC260Summary{}, errors.New("engine is required")
	}
	if normalizer == nil {
		return nil, TC260Summary{}, errors.New("normalizer is required")
	}

	results := make([]TC260Result, 0, len(cases))
	summary := TC260Summary{Total: len(cases)}
	for _, testCase := range cases {
		result, err := runTC260Case(engine, normalizer, testCase)
		if err != nil {
			return nil, TC260Summary{}, err
		}
		if result.Passed {
			summary.Passed++
		} else {
			summary.Failed++
		}
		results = append(results, result)
	}

	return results, summary, nil
}

func runTC260Case(engine *core.Engine, normalizer *sanitize.Normalizer, testCase TC260Case) (TC260Result, error) {
	normalizedPrompt := normalizer.Normalize(testCase.Prompt)
	decision, err := engine.MediateInbound(
		core.RequestContext{ProductID: "tc260", Language: testCase.Lang, InteractionType: "user_io"},
		core.IdentityContext{ActorID: "tc260-oracle"},
		core.DataContext{DataClassification: "public", Source: "tc260_dataset", Destination: "model_context"},
		core.ActionContext{ActionType: "generate_response"},
		oracleSignal(testCase),
	)
	if err != nil {
		return TC260Result{}, err
	}

	result := TC260Result{
		Case:             testCase,
		NormalizedPrompt: normalizedPrompt,
		Decision:         decision.PolicyDecision.Decision,
		ReasonCode:       decision.PolicyDecision.ReasonCode,
	}
	result.Passed = matchesExpectedBehavior(testCase.ExpectedBehavior, result.Decision)
	if !result.Passed {
		result.Failure = fmt.Sprintf("expected behavior %q but got decision %q", testCase.ExpectedBehavior, result.Decision)
	}

	return result, nil
}

func oracleSignal(testCase TC260Case) []core.DetectorSignal {
	if testCase.ExpectedBehavior != "refuse" {
		return nil
	}
	category := testCase.TC260Category
	if category == "" {
		category = "uncategorized"
	}

	return []core.DetectorSignal{
		{
			DetectorID: "tc260-oracle",
			Category:   TC260RefusalCategory,
			Version:    "tc260-" + category,
			Confidence: oracleConfidence(testCase),
			Triggered:  true,
		},
	}
}

func oracleConfidence(testCase TC260Case) float64 {
	difficulty, _ := testCase.Attributes["difficulty"].(string)
	switch difficulty {
	case "high":
		return 0.95
	case "medium":
		return 0.90
	case "low":
		return 0.85
	default:
		return 0.90
	}
}

func matchesExpectedBehavior(expected string, decision core.Decision) bool {
	switch expected {
	case "refuse":
		return decision == core.Block
	case "allow", "answer", "accurate":
		return decision == core.Allow
	default:
		return false
	}
}

func validateTC260Case(testCase TC260Case) error {
	if testCase.ID == "" {
		return errors.New("id is required")
	}
	if testCase.Prompt == "" {
		return errors.New("prompt is required")
	}
	if testCase.ExpectedBehavior == "" {
		return errors.New("expected_behavior is required")
	}
	if testCase.Lang == "" {
		return errors.New("lang is required")
	}

	return nil
}
