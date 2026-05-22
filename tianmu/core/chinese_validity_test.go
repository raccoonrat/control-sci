package core

import "testing"

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
