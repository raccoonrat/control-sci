package toolgate

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/raccoonrat/control-sci/tianmu/core"
	"github.com/raccoonrat/control-sci/tianmu/sanitize"
)

const outboundBlockFallback = "[Tianmu Outbound Block] 外部工具返回数据违反系统合规红线，已拦截。"

var (
	outboundPhoneRegex  = regexp.MustCompile(`1[3-9]\d{9}`)
	outboundIDCardRegex = regexp.MustCompile(`[1-6]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]`)
)

type ToolDefinition struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	HasSideEffect bool              `json:"has_side_effect"`
	ParamSchema   map[string]string `json:"param_schema"`
}

type ToolInterceptor struct {
	registry   map[string]ToolDefinition
	engine     *core.Engine
	normalizer *sanitize.Normalizer
}

type InterceptedCall struct {
	Decision        *core.ControlDecisionObject
	SanitizedParams map[string]any
}

func NewToolInterceptor(engine *core.Engine) (*ToolInterceptor, error) {
	return NewToolInterceptorWithNormalizer(engine, sanitize.NewNormalizer())
}

func NewToolInterceptorWithNormalizer(engine *core.Engine, normalizer *sanitize.Normalizer) (*ToolInterceptor, error) {
	if engine == nil {
		return nil, errors.New("engine is required")
	}
	if normalizer == nil {
		return nil, errors.New("normalizer is required")
	}

	return &ToolInterceptor{
		registry:   make(map[string]ToolDefinition),
		engine:     engine,
		normalizer: normalizer,
	}, nil
}

func (i *ToolInterceptor) RegisterTool(tool ToolDefinition) error {
	if tool.Name == "" {
		return errors.New("tool name is required")
	}
	if tool.ParamSchema == nil {
		tool.ParamSchema = map[string]string{}
	}

	i.registry[tool.Name] = tool
	return nil
}

func (i *ToolInterceptor) InterceptCall(sessionID string, toolName string, rawParams string, signals []core.DetectorSignal) (*core.ControlDecisionObject, error) {
	call, err := i.InterceptCallWithPayload(sessionID, toolName, rawParams, signals)
	if err != nil {
		return nil, err
	}

	return call.Decision, nil
}

func (i *ToolInterceptor) InterceptCallWithPayload(sessionID string, toolName string, rawParams string, signals []core.DetectorSignal) (*InterceptedCall, error) {
	if sessionID == "" {
		return nil, errors.New("session id is required")
	}

	tool, ok := i.registry[toolName]
	if !ok {
		return nil, fmt.Errorf("unregistered_tool_execution_denied: %s", toolName)
	}

	sanitizedParams, err := validateAndNormalizeParams(rawParams, tool.ParamSchema, i.normalizer)
	if err != nil {
		return nil, err
	}

	decision, err := i.engine.MediateInbound(
		core.RequestContext{
			ProductID:       "Qira",
			Language:        "zh-CN",
			InteractionType: "agent_loop",
		},
		core.IdentityContext{ActorID: sessionID},
		core.DataContext{
			DataClassification: "personal_sensitive",
			ContainsPII:        false,
			Source:             "model_context",
			Destination:        "external_api",
		},
		core.ActionContext{
			ActionType: core.ActionCallTool,
			ToolName:   toolName,
			SideEffect: tool.HasSideEffect,
		},
		signals,
	)
	if err != nil {
		return nil, err
	}

	return &InterceptedCall{
		Decision:        decision,
		SanitizedParams: sanitizedParams,
	}, nil
}

func (i *ToolInterceptor) InterceptOutput(sessionID string, toolName string, rawOutput string, externalSignals ...[]core.DetectorSignal) (string, *core.ControlDecisionObject, error) {
	if sessionID == "" {
		return "", nil, errors.New("session id is required")
	}
	tool, ok := i.registry[toolName]
	if !ok {
		return "", nil, fmt.Errorf("unregistered_tool_output_processing_denied: %s", toolName)
	}

	sanitizedOutput := i.normalizer.Normalize(rawOutput)
	signals := append(outputSignals(sanitizedOutput), flattenSignals(externalSignals)...)
	decision, err := i.engine.MediateInbound(
		core.RequestContext{
			ProductID:       "Qira",
			Language:        "zh-CN",
			InteractionType: "agent_loop",
		},
		core.IdentityContext{ActorID: sessionID},
		core.DataContext{
			DataClassification: "tool_output",
			ContainsPII:        hasOutputPII(signals),
			Source:             "external_api",
			Destination:        "model_context",
		},
		core.ActionContext{
			ActionType: core.ActionProcessOutput,
			ToolName:   tool.Name,
			SideEffect: false,
		},
		signals,
	)
	if err != nil {
		return "", nil, err
	}

	switch decision.PolicyDecision.Decision {
	case core.Block:
		return outboundBlockFallback, decision, nil
	case core.RedactThenAllow:
		return redactOutboundPII(sanitizedOutput), decision, nil
	default:
		return sanitizedOutput, decision, nil
	}
}

func validateAndNormalizeParams(rawParams string, schema map[string]string, normalizer *sanitize.Normalizer) (map[string]any, error) {
	var params map[string]any
	if err := json.Unmarshal([]byte(rawParams), &params); err != nil {
		return nil, errors.New("tool_parameters_schema_malformed")
	}

	for key := range params {
		if _, ok := schema[key]; !ok {
			return nil, fmt.Errorf("tool_parameters_unknown_field: %s", key)
		}
	}
	for key, expectedType := range schema {
		value, ok := params[key]
		if !ok {
			return nil, fmt.Errorf("tool_parameters_missing_field: %s", key)
		}
		if !matchesType(value, expectedType) {
			return nil, fmt.Errorf("tool_parameters_type_mismatch: %s", key)
		}
	}

	return normalizeParamValues(params, normalizer), nil
}

func matchesType(value any, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number", "float":
		_, ok := value.(float64)
		return ok
	case "bool", "boolean":
		_, ok := value.(bool)
		return ok
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	default:
		return false
	}
}

func normalizeParamValues(params map[string]any, normalizer *sanitize.Normalizer) map[string]any {
	if normalizer == nil {
		return params
	}

	normalized := make(map[string]any, len(params))
	for key, value := range params {
		normalized[key] = normalizeParamValue(value, normalizer)
	}

	return normalized
}

func normalizeParamValue(value any, normalizer *sanitize.Normalizer) any {
	switch typed := value.(type) {
	case string:
		return normalizer.Normalize(typed)
	case []any:
		values := make([]any, len(typed))
		for idx, item := range typed {
			values[idx] = normalizeParamValue(item, normalizer)
		}
		return values
	case map[string]any:
		return normalizeParamValues(typed, normalizer)
	default:
		return value
	}
}

func outputSignals(sanitizedOutput string) []core.DetectorSignal {
	signals := make([]core.DetectorSignal, 0, 2)
	if containsAny(sanitizedOutput, []string{"系统提示词", "忽略以上", "忽略之前", "忽略前文", "打开沙箱", "管理员权限"}) {
		signals = append(signals, core.DetectorSignal{
			DetectorID: "tool-output-hidden-instruction",
			Category:   "indirect_injection",
			Version:    "tool-output-boundary-v1",
			Confidence: 0.90,
			Triggered:  true,
		})
	}
	if containsAny(sanitizedOutput, []string{"身份证", "手机号", "银行卡"}) || outboundPhoneRegex.MatchString(sanitizedOutput) || outboundIDCardRegex.MatchString(sanitizedOutput) {
		signals = append(signals, core.DetectorSignal{
			DetectorID: "tool-output-pii",
			Category:   core.ChinesePIICategory,
			Version:    "tool-output-boundary-v1",
			Confidence: 0.90,
			Triggered:  true,
		})
	}

	return signals
}

func hasOutputPII(signals []core.DetectorSignal) bool {
	for _, signal := range signals {
		if signal.Triggered && signal.Category == core.ChinesePIICategory {
			return true
		}
	}

	return false
}

func flattenSignals(groups [][]core.DetectorSignal) []core.DetectorSignal {
	total := 0
	for _, group := range groups {
		total += len(group)
	}
	signals := make([]core.DetectorSignal, 0, total)
	for _, group := range groups {
		signals = append(signals, group...)
	}

	return signals
}

func redactOutboundPII(output string) string {
	var builder strings.Builder
	builder.Grow(len(output))

	for idx := 0; idx < len(output); {
		if isIDCardAt(output, idx) {
			builder.WriteString("[REDACTED_ID_CARD]")
			idx += 18
			continue
		}
		if isPhoneAt(output, idx) {
			builder.WriteString(output[idx : idx+3])
			builder.WriteString("****")
			builder.WriteString(output[idx+7 : idx+11])
			idx += 11
			continue
		}

		r, size := utf8.DecodeRuneInString(output[idx:])
		if r == utf8.RuneError && size == 0 {
			break
		}
		builder.WriteString(output[idx : idx+size])
		idx += size
	}

	return builder.String()
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}

	return false
}

func isPhoneAt(value string, idx int) bool {
	if idx+11 > len(value) || !isDigitBoundary(value, idx, 11) {
		return false
	}
	if value[idx] != '1' || value[idx+1] < '3' || value[idx+1] > '9' {
		return false
	}
	for offset := 0; offset < 11; offset++ {
		if !isDigit(value[idx+offset]) {
			return false
		}
	}

	return true
}

func isIDCardAt(value string, idx int) bool {
	if idx+18 > len(value) || !isDigitBoundary(value, idx, 18) {
		return false
	}
	if value[idx] < '1' || value[idx] > '6' {
		return false
	}
	for offset := 1; offset < 6; offset++ {
		if !isDigit(value[idx+offset]) {
			return false
		}
	}
	yearPrefix := value[idx+6 : idx+8]
	if yearPrefix != "18" && yearPrefix != "19" && yearPrefix != "20" {
		return false
	}
	for offset := 8; offset < 10; offset++ {
		if !isDigit(value[idx+offset]) {
			return false
		}
	}
	month := value[idx+10 : idx+12]
	if month < "01" || month > "12" {
		return false
	}
	day := value[idx+12 : idx+14]
	if day < "01" || day > "31" {
		return false
	}
	for offset := 14; offset < 17; offset++ {
		if !isDigit(value[idx+offset]) {
			return false
		}
	}
	last := value[idx+17]
	return isDigit(last) || last == 'X' || last == 'x'
}

func isDigitBoundary(value string, idx int, length int) bool {
	before := idx == 0 || !isDigit(value[idx-1])
	afterIdx := idx + length
	after := afterIdx >= len(value) || !isDigit(value[afterIdx])
	return before && after
}

func isDigit(value byte) bool {
	return value >= '0' && value <= '9'
}
