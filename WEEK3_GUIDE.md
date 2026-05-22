# Week 3 说明与指南

## 目标

Week 3 的目标是建立 Tool I/O Boundary：任何 Agent 工具调用都必须先经过注册、参数 schema 校验和控制面决策，不能让模型直接调用外部系统。

本节点只解决工具调用前的最小闭环：工具是否存在、参数是否合规、动作是否有副作用、风险信号是否需要动态调解。真实工具执行 runtime 不在本周范围内。

## 当前交付物

| 文件 | 作用 |
| --- | --- |
| `tianmu/core/mediation.go` | action-aware 动态调解路由，支持副作用确认和中文 PII 降级。 |
| `tianmu/core/mediation_test.go` | 验证 PII 到 `RedactThenAllow`、`Block` 优先级高于确认。 |
| `tianmu/core/engine.go` | `Engine` 改为使用 `EvaluateAction` 做决策收敛。 |
| `tianmu/toolgate/interceptor.go` | 工具注册中心、工具调用拦截、JSON 参数 schema 校验。 |
| `tianmu/toolgate/interceptor_test.go` | 覆盖副作用闸门、幽灵工具阻断、schema 错误、PII 调解和 benchmark。 |

## 核心契约或接口

`core.Evaluator` 新增 action-aware 入口：

```go
EvaluateAction(risk RiskEvaluation, action ActionContext) PolicyDecision
```

决策顺序：

1. 先执行基础 `PolicyPack` 规则。
2. 如果已经得到 `Block` 等更高优先级决策，不降级。
3. 如果基础决策为 `Allow` 且 `ActionContext.SideEffect=true`，升级为 `AskConfirmation`。
4. 如果基础决策为 `Allow` 且触发中文 PII 高置信信号，降级为 `RedactThenAllow`。

`toolgate.ToolInterceptor` 提供工具边界入口：

```go
RegisterTool(tool ToolDefinition) error
InterceptCall(sessionID string, toolName string, rawParams string, signals []core.DetectorSignal) (*core.ControlDecisionObject, error)
```

工具定义：

```go
type ToolDefinition struct {
  Name          string
  Description   string
  HasSideEffect bool
  ParamSchema   map[string]string
}
```

当前支持的 schema 类型：

- `string`
- `float` / `number`
- `bool` / `boolean`
- `object`
- `array`

## 决策规则

工具调用必须满足：

1. 工具名已注册。
2. 参数是 JSON object。
3. 参数字段不能缺失。
4. 参数不能包含 schema 外字段。
5. 参数类型必须匹配 schema。
6. 工具副作用声明为 `HasSideEffect=true` 时，默认要求人工确认。

未注册工具不进入 `Engine`，直接返回 `unregistered_tool_execution_denied`。这是边界隔离，不是普通策略失败。

## 验收命令

每次修改 Week 3 相关代码后，至少运行：

```bash
go test ./...
go test -bench=BenchmarkToolInterceptorInterceptCall -benchmem ./tianmu/toolgate
```

当前基准结果：

```text
BenchmarkToolInterceptorInterceptCall-22    1654 ns/op    1241 B/op    21 allocs/op
```

验收线：工具上下文构建和动态路由必须低于 `1.5ms`。当前约 `0.001654ms`，通过。

## Week 3 里程碑检查单

- [x] 幽灵工具防御隔离：未注册工具默认阻断，不进入核心引擎。
- [x] 副作用物理闸门就位：`HasSideEffect=true` 工具自动路由到 `AskConfirmation`。
- [x] 多态调解路由对齐：中文 PII 高置信信号可路由到 `RedactThenAllow`。
- [x] 参数 schema 校验： malformed、missing、unknown、type mismatch 均有测试覆盖。
- [x] 性能验收完成：工具拦截 fast-path 低于 `1.5ms`。

## 风险与下一步

当前 schema 是轻量类型校验，不是完整 JSON Schema。它足够支撑 Week 3 的边界闭环，但还不能表达 enum、min/max、pattern、nested required 等复杂约束。

下一步进入 Week 4 时，应把工具拦截结果纳入 release evidence：记录工具名、策略版本、决策、reason code 和 trace id，但不要记录完整 raw params。对于真实工具执行，还需要补 `ToolOutput` 边界，防止工具输出把敏感内容重新注入模型上下文。
