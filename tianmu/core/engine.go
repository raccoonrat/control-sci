package core

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type Engine struct {
	evaluator      *Evaluator
	releaseStage   ReleaseStage
	sessionTracker *SessionTracker
}

func NewEngine(stage ReleaseStage, evaluator *Evaluator) (*Engine, error) {
	if stage == "" {
		return nil, errors.New("release stage is required")
	}
	if evaluator == nil {
		return nil, errors.New("evaluator is required")
	}

	return &Engine{
		evaluator:    evaluator,
		releaseStage: stage,
	}, nil
}

func (e *Engine) AttachSessionTracker(tracker *SessionTracker) {
	e.sessionTracker = tracker
}

func (e *Engine) MediateInbound(
	req RequestContext,
	identity IdentityContext,
	data DataContext,
	action ActionContext,
	signals []DetectorSignal,
) (*ControlDecisionObject, error) {
	now := time.Now().UTC()
	risk := summarizeRisk(signals)
	decision := e.evaluator.EvaluateAction(risk, action)
	decision = e.applySessionRisk(identity.ActorID, risk, decision)

	traceID, err := newTraceID()
	if err != nil {
		return nil, err
	}

	return &ControlDecisionObject{
		ControlID:       fmt.Sprintf("tmctl-%d", now.UnixNano()),
		Timestamp:       now,
		ReleaseStage:    e.releaseStage,
		RequestContext:  req,
		IdentityContext: identity,
		DataContext:     data,
		ActionContext:   action,
		RiskEvaluation:  risk,
		PolicyDecision:  decision,
		ReleaseEvidence: ReleaseEvidenceLite{
			EvidenceLevel:     "release_evidence_lite",
			TraceID:           traceID,
			Timestamp:         now,
			RegressionPassTag: true,
		},
	}, nil
}

func (e *Engine) applySessionRisk(sessionID string, risk RiskEvaluation, decision PolicyDecision) PolicyDecision {
	if e.sessionTracker == nil || sessionID == "" {
		return decision
	}

	e.sessionTracker.RecordTurn(sessionID, risk)
	if decision.Decision != Allow || !e.sessionTracker.EvaluateCumulativeRisk(sessionID) {
		return decision
	}

	return PolicyDecision{
		Decision:          AskConfirmation,
		PolicyPackVersion: decision.PolicyPackVersion,
		ReasonCode:        "cumulative_session_risk_requires_confirmation",
	}
}

func summarizeRisk(signals []DetectorSignal) RiskEvaluation {
	categories := make([]string, 0, len(signals))
	versions := make([]string, 0, len(signals))
	seenCategories := map[string]struct{}{}
	seenVersions := map[string]struct{}{}
	maxScore := 0.0

	for _, signal := range signals {
		if signal.Version != "" {
			if _, ok := seenVersions[signal.Version]; !ok {
				versions = append(versions, signal.Version)
				seenVersions[signal.Version] = struct{}{}
			}
		}
		if !signal.Triggered {
			continue
		}
		if signal.Category != "" {
			if _, ok := seenCategories[signal.Category]; !ok {
				categories = append(categories, signal.Category)
				seenCategories[signal.Category] = struct{}{}
			}
		}
		if signal.Confidence > maxScore {
			maxScore = signal.Confidence
		}
	}

	return RiskEvaluation{
		RiskCategories:   categories,
		MaxRiskScore:     maxScore,
		DetectorVersions: versions,
		Signals:          signals,
	}
}

func newTraceID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate trace id: %w", err)
	}

	return hex.EncodeToString(buf[:]), nil
}
