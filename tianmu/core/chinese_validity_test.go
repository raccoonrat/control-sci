package core

import (
	"testing"
	"time"
)

func TestChineseValidityMultiTurnIncrementalDrift(t *testing.T) {
	tracker := NewSessionTracker(4)
	sessionID := "session-user-test-01"

	tracker.RecordTurn(sessionID, RiskEvaluation{
		RiskCategories: []string{"alignment_drift"},
		MaxRiskScore:   0.65,
	})
	tracker.RecordTurn(sessionID, RiskEvaluation{
		RiskCategories: []string{"alignment_drift"},
		MaxRiskScore:   0.70,
	})
	tracker.RecordTurn(sessionID, RiskEvaluation{
		RiskCategories: []string{"alignment_drift"},
		MaxRiskScore:   0.72,
	})

	if !tracker.EvaluateCumulativeRisk(sessionID) {
		t.Fatal("multi-turn cumulative risk gate did not trigger")
	}
}

func TestChineseValiditySessionWindowBoundsHistory(t *testing.T) {
	tracker := NewSessionTracker(2)
	sessionID := "session-user-test-02"

	tracker.RecordTurn(sessionID, RiskEvaluation{RiskCategories: []string{"benign"}, MaxRiskScore: 0.10})
	tracker.RecordTurn(sessionID, RiskEvaluation{RiskCategories: []string{"alignment_drift"}, MaxRiskScore: 0.65})
	tracker.RecordTurn(sessionID, RiskEvaluation{RiskCategories: []string{"alignment_drift"}, MaxRiskScore: 0.70})

	history, ok := tracker.Snapshot(sessionID)
	if !ok {
		t.Fatal("session history must exist")
	}
	if len(history.Turns) != 2 {
		t.Fatalf("history turns = %d, want 2", len(history.Turns))
	}
	if history.Turns[0].MaxRiskScore != 0.65 {
		t.Fatalf("oldest retained score = %v, want 0.65", history.Turns[0].MaxRiskScore)
	}
}

func TestChineseValiditySessionTrackerEvictsExpiredSessions(t *testing.T) {
	tracker := NewSessionTrackerWithTTL(4, time.Minute)
	tracker.sessions["expired-session"] = &SessionHistory{
		SessionID:  "expired-session",
		CreatedAt:  time.Now().UTC().Add(-2 * time.Hour),
		LastActive: time.Now().UTC().Add(-2 * time.Hour),
	}

	tracker.RecordTurn("active-session", RiskEvaluation{RiskCategories: []string{"alignment_drift"}, MaxRiskScore: 0.65})

	if _, ok := tracker.Snapshot("expired-session"); ok {
		t.Fatal("expired session must be evicted")
	}
	if _, ok := tracker.Snapshot("active-session"); !ok {
		t.Fatal("active session must exist")
	}
}

func TestChineseValidityEngineUsesSessionTrackerForCumulativeRisk(t *testing.T) {
	evaluator, err := NewEvaluator(PolicyPack{Version: "session-risk-policy-v1"})
	if err != nil {
		t.Fatalf("new evaluator: %v", err)
	}
	engine, err := NewEngine(PersonalAI, evaluator)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.AttachSessionTracker(NewSessionTracker(4))

	req := RequestContext{ProductID: "Qira", Language: "zh-CN", InteractionType: "user_io"}
	identity := IdentityContext{ActorID: "session-user-test-03"}
	data := DataContext{DataClassification: "public", Source: "user_input", Destination: "model_context"}
	action := ActionContext{ActionType: "generate_response"}
	signals := []DetectorSignal{
		{
			DetectorID: "alignment-drift-fastpath",
			Category:   "alignment_drift",
			Version:    "alignment-drift-v1",
			Confidence: 0.65,
			Triggered:  true,
		},
	}

	for i := 0; i < 2; i++ {
		decision, err := engine.MediateInbound(req, identity, data, action, signals)
		if err != nil {
			t.Fatalf("mediate inbound: %v", err)
		}
		if decision.PolicyDecision.Decision != Allow {
			t.Fatalf("decision before cumulative threshold = %q, want %q", decision.PolicyDecision.Decision, Allow)
		}
	}

	decision, err := engine.MediateInbound(req, identity, data, action, signals)
	if err != nil {
		t.Fatalf("mediate inbound: %v", err)
	}
	if decision.PolicyDecision.Decision != AskConfirmation {
		t.Fatalf("decision after cumulative threshold = %q, want %q", decision.PolicyDecision.Decision, AskConfirmation)
	}
	if decision.PolicyDecision.ReasonCode != "cumulative_session_risk_requires_confirmation" {
		t.Fatalf("reason = %q, want cumulative_session_risk_requires_confirmation", decision.PolicyDecision.ReasonCode)
	}
}

func BenchmarkSessionTrackerRecordAndEvaluate(b *testing.B) {
	tracker := NewSessionTracker(4)
	sessionID := "benchmark-session"
	risk := RiskEvaluation{
		RiskCategories: []string{"alignment_drift"},
		MaxRiskScore:   0.65,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.RecordTurn(sessionID, risk)
		if i%3 == 0 {
			_ = tracker.EvaluateCumulativeRisk(sessionID)
		}
	}
}
