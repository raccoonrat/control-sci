# Week 6 说明与指南

## 目标

Week 6 的目标是把 detector 质量从“是否通过”升级为可审计的四象限混淆矩阵。系统必须同时量化漏拦和误伤：`refuse -> Allow` 是安全逃逸，`answer/accurate -> Block` 是可用性误伤。

本节点继续遵守 evidence 脱敏红线：报告和 profile 不包含原始 prompt。

## 当前交付物

| 文件 | 作用 |
| --- | --- |
| `tianmu/regression/matrix.go` | 定义 `ConfusionMatrix` 和派生指标。 |
| `tianmu/regression/profiler.go` | 按 category 和 difficulty 聚合失败归因。 |
| `tianmu/regression/live.go` | 使用 live detector 跑 TC260 case 并输出矩阵。 |
| `tianmu/regression/report.go` | `TC260Report` 支持 matrix、metrics、profiler。 |
| `tianmu/regression/*_test.go` | 覆盖数学公式、维度归因、report 脱敏。 |
| `Makefile` | `verify-phase2` 增加 Week 6 度量测试。 |

## 核心接口

混淆矩阵：

```go
type ConfusionMatrix struct {
  TruePositive  int
  FalsePositive int
  TrueNegative  int
  FalseNegative int
}
```

派生指标：

```go
type DerivedMetrics struct {
  TotalCases       int
  Accuracy         float64
  Recall           float64
  FalseRefusalRate float64
}
```

live detector 回归：

```go
RunLiveDetectorsRegression(ctx, engine, normalizer, detectors, cases)
```

quality report：

```go
BuildTC260QualityReport(datasetPath, manifest, results, summary, matrix, profiler)
```

## 四象限定义

- `TP`：`refuse -> Block`
- `FN`：`refuse -> Allow`
- `TN`：`answer/accurate -> Allow`
- `FP`：`answer/accurate -> Block`

当前调解决策如 `RedactThenAllow`、`AskConfirmation` 不计作 `Allow` 逃逸。后续如果产品需要更细粒度质量度量，可以扩展为多类混淆矩阵。

## 指标定义

```text
Accuracy = (TP + TN) / Total
Recall = TP / (TP + FN)
FalseRefusalRate = FP / (FP + TN)
```

当 `FP + TN == 0`，说明测试集没有 positive-control 样本，`FalseRefusalRate` 必须返回 `0`，禁止产生 `NaN` 或破坏 JSON report。

质量门禁：

- FN / `refuse -> Allow`：硬风险，应进入 release gate。
- FP / `answer/accurate -> Block`：可用性风险，当前为软预警。

## 验收命令

Week 6 验收：

```bash
make verify-phase2
```

定向测试：

```bash
go test -run 'TestConfusionMatrix|TestProfiler|TestRunLiveDetectorsRegression' ./tianmu/regression
```

完整测试：

```bash
go test ./...
```

## Week 6 里程碑检查单

- [x] 四象限混淆矩阵就位：`TP`、`FP`、`TN`、`FN` 可计算。
- [x] 派生指标就位：`Accuracy`、`Recall`、`FalseRefusalRate` 可计算。
- [x] 多维归因就位：按 `tc260_category` 和 `difficulty` 聚合。
- [x] live detector 回归入口就位：可用真实 detector 跑 case。
- [x] Evidence 脱敏保持：quality report 不包含原始 prompt。

## 风险与下一步

当前 live detector 仍是轻量同步规则，TC260 全量真实效果不会等同于最终生产能力。Week 7 应实现跨版本 baseline report diff，把当前矩阵与历史稳定版对比，阻止 detector 更新导致召回率退化。
