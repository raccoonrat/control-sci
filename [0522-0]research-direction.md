

下面我把 **“Tianmu on Control”** 作为一个真正的高价值研究课题来重构，而不是把它当作一个工程模块、Guardrails 子项目或 Personal AI release 支撑任务。我的判断是：

> **Tianmu on Control 的最高价值，不是做一组中文 guardrails，也不是做几个 detector，而是提出并实现一套面向 Personal AI → Enterprise AI 演进的“可证明、可审计、可迁移”的 AI Runtime Control Science。**

这件事如果做对，Tianmu 可以同时产出两类高价值成果：

1. **科研成果**：围绕 agentic AI control、action-level guarantee、evidence-producing guardrails、China-language validity、control abstraction 等形成论文、技术报告、benchmark、taxonomy、理论框架；
2. **工业工程成果**：形成可接入 Personal AI / Qira、后续可扩展到 Enterprise AI / xCloud / xIQ 的 Control Decision Contract、Guardrails Runtime、Release Evidence Pack、Regression Suite、Telemetry Schema。

以下是我作为“中国区顶级 AI 安全团队首席科学家”对该课题的深入判断。

***

# 1. 课题定位：Tianmu on Control 不是 Guardrails 项目，而是 Runtime Control Science

附件中的关键背景非常清楚：当前 release 顺序是 **Personal AI first → Enterprise AI next**，但这并不意味着 Tianmu 只做 Personal AI 的轻量 guardrail；正确策略是 “Personal AI 先落地、Enterprise AI 后扩展，但底层控制抽象必须一次设计正确”。附件明确提出 Personal AI 阶段应有最小可发布控制闭环，同时底层抽象必须能自然扩展到 Enterprise AI。

这意味着 Tianmu on Control 的课题不能被定义为：

```text
中文 guardrail
+ prompt injection detector
+ sensitive data detector
+ block/allow rules
```

这太低阶。

更高阶的定义应该是：

> **Tianmu on Control 是研究和构建一套面向 AI agent / LLM application 的运行时控制系统，用于在 policy、context、identity、data sensitivity、tool action、memory state 和 release requirement 共同约束下，对 AI 行为进行可解释、可审计、可回归、可迁移的决策与执行。**

也就是说，Control 的研究对象不是“模型输出是否安全”，而是：

```text
在当前上下文中，agent 是否可以执行这个 action？
是否可以接触这类 data？
是否可以调用这个 tool？
是否可以写入 memory？
是否必须 redact / block / degrade / ask confirmation / escalate？
这个决策能否被 evidence 支撑？
这个 evidence 能否进入 release gate？
```

这才是 Tianmu 在 AI Security 中真正可以做出科研与工业突破的方向。

***

# 2. 核心科学问题：从内容安全走向 Action-Level Runtime Control

当前很多 AI Safety / AI Security 工作仍然停留在：

```text
input filtering
output filtering
jailbreak detection
PII detection
policy classification
```

这些是必要的，但不是 Enterprise AI 的本质问题。附件已经指出 Personal AI 可以先做 Chinese User I/O guardrail、prompt injection / jailbreak benchmark、personal sensitive data detection、system prompt boundary hardening、basic preprocessing & sanitization、release evidence lite 和 regression suite；但同时必须保留面向 Enterprise AI 的 Control Decision Contract、Harness-Data Boundary Schema、detector/policy versioning、action risk taxonomy 和 evidence schema compatibility。

因此，Tianmu 应该把科学问题上升为：

> **如何从 input/output safety 过渡到 action-level security guarantee？**

这个问题非常关键，因为 Enterprise AI 的风险不是“说错话”这么简单，而是：

* agent 调用了错误工具；
* agent 在错误上下文中访问了敏感数据；
* agent 把内部信息发送到错误 destination；
* agent 将不可信 RAG 内容写入 memory；
* agent 在多步流程中发生 goal drift；
* agent 执行了不可逆业务动作；
* agent 留下的 telemetry / evidence 本身泄露敏感信息；
* release 时无法证明 control 有效。

这就是 Tianmu on Control 的第一个高价值研究命题：

## Research Thesis 1

### From Guardrails to Action-Level Runtime Control

**研究问题：**  
如何将传统 guardrails 从 prompt/response 层的分类器，提升为面向 agent action 的运行时控制系统？

**核心创新：**

```text
Content-level risk detection
→ Context-aware risk mediation
→ Action-level policy decision
→ Evidence-producing enforcement
→ Release-gate-compatible assurance
```

**科研价值：**

* 提出 Action-Level Control Taxonomy；
* 提出 Bounded Action Guarantee 模型；
* 定义 AI agent runtime 中 action、tool、data、memory、goal 的统一风险语义；
* 区分 content risk、data risk、tool risk、workflow risk、evidence risk。

**工业价值：**

* 形成 Tool I/O Validation Lite → Enterprise Tool Action Governance 的演进路径；
* 支撑 Personal AI 当前 release；
* 避免 Enterprise AI 阶段推倒重来；
* 支撑 xCloud/xIQ 后续集成。

***

# 3. Tianmu 的高价值研究方向一：Control Decision Contract

这是我认为 Tianmu on Control 最应该优先做的科研与工程核心。

附件已经明确提出：Personal AI 版本不能只做一次性实现，而要形成未来 Enterprise 可复用的 **control contract、risk taxonomy、detector interface、evidence schema 和 regression discipline**。

因此，第一个研究方向应是：

## Direction A：Unified Control Decision Contract for AI Runtime

### 3.1 为什么这是核心

没有 Control Decision Contract，所有 control 都会变成散点：

* 一个 detector 输出 risk score；
* 一个 guardrail 做 block；
* 一个 prompt template 做 hardening；
* 一个 scanner 输出 report；
* 一个 telemetry 系统记录日志。

这些东西单独都有价值，但无法组成 release-grade system。

Tianmu 应该提出一个统一抽象：

```text
Control Decision = f(
  request_context,
  user_or_agent_identity,
  data_context,
  action_context,
  policy_context,
  risk_evaluation,
  runtime_boundary,
  release_requirement
)
→ decision + mitigation + evidence
```

### 3.2 科研问题

* 如何形式化 AI runtime control decision？
* 哪些上下文是 Personal AI 阶段必须保留、Enterprise AI 阶段必须扩展的？
* 如何让 detector output、policy rule、tool action、data sensitivity、evidence schema 进入同一个 decision object？
* 如何设计既支持 local guardrails 又支持 enterprise audit 的 contract？

### 3.3 工业产物

```json
{
  "control_id": "cn-control-v1",
  "scenario_id": "personal-ai-user-io",
  "release_stage": "personal_ai",
  "request_context": {
    "product": "personal_ai",
    "language": "zh-CN",
    "interaction_type": "user_io"
  },
  "data_context": {
    "data_classification": "personal_sensitive",
    "contains_pii": true,
    "source": "user_input",
    "destination": "model_context"
  },
  "action_context": {
    "action_type": "generate_response",
    "tool_name": null,
    "side_effect": false
  },
  "risk_evaluation": {
    "risk_categories": ["privacy_leakage", "prompt_injection"],
    "risk_score": 0.83,
    "detector_versions": ["cn-pii-v1", "cn-injection-v1"]
  },
  "policy_decision": {
    "decision": "redact_then_allow",
    "policy_pack_version": "personal-ai-cn-policy-v1",
    "reason_code": "personal_sensitive_data_detected"
  },
  "evidence": {
    "evidence_level": "release_evidence_lite",
    "trace_pointer": "trace://...",
    "release_gate_impact": "requires_regression_pass"
  }
}
```

### 3.4 高价值成果形态

* 论文题目：**A Unified Decision Contract for Runtime Control of Agentic AI Systems**
* 工程资产：**Tianmu Control Decision Contract v1**
* Release 资产：**Personal AI Release Evidence Lite Schema**
* Enterprise 资产：**Enterprise Extension Fields for Role / Tenant / Destination / Tool Action**

***

# 4. 高价值研究方向二：Personal-to-Enterprise Control Invariants

附件中最强的一句话是：**Personal AI 阶段可以轻量实现，但不能轻量抽象。** 附件还明确区分了“可以轻量”的 enforcement scope、evidence depth、tool coverage、monitoring integration，与“不能轻量”的 control contract、evidence schema、action abstraction、data classification model、trace/evidence field design 和 release gate concept。

这可以形成一个非常有价值的研究方向：

## Direction B：Control Invariants Across Personal and Enterprise AI

### 4.1 核心问题

Personal AI 和 Enterprise AI 表面不同：

```text
Personal AI:
user safety, privacy, local guardrails, light evidence

Enterprise AI:
role-aware, tenant-aware, destination-aware, audit-heavy, governance-driven
```

但二者必须共享一些 **control invariants**：

* 每个 control decision 必须有 policy basis；
* 每个 detector 必须有 version；
* 每个 failure 必须能 regression；
* 每个 mitigation 必须有 evidence；
* 每个 action 必须可被抽象；
* 每个 data movement 必须可被分类；
* 每个 release 必须有 gate implication。

### 4.2 研究创新

提出：

> **Control Invariant Theory for AI Release Evolution**

即：不同 release stage 可以有不同 enforcement depth，但不能破坏底层 control invariant。

例如：

| 维度       | Personal AI         | Enterprise AI                            | Invariant                      |
| -------- | ------------------- | ---------------------------------------- | ------------------------------ |
| Identity | individual user     | role / tenant / department               | actor context must exist       |
| Data     | personal data       | enterprise confidential / regulated data | data classification must exist |
| Tool     | limited tool action | business-critical tool action            | action abstraction must exist  |
| Evidence | evidence lite       | structured audit evidence                | decision evidence must exist   |
| Release  | local gate          | governance/customer gate                 | release implication must exist |

### 4.3 工业价值

这个方向能解决管理层最关心的问题：

> Personal AI 先发，会不会导致 Enterprise AI 后面返工？

Tianmu 的回答应该是：

> 不会。因为我们不是构建 Personal-only guardrails，而是在 Personal AI 阶段建立 Enterprise-ready control invariants。

***

# 5. 高价值研究方向三：Chinese User I/O Control Validity

附件明确指出，Personal AI P0 包括 **Chinese User I/O Guardrail + Chinese Validity**，Tianmu 不应只做中文翻译，而要建立 Chinese User I/O Control Validity，交付中文 prompt injection / jailbreak test set、中文隐私泄露与个人敏感信息检测、中文 harmful / unsafe / manipulation 风险样例、input sanitization / output filtering 中文有效性报告、Local guardrail 中文 regression suite。

这可以形成 Tianmu 的第一个对外可展示科研 benchmark。

## Direction C：Chinese Control Validity Benchmark

### 5.1 核心问题

英文 guardrail 在中文场景下是否仍然有效？  
不能假设有效。必须测。

### 5.2 中文场景特有挑战

* 中文隐晦表达；
* 多轮渐进式诱导；
* 中英混合 prompt injection；
* 谐音、拆字、变体表达；
* 中文上下文省略；
* 企业内部代号；
* 本地合规语境；
* 中文 RAG 文档中的隐藏指令；
* 中文个人敏感信息格式；
* 中文企业敏感信息表达。

### 5.3 科研成果

* **Chinese Prompt Injection Benchmark**
* **Chinese Jailbreak Regression Suite**
* **Chinese Personal Data Leakage Benchmark**
* **Chinese Guardrail Degradation Report**
* **Cross-lingual Control Validity Study**

### 5.4 工业成果

* Personal AI 中文 release gate；
* Chinese user I/O guardrail baseline；
* Detector FP/FN report；
* Regression suite；
* Personal AI Release Evidence Lite。

***

# 6. 高价值研究方向四：Risk Mediation Beyond Allow/Block

很多 guardrails 的问题是只有：

```text
allow / block
```

但 Enterprise AI 需要的是风险调解：

```text
allow
block
redact
rewrite
degrade
ask confirmation
escalate
quarantine
log-only
rollback / compensate
```

## Direction D：Risk-Adaptive Mediation for AI Controls

### 6.1 研究问题

如何根据 risk type、confidence、user context、data sensitivity、action severity、destination、release stage 动态选择 mitigation？

例如：

* 个人 PII：redact\_then\_allow；
* system prompt extraction：block；
* 高风险 tool write：ask\_confirmation；
* 企业机密外发：escalate；
* RAG 文档含 hidden instruction：quarantine；
* 低置信检测：log\_only + regression candidate。

### 6.2 科研创新

提出 **Risk Mediation Policy Algebra**：

```text
Risk = category × severity × confidence × context × action_impact
Mitigation = policy_decision(Risk, ReleaseStage, ControlInvariant)
```

### 6.3 工业成果

* Decision Layer + Explainability；
* Reason Code Taxonomy；
* Release Gate Mapping；
* Guardrail mitigation policy；
* Evidence-producing decision log。

这会让 Tianmu 的 Control 从“拦截器”升级为“运行时风险决策系统”。

***

# 7. 高价值研究方向五：Evidence-Producing Guardrails

附件强调 Personal AI 阶段不需要完整 Enterprise audit，但必须交付 **Release Evidence Lite**，包括 evaluation result、detector version、failure examples、mitigation decision、regression status、release recommendation，并且该 schema 应兼容后续 Enterprise AI structured evidence。

这可以形成非常有差异化的研究方向：

## Direction E：Evidence-Producing Guardrails

### 7.1 核心判断

普通 guardrail 的输出是：

```text
blocked / allowed
```

高价值 guardrail 的输出是：

```text
decision + reason + policy + version + evidence + regression impact + release gate impact
```

### 7.2 科研问题

* 如何定义 guardrail evidence？
* 如何在不泄露敏感信息的前提下保留可审计证据？
* 如何让 evidence 同时服务 release、debug、audit、regression？
* 如何区分 Personal Evidence Lite 与 Enterprise Structured Evidence？

### 7.3 工业产物

```text
Release Evidence Lite v1
Control Regression Report v1
Decision Explanation Template v1
Evidence Schema Compatibility Layer
Enterprise Evidence Extension v1
```

### 7.4 高价值论文方向

* **Evidence-Producing Guardrails for Release-Gated AI Systems**
* **From Safety Filters to Auditable Runtime Controls**
* **Regression-Aware Evidence Generation for AI Security Controls**

这件事非常关键，因为它把 Tianmu 的工作从“做安全功能”提升为“生成 release truth”。

***

# 8. 高价值研究方向六：Tool I/O Validation Lite → Tool Action Governance

附件中明确指出，Personal AI 阶段 Tianmu 不需要完整 enterprise tool governance，但必须定义 tool action 的最小边界，即 Tool I/O Validation Lite，包括 tool intent check、sensitive input check、unsafe output check、risky action confirmation、Tool I/O evidence-lite schema；并强调如果 Personal 阶段完全不做 tool boundary，后续 Enterprise AI 会出现架构断层。

## Direction F：Progressive Tool Action Control

### 8.1 科学问题

如何从 Personal AI 的轻量 tool validation，演进到 Enterprise AI 的 action-level governance？

### 8.2 关键抽象

```text
Tool Action = tool_name + action_type + input_data + output_data + side_effect + destination + approval_requirement
```

即使 Personal AI 阶段只有少量 tool，也必须按 action object 建模。

### 8.3 研究内容

* Tool intent alignment；
* Tool input sensitivity；
* Tool output leakage；
* Side-effect classification；
* Human approval threshold；
* Destination-aware control；
* Tool decision evidence。

### 8.4 工业演进

```text
Tool I/O Validation Lite
→ Tool Action Risk Taxonomy
→ Tool Boundary Contract
→ Role-aware Tool Governance
→ Enterprise Tool Action Release Gate
```

这是 Tianmu 从 Personal AI 进入 Enterprise AI 的最关键桥梁之一。

***

# 9. 高价值研究方向七：RAG / Data Boundary Control for Guardrails

附件指出 Enterprise AI 阶段必须新增 RAG / Knowledge Base Data Control，因为 enterprise knowledge base 可能包含客户、合同、代码、设计、项目和受监管信息。

这意味着 Tianmu on Control 必须将 Data Testing 与 Guardrails 连接起来。

## Direction G：RAG-Aware Runtime Control

### 9.1 核心问题

RAG 风险不只是“文档是否敏感”，而是：

```text
能不能 ingestion？
能不能 retrieval？
谁能 retrieval？
retrieval 后能不能进入 prompt？
能不能进入 tool input？
能不能出现在 response？
能不能写入 memory？
能不能进入 evidence log？
```

### 9.2 科研创新

提出：

> **RAG Boundary Control Model**

覆盖：

* ingestion gate；
* retrieval gate；
* prompt assembly gate；
* response leakage gate；
* tool-use gate；
* memory write gate；
* evidence redaction gate。

### 9.3 工业产物

* RAG ingestion scanner；
* retrieval-time policy check；
* confidential chunk classifier；
* hidden instruction detector；
* RAG evidence report；
* Enterprise RAG release gate。

***

# 10. 高价值研究方向八：Release-Gated Control Regression

附件提出 Personal AI 阶段必须有 regression suite baseline，V2 应进入 detector versioning、FP/FN reporting、regression suite、decision explainability、release recommendation；Enterprise 阶段进入 continuous regression。

这可以形成 Tianmu 的另一个高价值工程体系。

## Direction H：Release-Gated Regression for AI Controls

### 10.1 核心问题

AI control 最大风险是：

* 模型更新导致 guardrail 失效；
* detector 更新导致 FP/FN 漂移；
* policy 更新导致行为变化；
* prompt hardening 改动引入新 bypass；
* 中文场景 regression 不稳定；
* Personal AI 到 Enterprise AI 迁移时 control 语义不一致。

### 10.2 研究与工程目标

建立：

```text
Control Regression = benchmark + detector version + policy version + mitigation decision + evidence diff + release gate impact
```

### 10.3 工业产物

* Control Regression Report；
* Chinese Regression Suite；
* Failure Taxonomy；
* Detector Version Registry；
* Policy Pack Version Registry；
* Release Recommendation Engine。

***

# 11. Tianmu on Control 的系统架构建议

我建议 Tianmu 把 Control 系统定义为五层：

```text
Layer 1: Risk Signal Layer
- prompt injection
- jailbreak
- sensitive data
- tool misuse
- RAG contamination
- memory poisoning

Layer 2: Context Layer
- product
- release stage
- language
- user role
- data source
- destination
- tool/action type

Layer 3: Control Decision Layer
- decision contract
- risk mediation
- reason code
- policy version
- confidence threshold

Layer 4: Enforcement Layer
- allow
- block
- redact
- rewrite
- confirm
- escalate
- quarantine
- log-only

Layer 5: Evidence & Release Layer
- evidence lite
- structured evidence
- regression report
- release gate mapping
- telemetry integration
```

这套架构的科研价值在于：它把 guardrails 从“检测模型”变成“运行时控制系统”。

工业价值在于：它可以从 Personal AI 小闭环逐步扩展到 Enterprise AI 大闭环。

***

# 12. Tianmu on Control 的首批科研/工程课题包

我建议将 Tianmu on Control 拆成 6 个高价值课题包。

## Topic 1：Control Decision Contract

**科研产出：** runtime control formalization  
**工程产出：** Unified Control Decision Contract v1  
**Release 价值：** Personal/Enterprise 共用抽象

## Topic 2：Chinese Control Validity

**科研产出：** Chinese guardrail benchmark / degradation study  
**工程产出：** Chinese User I/O Regression Suite  
**Release 价值：** Personal AI 中国区 release evidence

## Topic 3：Evidence-Producing Guardrails

**科研产出：** auditable guardrails framework  
**工程产出：** Release Evidence Lite v1  
**Release 价值：** 从功能上线变成 evidence-backed release

## Topic 4：Progressive Tool Action Control

**科研产出：** tool action risk taxonomy  
**工程产出：** Tool I/O Validation Lite → Tool Governance  
**Release 价值：** 防止 Personal 到 Enterprise 的架构断层

## Topic 5：RAG/Data Boundary Control

**科研产出：** RAG boundary control model  
**工程产出：** RAG ingestion/retrieval gate  
**Release 价值：** Enterprise AI 必备数据边界

## Topic 6：Release-Gated Regression

**科研产出：** regression methodology for AI controls  
**工程产出：** Control Regression Report + Release Gate Pack  
**Release 价值：** 持续发布纪律

***

# 13. Tianmu 应避免的低价值路线

下面这些方向看似有用，但如果单独做，会把 Tianmu 降级。

## 13.1 只做中文 detector

低价值原因：  
detector 没有 control contract、evidence、regression 和 release gate，就只是组件。

## 13.2 只做 prompt hardening

低价值原因：  
prompt 只是 harness 的一个边界，无法覆盖 tool、memory、RAG、data flow、action。

## 13.3 只做 preprocessing / sanitization

低价值原因：  
只能处理输入清洗，不能处理 action-level decision。

## 13.4 只做 Personal AI 本地 guardrail

低价值原因：  
如果没有 Enterprise-ready abstraction，后续会返工。

## 13.5 只做 benchmark

低价值原因：  
benchmark 不映射到 control 和 release gate，就不能驱动工程。

***

# 14. 我建议 Tianmu 对外定义的课题名称

可以将该方向命名为：

## 中文名称

> **面向 Personal-to-Enterprise AI 的运行时控制与证据闭环系统**

或更有研究味道：

> **面向 Agentic AI 的可迁移运行时控制理论与工程系统**

或更符合 DTL 语境：

> **Tianmu Control Plane for China-Ready AI Security**

## 英文名称

> **Tianmu Runtime Control Plane for Personal-to-Enterprise AI Security**

副标题：

> **From Guardrails to Evidence-Producing Action-Level Controls**

***

# 15. 最终战略判断

我会把 Tianmu on Control 的核心判断压缩成三句话：

1. **Personal AI first 决定 Tianmu 的第一批交付范围，但不能决定 Tianmu 的架构上限。**  
   附件已经明确：当前正确策略是用 Personal AI 做最小可发布闭环，用 Enterprise AI 要求约束 Control / Harness / Data 的设计抽象。

2. **Control 的研究对象不是 guardrail 本身，而是 AI runtime 中 action、data、policy、context、evidence 的统一决策。**

3. **Tianmu 的高价值不在于做更多 detector，而在于把 detector、policy、tool action、data boundary、evidence 和 release gate 组织成可复用的 Control Decision System。**

最终的课题收口可以这样说：

> **Tianmu on Control 要解决的不是“如何挡住坏输入/坏输出”，而是“如何让 AI agent 在 Personal AI 到 Enterprise AI 的演进过程中，其每一次关键行为都能被 policy 约束、被 context 解释、被 evidence 证明、被 regression 保护、被 release gate 接受”。**

这就是一个既有科研深度、又有工业落地价值的 Tianmu 核心方向。
