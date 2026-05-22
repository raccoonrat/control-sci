# Week 7 说明与指南

## 目标

Week 7 的目标是把单次 quality report 升级为跨版本 Artifact Diff。发布门禁不能只看当前版本是否通过，还必须比较历史稳定版和当前构建，阻止已知用例发生防御漂移。

核心红线：历史版本 `Block` 的用例，当前版本不能变成 `Allow`。

## 当前交付物

| 文件 | 作用 |
| --- | --- |
| `tianmu/regression/artifact_diff.go` | 跨版本 evidence report 解析与 diff 引擎。 |
| `tianmu/regression/artifact_diff_test.go` | 验证 critical slip、usability regression 和新增用例行为。 |
| `tianmu/regression/report.go` | Evidence report 增加脱敏逐用例 `results`。 |
| `cmd/tianmu-regression/main.go` | CLI 新增 `-baseline`，执行跨版本差分门禁。 |
| `cmd/tianmu-regression/main_test.go` | 覆盖 baseline diff 触发 release gate。 |
| `Makefile` | `verify-phase2` 增加 artifact diff 测试。 |

## Diff 类型

`CriticalSlip`：

```text
baseline: Block
current:  Allow
```

这是硬阻断，CLI 返回 release gate 错误。

`UsabilityRegression`：

```text
baseline: Allow
current:  Block
```

这是可用性劣化，进入 diff report 和 warning，不作为当前硬阻断。

## Evidence Results

Week 7 起，`TC260Report` 包含逐用例脱敏指纹：

```go
type ReportCaseResult struct {
  ID               string
  TC260Category    string
  ExpectedBehavior string
  Decision         core.Decision
  ReasonCode       string
  Passed           bool
  Source           string
  Difficulty       string
}
```

禁止包含：

- `prompt`
- normalized prompt
- raw params
- tool output 原文

## CLI 用法

生成当前 report：

```bash
go run ./cmd/tianmu-regression \
  -dataset datasets/tc260/dataset_v6/dataset.jsonl \
  -manifest datasets/tc260/dataset_v6/manifest.json \
  -out reports/current.json
```

带 baseline 执行差分门禁：

```bash
go run ./cmd/tianmu-regression \
  -dataset datasets/tc260/dataset_v6/dataset.jsonl \
  -manifest datasets/tc260/dataset_v6/manifest.json \
  -baseline reports/baseline.json \
  -out reports/current.json
```

退出码：

- `0`：无 critical slip。
- `1`：输入、解析、manifest 或写文件错误。
- `2`：检测到 critical slip。

## 验收命令

Week 7 验收：

```bash
make verify-phase2
```

定向测试：

```bash
go test -run 'TestArtifactDiff|TestExecuteArtifactDiff' ./tianmu/regression ./cmd/tianmu-regression
```

完整测试：

```bash
go test ./...
```

## Week 7 里程碑检查单

- [x] 跨版本差分算子完成：支持 baseline/current report 比对。
- [x] 用例级脱敏结果就位：report 可用于 case id 级 diff。
- [x] Critical slip 硬门禁完成：`Block -> Allow` 触发 release gate。
- [x] Usability regression 软统计完成：`Allow -> Block` 进入 diff report。
- [x] CLI `-baseline` 完成：可在 CI 中接入历史 artifact。

## 风险与下一步

当前 artifact diff 依赖 report 中的 `results` 字段。旧 report 如果没有逐用例结果，只能做宏观指标比较，不能执行 case-level 防退化。进入 Week 8 前，应确保 CI 保存的 baseline report 使用 Week 7 之后的格式。

下一步 Week 8 应继续推进 Tool Outbound Validation，把工具输出边界的结果也纳入 evidence 和 artifact diff。
