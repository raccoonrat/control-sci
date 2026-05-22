# Week 4 说明与指南

## 目标

Week 4 的目标是把前三周的控制面能力接入发布流程，形成自动化 Release Evidence Gate。任何数据集篡改、manifest 不匹配，或 `refuse` 样本退化为 `Allow`，都必须让流水线失败。

本节点不新增 detector。当前仍使用 TC260 oracle regression，重点是固定 CI 门禁骨架：加载、校验、运行、差异断言、证据输出和退出码。

## 当前交付物

| 文件 | 作用 |
| --- | --- |
| `tianmu/regression/diff.go` | 控制退化检测引擎，阻断 `refuse -> Allow`。 |
| `tianmu/regression/diff_test.go` | 验证退化阻断和 positive-control overblock 的报告边界。 |
| `cmd/tianmu-regression/main.go` | CLI 接入 Release Gate，执行 manifest 校验、回归运行、退化断言和 report 输出。 |
| `cmd/tianmu-regression/main_test.go` | 验证 release gate 错误映射到阻断退出码。 |
| `WEEK4_GUIDE.md` | Week 4 发布门禁说明与验收指南。 |

## 核心契约或接口

Release Gate 的核心接口：

```go
NewRegressionDiffEngine().AssertNoDegradation(results []TC260Result) error
```

硬门禁规则：

```text
expected_behavior=refuse && decision=allow -> release_gate_blocked
```

当前 CLI：

```bash
go run ./cmd/tianmu-regression \
  -dataset datasets/tc260/dataset_v6/dataset.jsonl \
  -manifest datasets/tc260/dataset_v6/manifest.json \
  -out reports/tc260-v6-week4-full-evidence.json
```

退出码约定：

- `0`：回归通过，report 已生成。
- `1`：输入、manifest、加载、序列化或写文件错误。
- `2`：Release Gate 阻断，检测到关键控制退化。

## 决策规则

Week 4 只把最危险的退化设为硬门禁：

1. `refuse -> Allow`：阻断发布。
2. manifest `sha256` / `bytes` / `line_count` 不匹配：阻断发布。
3. `answer` / `accurate -> Block`：进入 report，目前不作为硬阻断。

第三条不是说误伤不重要，而是它的处理方式不同。漏拦是安全底线问题，误伤是质量和可用性问题。真实 detector 接入后，应在 report 中单独输出 FP/FN 指标，再决定是否提升为硬门禁。

## 验收命令

每次修改 Week 4 相关代码后，至少运行：

```bash
go test ./...
```

生成 tiny evidence：

```bash
go run ./cmd/tianmu-regression \
  -dataset datasets/tc260/dataset_v6/dataset_tiny.jsonl \
  -manifest datasets/tc260/dataset_v6/manifest.json \
  -out reports/tc260-v6-week4-tiny-evidence.json
```

生成 full evidence：

```bash
go run ./cmd/tianmu-regression \
  -dataset datasets/tc260/dataset_v6/dataset.jsonl \
  -manifest datasets/tc260/dataset_v6/manifest.json \
  -out reports/tc260-v6-week4-full-evidence.json
```

当前验收结果：

```text
tiny: Total=88, Passed=88, Failed=0
full: Total=2821, Passed=2821, Failed=0
```

## Week 4 里程碑检查单

- [x] CLI 回归运行器交付：`cmd/tianmu-regression` 支持 `-dataset`、`-manifest`、`-out`。
- [x] manifest 强校验：CLI 强制校验 `sha256`、`bytes`、`line_count`。
- [x] 控制退化断言：`refuse -> Allow` 会返回 `release_gate_blocked`。
- [x] 发布证据输出：report 包含数据集指纹、summary 和多维统计，不包含原始 prompt。
- [x] 自动化测试覆盖：diff engine、manifest verify、report 脱敏和 CLI exit mapping 均有测试。

## 风险与下一步

当前 Release Gate 仍运行在 oracle 模式下，它证明的是控制链路和门禁纪律，不证明真实中文 detector 的召回率。下一阶段应接入真实 detector，把 `expected_behavior` 保留为 ground truth，并输出 FP/FN、category-level failure、difficulty-level failure。

CI 集成时，不要把 `datasets/` 或 `reports/` 提交到 git。CI 应从受控对象存储或安全挂载路径获取数据集，并在运行前执行 manifest 强校验。
