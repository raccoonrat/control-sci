# Phase 2 核心接口与不变量设计规范

## 目标

Phase 2 的目标是从 Phase 1 的 oracle 控制链路，演进到真实 detector 与双向工具边界治理。当前节点先冻结核心接口和不变量，不直接承诺完整真实 detector 效果。

本次实施范围：

- 定义标准 detector 接口。
- 提供 detector runner，统一执行 normalizer。
- 升级 ToolInterceptor，支持工具输出边界 `InterceptOutput`。
- 固定 Phase 2 验收入口 `make verify-phase2`。

## 当前交付物

| 文件 | 作用 |
| --- | --- |
| `tianmu/detector/interface.go` | 定义 `LLMDetector` 标准接口。 |
| `tianmu/detector/runner.go` | 统一归一化 prompt，并执行多 detector。 |
| `tianmu/detector/runner_test.go` | 验证 detector metadata、归一化和错误传播。 |
| `tianmu/toolgate/interceptor.go` | 增加 `InterceptOutput` 双向工具边界。 |
| `tianmu/toolgate/interceptor_test.go` | 覆盖 hidden instruction、PII 和未注册工具输出阻断。 |
| `Makefile` | 新增 `verify-phase2` 验收入口。 |

## Detector 接口

所有真实 detector 必须实现：

```go
type LLMDetector interface {
  ID() string
  Category() string
  Version() string
  Detect(ctx context.Context, normalizedPrompt string) (core.DetectorSignal, error)
}
```

不变量：

- `ID()` 不能为空，且同一个 runner 内不能重复。
- `Category()` 必须映射到统一风险分类。
- `Version()` 必须进入 `DetectorSignal`，用于后续 evidence diff。
- `Detect` 输入必须是 normalized prompt。
- detector 不应在 fast-path 中做不可控外部 I/O。

## Detector Runner

Runner 入口：

```go
runner, err := detector.NewRunner(sanitize.NewNormalizer(), detectors...)
signals, err := runner.Detect(ctx, prompt)
```

Runner 负责：

- 统一执行 `sanitize.Normalizer`。
- 校验 detector 元数据。
- 补齐 `DetectorSignal.DetectorID`、`Category`、`Version`。
- 传播 detector 错误。
- 遵守 `context.Context` 取消语义。

Runner 不负责：

- 决策收敛。
- 策略路由。
- evidence report。

这些继续归属 `core.Engine` 和 `regression`。

## Tool Output Boundary

Phase 2 增加工具输出边界：

```go
InterceptOutput(sessionID string, toolName string, rawOutput string) (string, *core.ControlDecisionObject, error)
```

返回值：

- `string`：归一化后的工具输出。
- `ControlDecisionObject`：工具输出进入模型上下文前的控制决策。
- `error`：未注册工具或其他边界错误。

当前规则：

- 未注册工具输出默认拒绝。
- 工具输出会先归一化。
- 命中 hidden instruction 特征时生成 `prompt_injection` 信号，路由到 `Block`。
- 命中 PII 特征时生成 `chinese_pii` 信号，路由到 `RedactThenAllow`。

当前 output detector 是轻量边界规则，不代表 Phase 2 最终真实 detector。

## 硬不变量

- Evidence report 禁止包含原始 prompt。
- 未注册工具输入或输出都不得进入下游 runtime。
- `refuse -> Allow` 仍是 release gate 硬阻断。
- detector metadata 必须可追踪。
- fast-path 累计时延目标维持在 `<= 2ms`。

## 验收命令

基础验收：

```bash
make verify-phase2
```

完整 Foundation 验收：

```bash
make verify
```

TC260 oracle 回归仍可使用：

```bash
make regression-tiny
make regression-full
```

## 实施顺序

Week 5：

- 实现首批真实 detector。
- 将 TC260 runner 从 oracle 模式扩展为 detector 模式。

Week 6：

- 在 report 中加入 confusion matrix。
- 输出 FP/FN、category 和 difficulty 维度统计。

Week 7：

- 支持 baseline report 输入。
- 做跨版本 evidence diff。

Week 8：

- 扩展 Tool output detector。
- 增加工具输出脱敏与 report 统计。

## 当前边界

本节点冻结的是接口，不是最终检测能力。Phase 2 仍需要后续真实 detector、混淆矩阵、跨版本 diff 和更完整的工具输出治理。
