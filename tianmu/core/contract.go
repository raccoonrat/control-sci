package core

import "time"

type ReleaseStage string

const (
	PersonalAI   ReleaseStage = "personal_ai"
	EnterpriseAI ReleaseStage = "enterprise_ai"
)

type Decision string

const (
	Allow           Decision = "allow"
	Block           Decision = "block"
	RedactThenAllow Decision = "redact_then_allow"
	Rewrite         Decision = "rewrite"
	AskConfirmation Decision = "ask_confirmation"
	Escalate        Decision = "escalate"
	LogOnly         Decision = "log_only"
)

const (
	ActionCallTool      = "call_tool"
	ActionProcessOutput = "process_tool_output"
)

type RequestContext struct {
	ProductID       string `json:"product_id"`
	Language        string `json:"language"`
	InteractionType string `json:"interaction_type"`
}

type IdentityContext struct {
	ActorID    string            `json:"actor_id"`
	TenantID   string            `json:"tenant_id,omitempty"`
	Role       string            `json:"role,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type DataContext struct {
	DataClassification string `json:"data_classification"`
	ContainsPII        bool   `json:"contains_pii"`
	Source             string `json:"source"`
	Destination        string `json:"destination"`
}

type ActionContext struct {
	ActionType string `json:"action_type"`
	ToolName   string `json:"tool_name,omitempty"`
	SideEffect bool   `json:"side_effect"`
}

type DetectorSignal struct {
	DetectorID string  `json:"detector_id"`
	Category   string  `json:"category"`
	Version    string  `json:"version"`
	Confidence float64 `json:"confidence"`
	Triggered  bool    `json:"triggered"`
}

type RiskEvaluation struct {
	RiskCategories   []string         `json:"risk_categories"`
	MaxRiskScore     float64          `json:"max_risk_score"`
	DetectorVersions []string         `json:"detector_versions"`
	Signals          []DetectorSignal `json:"signals"`
}

type PolicyDecision struct {
	Decision          Decision `json:"decision"`
	PolicyPackVersion string   `json:"policy_pack_version"`
	ReasonCode        string   `json:"reason_code"`
}

type ReleaseEvidenceLite struct {
	EvidenceLevel     string    `json:"evidence_level"`
	TraceID           string    `json:"trace_id"`
	Timestamp         time.Time `json:"timestamp"`
	RegressionPassTag bool      `json:"regression_pass_tag"`
}

type ControlDecisionObject struct {
	ControlID       string              `json:"control_id"`
	Timestamp       time.Time           `json:"timestamp"`
	ReleaseStage    ReleaseStage        `json:"release_stage"`
	RequestContext  RequestContext      `json:"request_context"`
	IdentityContext IdentityContext     `json:"identity_context"`
	DataContext     DataContext         `json:"data_context"`
	ActionContext   ActionContext       `json:"action_context"`
	RiskEvaluation  RiskEvaluation      `json:"risk_evaluation"`
	PolicyDecision  PolicyDecision      `json:"policy_decision"`
	ReleaseEvidence ReleaseEvidenceLite `json:"release_evidence"`
}
