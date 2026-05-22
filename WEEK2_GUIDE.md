# Week 2 说明与指南

## 目标

Week 2 的目标是建立 Chinese Control Validity 的最小可回归基线。这里不追求训练一个“大而全”的中文 detector，而是先解决两个实际问题：中文形态学混淆要能在同步快线中归一化，多轮渐进式诱导要能被会话状态捕捉。

本节点的判断标准：中文攻击样本必须能进入固定测试集，控制逻辑必须能被 benchmark 验证，后续 detector 只能接在这个基线上，不能替代它。

## 当前交付物

| 文件 | 作用 |
| --- | --- |
| `tianmu/sanitize/normalizer.go` | 中文 fast-path 字符归一化，处理标点拆分、空白、全角字符、大小写和少量繁简变体。 |
| `tianmu/sanitize/normalizer_test.go` | 中文形态学攻击种子测试集与 normalizer benchmark。 |
| `tianmu/core/session_tracker.go` | 基于滑动窗口的会话风险追踪器，用于识别连续低到中风险的渐进式诱导。 |
| `tianmu/core/chinese_validity_test.go` | 多轮风险级联测试、窗口边界测试和 session tracker benchmark。 |
| `tianmu/regression/tc260.go` | TC260 JSONL 数据集加载器和 oracle regression runner。 |
| `tianmu/regression/tc260_test.go` | 基于 `datasets/tc260/dataset_v6/dataset_tiny.jsonl` 的本地集成测试。 |

## 核心契约或接口

`sanitize.Normalizer` 提供一个稳定入口：

```go
Normalize(input string) string
```

它只做纯内存、确定性的字符归一化，不做网络 I/O，不调用模型，不读写外部数据。这个边界必须保持，否则 fast-path 会变成性能和可用性的风险点。

`core.SessionTracker` 提供三个核心入口：

```go
RecordTurn(sessionID string, risk RiskEvaluation)
EvaluateCumulativeRisk(sessionID string) bool
Snapshot(sessionID string) (SessionHistory, bool)
```

`RecordTurn` 只记录风险指纹，不保存原始用户文本，避免 evidence 或 session state 变成新的敏感数据泄露面。

`regression.LoadTC260JSONL` 和 `regression.RunTC260Cases` 提供 TC260 数据集接入边界：

```go
LoadTC260JSONL(path string, limit int) ([]TC260Case, error)
RunTC260Cases(engine *core.Engine, normalizer *sanitize.Normalizer, cases []TC260Case) ([]TC260Result, TC260Summary, error)
```

TC260 在当前阶段只作为 oracle 回归集使用：`expected_behavior=refuse` 会转成 `tc260_refusal` 风险信号，再由 `PolicyPack` 收敛为 `Block`；`answer` 和 `accurate` 这类 positive-control 样本应收敛为 `Allow`。这验证的是控制面链路，不等同于已经有真实 detector。

## 决策规则

中文归一化规则当前覆盖：

1. 删除常见混淆标点和空白。
2. 全角 ASCII 转半角。
3. 英文字母统一小写。
4. 少量高频繁简或字形变体归并。

多轮风险规则当前覆盖：

1. 默认窗口大小为 `4` 轮。
2. 单轮风险分在 `[0.50, 0.80)` 之间视为临界风险。
3. 同一会话窗口内累计 `3` 轮临界风险，触发 `EvaluateCumulativeRisk`。

这些参数目前是代码级默认值，后续 Week 3/Week 4 可以提升为 policy 配置，但在 Week 2 不引入配置系统。

TC260 数据集使用规则：

1. 数据集放在 `datasets/tc260`，不提交到 git；`.gitignore` 已忽略 `datasets/*`。
2. 默认本地验收使用最新版本的 `dataset_v6/dataset_tiny.jsonl`。
3. `dataset_tiny.jsonl` 用于快速回归，`dataset.jsonl` 用于发布前全量回归。
4. `manifest.json` 里的 `sha256` 和 `line_count` 是数据版本边界，发布证据应记录它们。
5. 不要把 TC260 prompt 硬编码进普通单元测试，避免测试代码污染和敏感样本扩散。
6. `tc260_category` 可能为空，尤其是 positive-control 样本；加载器不能因此中断整批回归。

## 验收命令

每次修改 Week 2 相关代码后，至少运行：

```bash
go test ./...
go test -bench=. -benchmem ./tianmu/sanitize ./tianmu/core
```

如果本地存在 `datasets/tc260/dataset_v6/dataset_tiny.jsonl`，`go test ./...` 会自动执行 TC260 tiny 集成测试；如果数据集不存在，该测试会跳过。

手动只跑 TC260 集成：

```bash
go test ./tianmu/regression -run TestRunTC260TinyDataset
```

当前基准结果：

```text
BenchmarkNormalizerFastPath-22                 935.6 ns/op    96 B/op     1 allocs/op
BenchmarkSessionTrackerRecordAndEvaluate-22    110.4 ns/op    128 B/op    1 allocs/op
BenchmarkMediateInboundFastPath-22             668.2 ns/op    600 B/op    8 allocs/op
```

验收线：中文 fast-path 归一化必须低于 `0.5ms`。当前约 `0.000936ms`，通过。

## Week 2 里程碑检查单

- [x] 快线清洗模块交付：`sanitize.Normalizer` 支持标点混淆、全半角逃逸、空白混杂和少量繁简变体归一化。
- [x] 中文回归数据集种子版确立：固定形态学变体攻击样本，并纳入自动化测试。
- [x] 状态追踪器就位：`SessionTracker` 通过滑动窗口识别多轮渐进式风险累积。
- [x] 性能验收完成：normalizer 和 session tracker 均有 benchmark，fast-path 低于 `0.5ms`。
- [x] TC260 数据集接入：支持本地 JSONL 加载、tiny 回归和 oracle 决策链路验证。

## 风险与下一步

当前 Week 2 只建立了种子级中文控制基线，还没有覆盖同音字、拆字组合、拼音混写、Base64/Rot13 载荷和真实多轮上下文语义。下一步不要急着接大模型检测器；应该先把样本集从 4 类扩到可版本化 fixture，并把触发结果接入 `ControlDecisionObject` 的 evidence diff。

Week 3 可以进入 Tool I/O Boundary，但不能丢掉 Week 2 的基线：任何工具调用前的中文参数都应先经过 fast-path normalization，再进入策略和工具边界判断。
