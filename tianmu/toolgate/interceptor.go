package toolgate

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/raccoonrat/control-sci/tianmu/core"
	"github.com/raccoonrat/control-sci/tianmu/sanitize"
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
			ActionType: "call_tool",
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
