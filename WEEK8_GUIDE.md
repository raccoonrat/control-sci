# Week 8 说明与指南

## 目标

Week 8 的目标是完成 Tool Outbound Validation，把工具返回给模型上下文的数据纳入不可旁路控制面。外部 RAG、网页抓取、数据库读取等工具输出都应被视为不可信输入。

核心红线：工具输出不能把 hidden instruction 或 PII 原文带回模型上下文。

## 当前交付物

| 文件 | 作用 |
| --- | --- |
| `tianmu/core/contract.go` | 新增 `ActionProcessOutput` action 常量。 |
| `tianmu/toolgate/interceptor.go` | `InterceptOutput` 支持输出侧归一化、外部 signals、block fallback、PII redaction。 |
| `tianmu/toolgate/interceptor_test.go` | 覆盖二阶注入阻断、PII 脱敏、外部 signal、未注册工具输出拒绝和 benchmark。 |
| `Makefile` | `verify-phase2` 增加 outbound benchmark。 |

## 核心接口

```go
InterceptOutput(sessionID string, toolName string, rawOutput string, externalSignals ...[]core.DetectorSignal) (string, *core.ControlDecisionObject, error)
```

返回值：

- 第一个返回值：可安全进入模型上下文的输出文本。
- 第二个返回值：输出侧控制决策对象。
- 第三个返回值：边界错误。

## 决策行为

`Block`：

- 返回固定安全 fallback：

```text
[Tianmu Outbound Block] 外部工具返回数据违反系统合规红线，已拦截。
```

`RedactThenAllow`：

- 对手机号做掩码，例如 `13812345678 -> 138****5678`。
- 对身份证号替换为 `[REDACTED_ID_CARD]`。
- 脱敏路径使用单次 `strings.Builder` 扫描，避免大 payload 上连续正则替换产生中间字符串拷贝。

`Allow`：

- 返回经过 `sanitize.Normalizer` 归一化后的文本。

## 输出侧信号

内置输出侧信号：

- `indirect_injection`：二阶注入、hidden instruction、系统提示词等。
- `chinese_pii`：手机号、身份证、银行卡等 PII 指纹。

调用方也可以传入外部 detector signals：

```go
interceptor.InterceptOutput(sessionID, toolName, rawOutput, externalSignals)
```

## 不变量

- 未注册工具输出拒绝处理。
- 输出侧 action 必须是 `process_tool_output`。
- Block 不返回原始工具输出。
- Redact 不返回明文 PII。
- Normalizer 不做外部 I/O。
- 大 payload 脱敏不能通过多轮全量字符串替换实现。

## 验收命令

Week 8 验收：

```bash
make verify-phase2
```

定向测试：

```bash
go test -run TestToolInterceptorInterceptOutput ./tianmu/toolgate
```

性能验收：

```bash
go test -bench=BenchmarkToolInterceptorInterceptOutput -benchmem ./tianmu/toolgate
```

SLO：工具输出清洗 + 阻断判定应低于 `0.8ms`。

## Week 8 里程碑检查单

- [x] 双向拦截闭环完成：工具输入和输出均进入控制面。
- [x] 二阶注入阻断完成：hidden instruction 输出返回安全 fallback。
- [x] 输出侧 PII 脱敏完成：手机号和身份证不会明文返回。
- [x] 外部 signals 接入完成：输出侧可接收 detector 信号。
- [x] 性能验收入口完成：`verify-phase2` 覆盖 outbound benchmark。

## 风险与下一步

当前 redaction 是规则级掩码，足够支撑同步快线，但还不是完整 DLP。下一阶段应引入更完整的结构化 redaction span，输出脱敏位置、类型和证据指纹，但仍不能把原始输出写入 report。
