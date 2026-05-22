# Week 5 说明与指南

## 目标

Week 5 的目标是把 Phase 1 的 oracle 模拟信号推进到真实 detector 驱动的同步快线。当前实现只接入纯内存、无外部 I/O 的轻量 detector，避免在接口尚未稳定时引入远程模型或不可控时延。

本节点的判断标准：归一化后的中文输入必须能被 detector 识别，并通过 `Engine.InspectAndMediate` 自动汇聚信号、进入策略矩阵、产出 `ControlDecisionObject`。

## 当前交付物

| 文件 | 作用 |
| --- | --- |
| `tianmu/detector/interface.go` | `LLMDetector` 标准接口。 |
| `tianmu/detector/runner.go` | 多 detector runner，统一执行 normalizer 并补齐 signal 元数据。 |
| `tianmu/detector/keyword_injection.go` | 中文隐藏指令关键词 detector。 |
| `tianmu/detector/regex_pii.go` | 中国大陆手机号/身份证 PII detector。 |
| `tianmu/core/inspect.go` | `Engine.InspectAndMediate` 解耦管道。 |
| `tianmu/detector/pipeline_test.go` | live detector 到控制面端到端测试与 benchmark。 |
| `Makefile` | `verify-phase2` 加入 Week 5 detector benchmark。 |

## 核心接口

Detector 必须实现：

```go
type LLMDetector interface {
  ID() string
  Category() string
  Version() string
  Detect(ctx context.Context, normalizedPrompt string) (core.DetectorSignal, error)
}
```

核心管道：

```go
InspectAndMediate(ctx, normalizer, detectors, req, identity, data, action, rawPrompt)
```

执行顺序：

1. `sanitize.Normalizer` 清洗输入。
2. 并行执行所有 detector，并按注册顺序稳定汇总结果。
3. 补齐 signal 的 detector id、category、version。
4. 调用 `MediateInbound` 复用 Phase 1 决策底座。

## 内置 Detectors

`KeywordInjectionDetector`：

- category: `prompt_injection`
- 识别 `忽略上述指令`、`系统提示词`、`管理员权限`、`systemprompt` 等归一化后的隐藏指令原语。
- 命中置信度 `1.0`。

`RegexPIIDetector`：

- category: `chinese_pii`
- 识别中国大陆手机号与 18 位身份证号。
- 命中置信度 `0.95`。

## 不变量

- detector 输入必须是 normalized prompt。
- detector 必须声明 ID/category/version。
- 多路 detector 必须支持并行执行，结果顺序必须稳定。
- detector 不允许在 fast-path 中做不可控外部 I/O。
- `core` 不导入 `detector` 实现包，避免网关内核与算法实现耦合。
- `MediateInbound` 旧签名保持兼容。

## 验收命令

Week 5 验收：

```bash
make verify-phase2
```

定向测试：

```bash
go test ./tianmu/detector
```

性能验收：

```bash
go test -bench=BenchmarkBuiltinDetectors -benchmem ./tianmu/detector
go test -bench=BenchmarkInspectAndMediateLiveDetectors -benchmem ./tianmu/detector
```

SLO：多路内置 detector 与控制面管道应低于 `1.0ms`。

## Week 5 里程碑检查单

- [x] 探测器解耦接口冻结：`LLMDetector` 已稳定。
- [x] 首批真实 detector：中文隐藏指令与中国区 PII detector 已实现。
- [x] 解耦管道兼容：`InspectAndMediate` 保留旧 `MediateInbound`，新增自动探测路径。
- [x] 端到端验证：形态学隐藏指令可触发 `Block`，手机号可触发 `RedactThenAllow`。
- [x] 性能入口：`verify-phase2` 已包含 detector benchmark。

## 风险与下一步

当前 detector 是同步快线规则，不是语义模型。它能验证接口、信号和控制链路，但不能覆盖复杂语义绕过。

Week 6 应升级 `regression` report，加入 confusion matrix、refusal recall、false refusal rate，并让 TC260 支持 detector mode 与 oracle baseline 对比。
