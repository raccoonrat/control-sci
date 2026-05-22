# 中文数据集接入计划与指南

## 目标

中文数据集接入的目标不是把样本堆进测试代码里，而是建立一条可重复、可版本化、可审计的控制回归链路：

```text
dataset -> loader -> normalizer -> detector/oracle signal -> Engine -> ControlDecisionObject -> regression summary
```

当前阶段优先接入 TC260 数据集，用它验证中文控制面是否稳定。TC260 现在作为 oracle 回归集使用，不等同于真实 detector 已完成。

## 数据目录约定

数据集统一放在：

```text
datasets/<dataset_name>/<dataset_version>/
```

当前 TC260 实际路径：

```text
datasets/tc260/dataset_v6/
```

每个版本目录建议包含：

```text
manifest.json
dataset.jsonl
dataset_tiny.jsonl
dataset.json
```

`datasets/*` 已在 `.gitignore` 中忽略。数据集不提交到 git，只提交 loader、runner、测试和指南。这样可以避免仓库膨胀，也避免敏感 prompt 扩散。

## TC260 当前格式

JSONL 每行是一条样本，当前核心字段：

```json
{
  "id": "5",
  "prompt": "...",
  "tc260_category": "A.1.a",
  "expected_behavior": "refuse",
  "source": "mutation_llm",
  "lang": "zh",
  "attributes": {
    "case_type": "attack",
    "control_type": "attack",
    "difficulty": "medium"
  }
}
```

硬约束字段：

- `id`
- `prompt`
- `expected_behavior`
- `lang`

非硬约束字段：

- `tc260_category`：positive-control 样本可能为空。
- `attributes.difficulty`：缺失时使用默认置信度。

## 接入分层

### 1. Loader 层

当前实现：

```go
regression.LoadTC260JSONL(path string, limit int) ([]TC260Case, error)
```

职责：

- 逐行读取 JSONL。
- 校验最小必需字段。
- 支持 `limit`，方便快速抽样。
- 不做策略判断。
- 不调用模型。

### 2. Normalizer 层

当前实现：

```go
sanitize.NewNormalizer().Normalize(prompt)
```

职责：

- 删除中文混淆标点和空白。
- 全角转半角。
- 英文字母小写化。
- 少量繁简或字形归并。

Normalizer 是 fast-path，不允许引入外部 I/O。

### 3. Signal 层

当前 TC260 使用 oracle signal：

```text
expected_behavior=refuse -> tc260_refusal signal
expected_behavior=answer/accurate -> no risk signal
```

当前映射：

- `refuse` -> `Block`
- `answer` -> `Allow`
- `accurate` -> `Allow`

这个阶段验证的是控制面链路，不验证真实 detector 能力。

### 4. Engine 层

TC260 runner 通过 Week 1 的核心引擎生成：

```go
ControlDecisionObject
```

它必须保留：

- `PolicyDecision.Decision`
- `PolicyDecision.ReasonCode`
- `RiskEvaluation.DetectorVersions`
- `ReleaseEvidence.TraceID`

后续 release gate 应基于这些字段做 evidence diff。

### 5. Summary 层

当前实现：

```go
regression.RunTC260Cases(engine, normalizer, cases)
```

输出：

- `Total`
- `Passed`
- `Failed`
- 每条样本的 decision、reason code、normalized prompt 和 failure。

### 6. Evidence Report 层

当前实现：

```go
regression.BuildTC260Report(datasetPath, manifest, results, summary)
```

报告包含：

- 数据集名称、版本、文件名。
- manifest 中的 `sha256`、`line_count`、`created_at`。
- expected behavior 分布。
- category、difficulty、source 维度分布。
- decision 分布。
- failure examples。

报告不包含原始 prompt。

CLI 在传入 `-manifest` 时会强制校验目标数据文件的 `sha256`、`bytes` 和 `line_count`。校验失败时不生成 report。

## 使用方式

快速验证所有包：

```bash
go test ./...
```

只跑 TC260 tiny 集成：

```bash
go test ./tianmu/regression -run TestRunTC260TinyDataset
```

生成 tiny evidence report：

```bash
go run ./cmd/tianmu-regression \
  -dataset datasets/tc260/dataset_v6/dataset_tiny.jsonl \
  -manifest datasets/tc260/dataset_v6/manifest.json \
  -out reports/tc260-v6-tiny-evidence.json
```

生成 full evidence report：

```bash
go run ./cmd/tianmu-regression \
  -dataset datasets/tc260/dataset_v6/dataset.jsonl \
  -manifest datasets/tc260/dataset_v6/manifest.json \
  -out reports/tc260-v6-full-evidence.json
```

如果本地不存在：

```text
datasets/tc260/dataset_v6/dataset_tiny.jsonl
```

测试会自动跳过。

## 接入计划

### Phase 1：Oracle 回归集

状态：已完成。

目标：

- 使用 `dataset_tiny.jsonl` 做快速回归。
- 使用 `expected_behavior` 直接生成 oracle signal。
- 验证 `Engine`、`PolicyPack`、`ControlDecisionObject`、`ReleaseEvidenceLite` 链路。

验收标准：

```bash
go test ./tianmu/regression -run TestRunTC260TinyDataset
```

必须通过。

### Phase 2：全量本地回归

状态：已完成。

目标：

- 使用 `dataset.jsonl` 跑全量 TC260。
- 输出结构化 summary。
- 记录 `manifest.json` 中的 `sha256`、`line_count`、`created_at`。

已实现：

```text
tianmu/regression/manifest.go
tianmu/regression/report.go
cmd/tianmu-regression/main.go
```

当前验收结果：

```text
tiny: Total=88, Passed=88, Failed=0
full: Total=2821, Passed=2821, Failed=0
```

CLI 用法：

```bash
go run ./cmd/tianmu-regression \
  -dataset datasets/tc260/dataset_v6/dataset.jsonl \
  -manifest datasets/tc260/dataset_v6/manifest.json \
  -out reports/tc260-v6-evidence.json
```

### Phase 3：真实 Detector 替换 Oracle

目标：

- 保留 TC260 expected behavior 作为 ground truth。
- 用真实中文 detector 输出 `DetectorSignal`。
- 比较 detector decision 和 expected behavior。

关键原则：

- Oracle 不删除，只作为基线。
- Detector 失败不能修改数据集结果来“掩盖”。
- False Positive 和 False Negative 必须进入 report。

建议指标：

- refusal recall
- positive-control pass rate
- false refusal rate
- category-level failure count
- difficulty-level failure count

### Phase 4：Release Evidence Gate

目标：

- 每次策略、normalizer、detector 改动都跑 TC260。
- 把结果写入 `ReleaseEvidenceLite` 或独立 regression report。
- 阻止高风险退化。

硬门禁：

- 原本应 `Block` 的样本不能退化为 `Allow`。
- positive-control 样本不能大面积误伤为 `Block`。
- 数据集版本必须在 evidence 中可追踪。

## 数据集版本策略

默认开发使用最新 tiny：

```text
datasets/tc260/dataset_v6/dataset_tiny.jsonl
```

发布前使用最新 full：

```text
datasets/tc260/dataset_v6/dataset.jsonl
```

当新增 `dataset_v7` 时，不要直接覆盖旧版本。正确做法：

1. 保留 `dataset_v6`。
2. 新增 `dataset_v7`。
3. 先用 oracle runner 跑通 v7。
4. 对比 v6/v7 的 category、expected behavior 和 failure 分布。
5. 再把默认测试版本切到 v7。

## 失败处理

如果 loader 失败：

- 检查是否缺少 `id`、`prompt`、`expected_behavior`、`lang`。
- `tc260_category` 为空不应失败。
- JSONL 单行过长时，调大 scanner buffer。

如果 oracle 回归失败：

- 先检查 `expected_behavior` 是否出现新枚举。
- 再检查 policy 是否包含 `tc260_refusal`。
- 不要直接改测试期望来压掉失败。

如果真实 detector 回归失败：

- `refuse -> Allow` 是漏拦，必须优先处理。
- `answer/accurate -> Block` 是误伤，必须按 category 和 source 聚合分析。
- 正常样本误伤率不能靠白名单硬编码解决。

## 禁止事项

- 不要把完整数据集提交到 git。
- 不要把大量 prompt 复制进普通单元测试。
- 不要让 normalizer 读取文件、访问网络或调用模型。
- 不要把 `expected_behavior` 当作线上 detector 输出。
- 不要用硬编码 category 补丁掩盖 detector 失败。

## 当前状态

已完成：

- TC260 JSONL loader。
- TC260 oracle runner。
- `dataset_v6/dataset_tiny.jsonl` 本地集成测试。
- `answer` / `accurate` positive-control 映射。
- `tc260_category` 为空的兼容处理。
- manifest 加载。
- manifest 文件级强校验。
- JSON evidence report 生成。
- category/difficulty/source 维度统计。
- `cmd/tianmu-regression` CLI runner。
- `dataset_v6` tiny/full oracle 回归通过。

下一步：

- 接入真实中文 detector，与 oracle baseline 对比。
- 增加 detector 模式下的 FP/FN report。
- 将 release gate 接到 CI。
