# Phase 1 Foundation 收口说明与计划

## 阶段目标

Phase 1 的目标是建立 Tianmu Runtime Control Plane 的最小工业闭环：控制契约、中文控制基线、工具边界、数据集回归和发布门禁必须能在本地与 CI 中重复运行。

本阶段不证明真实 detector 的召回率。它证明的是控制面骨架、证据链路和发布纪律。

## 已完成工程

### Week 1：核心契约与决策引擎

- `ControlDecisionObject` 已冻结。
- `PolicyPack` 权重收敛已实现。
- `Engine.MediateInbound` 生成 trace 级 `ReleaseEvidenceLite`。
- fast-path benchmark 低于 `0.1ms`。

### Week 2：中文控制有效性基线

- `sanitize.Normalizer` 已实现并覆盖形态学混淆。
- `SessionTracker` 已实现，并已接入 `Engine` 主链路。
- 连续 3 轮临界风险会触发 `AskConfirmation`。
- TC260 oracle loader、runner、manifest 校验和 evidence report 已实现。

### Week 3：Tool I/O Boundary

- `ToolInterceptor` 已实现工具注册、schema 校验和未注册工具阻断。
- 工具参数中的中文字符串会经过 normalizer。
- `HasSideEffect=true` 工具会路由到 `AskConfirmation`。
- 中文 PII 高置信信号会路由到 `RedactThenAllow`。

### Week 4：Release Evidence Gate

- `cmd/tianmu-regression` 已支持 `-dataset`、`-manifest`、`-out`。
- manifest 强校验覆盖 `sha256`、`bytes`、`line_count`。
- `RegressionDiffEngine` 阻断 `refuse -> Allow`。
- CLI 端到端测试已覆盖 report 生成和 prompt 脱敏。
- GitHub Actions 已接入基础测试和 fast-path benchmark。

## 本次补齐项

本次收口补齐了 Phase 1 中最重要的接线缺口：

1. `SessionTracker` 不再只是独立测试组件，已经进入 `Engine` 决策路径。
2. `ToolInterceptor` 不再直接旁路中文清洗器，工具参数字符串会做归一化。
3. `cmd/tianmu-regression` 增加端到端测试，覆盖临时数据集、manifest 和 evidence report。
4. 新增 `Makefile`，统一本地验收命令。
5. 新增 `.github/workflows/ci.yml`，提供 CI 基础门禁。

## 当前必须遵守的不变量

- 数据集不进 git，`datasets/*` 保持忽略。
- 报告不进 git，`reports/*` 保持忽略。
- Evidence report 不允许包含原始 prompt。
- `refuse -> Allow` 是硬阻断。
- 未注册工具不得进入 `Engine`。
- 副作用工具默认必须确认。
- Normalizer 不允许访问网络、文件或模型。

## 验收命令

本地基础验收：

```bash
make test
```

本地完整 fast-path 验收：

```bash
make verify
```

TC260 tiny 回归：

```bash
make regression-tiny
```

TC260 full 回归：

```bash
make regression-full
```

无 `make` 环境时使用：

```bash
go test ./...
go test -bench=BenchmarkMediateInboundFastPath -benchmem ./tianmu/core
go test -bench=BenchmarkNormalizerFastPath -benchmem ./tianmu/sanitize
go test -bench=BenchmarkSessionTrackerRecordAndEvaluate -benchmem ./tianmu/core
go test -bench=BenchmarkToolInterceptorInterceptCall -benchmem ./tianmu/toolgate
```

## 仍未进入 Phase 1 的工作

以下工作不应再混入 Foundation 收口，应作为下一阶段展开：

- 真实中文 detector 接入。
- FP/FN、category、difficulty 维度的 detector 质量报告。
- 跨版本 baseline artifact diff。
- Tool output 边界与输出脱敏。
- 完整 JSON Schema 支持。
- 外部 PolicyPack 配置与版本注册表。

## 下一阶段计划

Phase 2 应按以下顺序推进：

1. 定义 detector 接口，替换 TC260 oracle signal。
2. 在 `TC260Report` 中增加 FP/FN 和混淆矩阵。
3. 增加 baseline report 输入，做跨版本 diff。
4. 增加 Tool output validation，防止工具输出重新污染模型上下文。
5. 将 CI 从基础测试升级为携带受控数据集的 release gate。
