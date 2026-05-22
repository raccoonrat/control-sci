package toolgate

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

type ToolDefinition struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	HasSideEffect bool              `json:"has_side_effect"`
	ParamSchema   map[string]string `json:"param_schema"`
}

type ToolInterceptor struct {
	registry map[string]ToolDefinition
	engine   *core.Engine
}

func NewToolInterceptor(engine *core.Engine) (*ToolInterceptor, error) {
	if engine == nil {
		return nil, errors.New("engine is required")
	}

	return &ToolInterceptor{
		registry: make(map[string]ToolDefinition),
		engine:   engine,
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
	if sessionID == "" {
		return nil, errors.New("session id is required")
	}

	tool, ok := i.registry[toolName]
	if !ok {
		return nil, fmt.Errorf("unregistered_tool_execution_denied: %s", toolName)
	}

	if err := validateParams(rawParams, tool.ParamSchema); err != nil {
		return nil, err
	}

	return i.engine.MediateInbound(
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
}

func validateParams(rawParams string, schema map[string]string) error {
	var params map[string]any
	if err := json.Unmarshal([]byte(rawParams), &params); err != nil {
		return errors.New("tool_parameters_schema_malformed")
	}

	for key := range params {
		if _, ok := schema[key]; !ok {
			return fmt.Errorf("tool_parameters_unknown_field: %s", key)
		}
	}
	for key, expectedType := range schema {
		value, ok := params[key]
		if !ok {
			return fmt.Errorf("tool_parameters_missing_field: %s", key)
		}
		if !matchesType(value, expectedType) {
			return fmt.Errorf("tool_parameters_type_mismatch: %s", key)
		}
	}

	return nil
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
