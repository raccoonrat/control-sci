package core

import (
	"sync"
	"time"
)

const defaultSessionWindow = 4

type MessageTurn struct {
	Timestamp      time.Time `json:"timestamp"`
	MaxRiskScore   float64   `json:"max_risk_score"`
	TriggeredRisks []string  `json:"triggered_risks"`
}

type SessionHistory struct {
	SessionID  string        `json:"session_id"`
	Turns      []MessageTurn `json:"turns"`
	CreatedAt  time.Time     `json:"created_at"`
	LastActive time.Time     `json:"last_active"`
}

type SessionTracker struct {
	mu         sync.RWMutex
	sessions   map[string]*SessionHistory
	windowSize int
}

func NewSessionTracker(windowSize int) *SessionTracker {
	if windowSize <= 0 {
		windowSize = defaultSessionWindow
	}

	return &SessionTracker{
		sessions:   make(map[string]*SessionHistory),
		windowSize: windowSize,
	}
}

func (t *SessionTracker) RecordTurn(sessionID string, risk RiskEvaluation) {
	if sessionID == "" {
		return
	}

	now := time.Now().UTC()
	turn := MessageTurn{
		Timestamp:      now,
		MaxRiskScore:   risk.MaxRiskScore,
		TriggeredRisks: append([]string(nil), risk.RiskCategories...),
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	history, ok := t.sessions[sessionID]
	if !ok {
		history = &SessionHistory{
			SessionID:  sessionID,
			CreatedAt:  now,
			LastActive: now,
		}
		t.sessions[sessionID] = history
	}

	history.LastActive = now
	history.Turns = append(history.Turns, turn)
	if len(history.Turns) > t.windowSize {
		history.Turns = history.Turns[len(history.Turns)-t.windowSize:]
	}
}

func (t *SessionTracker) EvaluateCumulativeRisk(sessionID string) bool {
	return t.EvaluateCumulativeRiskWithThreshold(sessionID, 3, 0.50, 0.80)
}

func (t *SessionTracker) EvaluateCumulativeRiskWithThreshold(sessionID string, minTurns int, low float64, high float64) bool {
	if sessionID == "" || minTurns <= 0 || low > high {
		return false
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	history, ok := t.sessions[sessionID]
	if !ok {
		return false
	}

	count := 0
	for _, turn := range history.Turns {
		if turn.MaxRiskScore >= low && turn.MaxRiskScore < high {
			count++
		}
	}

	return count >= minTurns
}

func (t *SessionTracker) Snapshot(sessionID string) (SessionHistory, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	history, ok := t.sessions[sessionID]
	if !ok {
		return SessionHistory{}, false
	}

	copyHistory := *history
	copyHistory.Turns = append([]MessageTurn(nil), history.Turns...)
	return copyHistory, true
}
