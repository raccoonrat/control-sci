# Week 1 说明与指南

## 目标

Week 1 的目标是冻结 Tianmu Runtime Control Plane 的最小可信底座：先把控制契约、策略决策、运行时拦截和发布证据固定下来。不要急着堆 detector、RAG、Tool 网关或 CI 平台；没有稳定契约，后面的功能都会退化成散点补丁。

本周只接受一个判断标准：每一次控制决策都必须能被统一对象表达、被策略包解释、被测试回归、被证据追踪。

## 当前交付物

| 文件 | 作用 |
| --- | --- |
| `go.mod` | 定义 Go 模块边界。 |
| `tianmu/core/contract.go` | 定义 `ControlDecisionObject`、上下文、风险评估、策略决策和 `ReleaseEvidenceLite`。 |
| `tianmu/core/evaluator.go` | 基于 `PolicyPack` 将风险信号收敛为确定性决策。 |
| `tianmu/core/engine.go` | 提供 `MediateInbound` 拦截入口，生成完整控制决策对象。 |
| `tianmu/core/engine_test.go` | 覆盖决策、证据、性能和风险摘要的 Week 1 验收测试。 |

## 核心契约

`ControlDecisionObject` 是 Week 1 的冻结核心。任何后续功能都应向它输入或从它输出，不应绕过它直接做业务判断。

必须保持的字段组：

- `RequestContext`：产品、语言、交互类型。
- `IdentityContext`：Actor、Tenant、Role 和扩展属性。
- `DataContext`：数据分级、PII 标记、来源、目的地。
- `ActionContext`：动作类型、工具名、副作用声明。
- `RiskEvaluation`：风险类别、最大风险分、detector 版本、原始信号。
- `PolicyDecision`：最终决策、策略包版本、原因码。
- `ReleaseEvidenceLite`：证据等级、trace id、时间戳、回归标记。

## 决策规则

业务风险决策必须通过 `PolicyPack` 表达。不要在业务代码中继续增加针对某个 prompt、某个 detector 或某个场景的临时 `if-else`。

当前决策权重从高到低：

1. `Block`
2. `Escalate`
3. `AskConfirmation`
4. `Rewrite`
5. `RedactThenAllow`
6. `LogOnly`
7. `Allow`

副作用动作是唯一的 Week 1 引擎级覆写：如果 action 带 `SideEffect` 且策略仍为 `Allow`，则升级为 `AskConfirmation`。这是运行时控制不变量，不是业务特例。

## 验收命令

每次修改 Week 1 核心包后，至少运行：

```bash
go test ./...
go test -bench=BenchmarkMediateInboundFastPath -benchmem ./tianmu/core
```

当前基准结果：

```text
BenchmarkMediateInboundFastPath-22    1826409    655.2 ns/op    600 B/op    8 allocs/op
```

验收线：单次 fast-path 拦截必须低于 `0.1ms`。当前结果约 `0.000655ms`，通过。

## Week 1 里程碑检查单

- [x] 契约定义冻结：`ControlDecisionObject` 核心规范制定完成，兼容多租户与 Action 字段。
- [x] 代数决策矩阵：实现权重动态收敛，支持 `Block` / `RedactThenAllow` / `AskConfirmation` 等多态调解。
- [x] 高性能底座验证：单次拦截业务逻辑纯内存开销低于 `0.1ms`。
- [x] 技术发布证据：`ReleaseEvidenceLite` 可随控制决策对象 JSON 序列化，支持后续 CI 门禁。

## 后续节点输出要求

从 Week 2 起，每个节点完成后必须输出同样结构的 Markdown 文档。文档名称建议使用：

```text
WEEK<N>_GUIDE.md
```

每份节点文档必须包含以下部分：

1. `目标`：说明本节点解决的实际问题，避免空泛口号。
2. `当前交付物`：列出新增或修改的关键文件及其作用。
3. `核心契约或接口`：说明本节点冻结了哪些结构、接口、schema 或命令。
4. `决策规则`：说明哪些行为由配置、策略或规则驱动，哪些行为是引擎不变量。
5. `验收命令`：列出必须运行的测试、benchmark、lint 或生成命令。
6. `里程碑检查单`：用可勾选列表记录本节点是否完成。
7. `风险与下一步`：只写真实风险，不写假想威胁。

## 写作原则

- 写事实，不写宣传稿。
- 写可运行命令，不写“应该验证”。
- 写接口和不变量，不写临时实现细节。
- 写失败条件，不只写成功路径。
- 每个节点都必须能让后来的人用文档独立复现验收。

## 风险与下一步

当前 Week 1 底座能跑，但还只是核心控制面原型。下一步进入 Week 2 时，应该优先建设中文 fast-path：字符归一化、中文 PII 样例、prompt injection 样例和 regression fixture。不要先接入大模型 detector；先把可重复的中文控制基线做出来。
