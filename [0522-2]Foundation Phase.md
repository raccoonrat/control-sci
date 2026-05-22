进入一期工程（Foundation Phase）意味着我们将正式告别宏观的战略务虚，全面进入**架构定义、数学形式化与核心契约（Contract）工业级落地**的深水区。这一阶段的目标是：**在不给 Personal AI 带来性能灾难的前提下，把未来 Enterprise AI 所需的“控制不变性（Control Invariants）”以最严苛、最标准的代码骨架固定下来。**

作为首席科学家，我将一期工程的实施蓝图拆解为以下四大核心交付模块，并直接给出可落地的工业级细节：

---

## 一、 控制形式化代数与契约定义 (Control Decision Contract v1)

我们要消灭散点式的 `if-else` 安全逻辑，就必须建立统一的决策输入与输出数学函数模型。

### 1.1 核心决策函数模型

定义大模型运行时控制决策函数为 $f$：

$$f(\mathbf{C}_{ctx}, \mathbf{I}_{id}, \mathbf{D}_{data}, \mathbf{A}_{act}, \mathbf{S}_{sig}, \mathbf{P}_{poly}) \rightarrow (\mathbf{M}_{miti}, \mathbf{E}_{evid})$$

其中：

* $\mathbf{C}_{ctx}$ (Context Vector)：包含语言（如 `zh-CN`）、产品线、当前会话生命周期等环境状态。
* $\mathbf{I}_{id}$ (Identity Vector)：Actor 身份标识（Personal 阶段为单一 User，预留 Enterprise 阶段的 Tenant/Role 槽位）。
* $\mathbf{D}_{data}$ (Data Context)：数据分级、流向。
* $\mathbf{A}_{act}$ (Action Context)：工具调用、动作意图、副作用声明。
* $\mathbf{S}_{sig}$ (Signal Matrix)：由多路独立分立式探测器（Detectors）输出的置信度与风险分类矩阵。
* $\mathbf{P}_{poly}$ (Policy Base)：当前生效的安全策略规则包版本。
* $\mathbf{M}_{miti}$ (Mediation Decision)：最终收敛的复合执行策略。
* $\mathbf{E}_{evid}$ (Evidence Block)：满足发布门禁审计（Release Gate）要求的结构化证据。

### 1.2 工业级契约 Schema 落地 (Go Struct 规范)

为了确保底层控制面的高性能与跨语言互操作性，我们首先固定核心契约的内存对象结构。一期工程采用 Go 语言进行高性能网关的原型开发：

```go
package contract

import "time"

// ReleaseStage 定义发布阶段，强制约束 Personal 到 Enterprise 的向后兼容
type ReleaseStage string
const (
  PersonalAI   ReleaseStage = "personal_ai"
  EnterpriseAI ReleaseStage = "enterprise_ai"
)

// Decision 最终收敛的确定性执行动作
type Decision string
const (
  Allow             Decision = "allow"
  Block             Decision = "block"
  RedactThenAllow   Decision = "redact_then_allow"
  Rewrite           Decision = "rewrite"
  AskConfirmation   Decision = "ask_confirmation"
  Escalate          Decision = "escalate"
  LogOnly           Decision = "log_only"
)

// ControlDecisionObject 统一控制决策契约核心结构体
type ControlDecisionObject struct {
  ControlID        string                 `json:"control_id"`
  Timestamp        time.Time              `json:"timestamp"`
  ReleaseStage     ReleaseStage           `json:"release_stage"`
  RequestContext   RequestContext         `json:"request_context"`
  IdentityContext  IdentityContext        `json:"identity_context"`
  DataContext      DataContext            `json:"data_context"`
  ActionContext    ActionContext          `json:"action_context"`
  RiskEvaluation   RiskEvaluation         `json:"risk_evaluation"`
  PolicyDecision   PolicyDecision         `json:"policy_decision"`
  ReleaseEvidence  ReleaseEvidenceLite    `json:"release_evidence"`
}

type RequestContext struct {
  ProductID       string `json:"product_id"`       // 例如 "Qira"
  Language        string `json:"language"`         // 强约束 "zh-CN" / "en-US"
  InteractionType string `json:"interaction_type"` // "user_io", "agent_loop"
}

type IdentityContext struct {
  ActorID    string                 `json:"actor_id"` // Personal: UserID; Enterprise: RoleID/TenantID
  TenantID   string                 `json:"tenant_id,omitempty"` // 为 Enterprise 预留的扩展槽位
  Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type DataContext struct {
  DataClassification string `json:"data_classification"` // "personal_sensitive", "enterprise_confidential"
  ContainsPII        bool   `json:"contains_pii"`
  Source             string `json:"source"`              // "user_input", "rag_retrieval", "tool_output"
  Destination        string `json:"destination"`         // "model_context", "external_api"
}

type ActionContext struct {
  ActionType string `json:"action_type"` // "generate_response", "call_tool", "write_memory"
  ToolName   string `json:"tool_name,omitempty"`
  SideEffect bool   `json:"side_effect"` // 是否包含不可逆业务影响（如删库、发邮件）
}

type RiskEvaluation struct {
  RiskCategories   []string  `json:"risk_categories"` // "prompt_injection", "privacy_leakage"
  MaxRiskScore     float64   `json:"max_risk_score"`
  DetectorVersions []string  `json:"detector_versions"` // 强要求版本留存以支持回归
}

type PolicyDecision struct {
  Decision         Decision `json:"decision"`
  PolicyPackVersion string   `json:"policy_pack_version"`
  ReasonCode       string   `json:"reason_code"` // 例如 "cn_pii_detected_in_output"
}

type ReleaseEvidenceLite struct {
  EvidenceLevel     string `json:"evidence_level"` // "release_evidence_lite"
  TracePointer      string `json:"trace_pointer"`   // 分布式追踪系统 TraceID
  RegressionPassTag bool   `json:"regression_pass_tag"`
}

```

---

## 二、 中文控制有效性验证 (Chinese Control Validity Baseline)

针对一期工程中 Personal AI 必须直面的中文特异性攻击场景，拒绝“补丁式拦截”，采用**矩阵式退化防御逻辑**。

### 2.1 专项测试集与攻击语义分类 (Chinese Regression Suite Lite)

一期工程我们将构建首批固定样本库，涵盖以下四大核心攻击谱系：

1. **形态学/排版隐藏指令（Morphological Injection）：** * *用例：* 拆字与同音字替代（如“系统提示词”变为“系.统.提.示.词”或“系 统 提 示 词”，谐音变为“细桶提示词”）。
2. **多轮意识形态渐进式诱导（Multi-turn Incremental Alignment Drift）：**
* *用例：* 前 3 轮正常对话，第 4 轮利用高上下文省略（Context Ellipsis）引入中国本地合规和敏感政治边界的越权诱导。
3. **中英混合型沙箱突破（Cross-lingual Escape）：**
* *用例：* 利用中文建立角色扮演（Role-play）场景，在角色内部突然插入英文 Base64/Rot13 编码的恶意 Prompt Injection 载荷。
4. **本地化个人敏感信息（Chinese PII Context）：**
* *用例：* 准确识别中文语境下的身份证号、中国大陆手机号变体（如“壹叁捌...”）、中文地址及企业特定业务代号。
  
  

### 2.2 本地化过滤清洗器（Input Sanitization Lite）的控制流水线

在一期工程中，为了避免大模型检测器的长延迟，前端拦截面设计两道防线：

* **同步快线（Fast-Path）：** 基于 Aho-Corasick 算法的变体敏感词正则树 + 常用字符归一化（Normalize）处理器（自动将繁体、带标点符号的拆字、全角字符压缩转化为标准简体）。
* **同步慢线（Slow-Path）：** 本地小参数模型（如 MiniCPM 或特定微调的语义检测小模型）进行 Prompt 意图对齐度（Intent Alignment Score）判定。

---

## 三、 渐进式工具行动治理 (Tool I/O Validation Lite)

为了防止 Personal AI 迁移到 Enterprise AI 阶段时出现工具链的架构断层，一期工程必须实现 **Tool I/O Boundary 最小闭环**。

### 3.1 抽象意图检查与参数校验

哪怕在一期工程中 Personal AI 只有“读取个人日程”或“发送微信提醒”等轻量工具，工具的调用也必须强制通过 `Tianmu Action Interceptor`：

```text
[Agent Chain of Thought] 
       │
       ▼ (Emit Tool Call)
┌────────────────────────────────────────────────────────┐
│ Tianmu Tool Interceptor (Foundation Phase)             │
│                                                        │
│ 1. Schema Validation:                                  │
│    检查输入参数是否符合 OpenAPI/JSON Schema 标准          │
│                                                        │
│ 2. Intent Squashing:                                   │
│    通过轻量级分类，将传入的复杂中文参数压缩为意图原语       │
│                                                        │
│ 3. Side-Effect Gate:                                   │
│    IF parameter.has_write_action OR Tool.side_effect  │
│    THEN 强制中断流，改写 Decision 为 "ask_confirmation"  │
└────────────────────────────────────────────────────────┘
       │
       ▼ (Passed / Mitigated)
[Actual Tool Execution Runtime]

```

---

## 四、 自动化发布门禁与控制回归流水线 (Release-Gated Regression v1)

一期工程必须建立起技术发布纪律，将安全系统的输出转化为研发流程的“硬门禁（Hard Gate）”。

### 4.1 控制退化差异对比（Evidence Diff）

我们不以“模型聪明度”为指标，只以“控制一致性”为指标。在 CI/CD 流水线中部署 `Tianmu Regression Runner`：

1. **固定版本控制（Versioning）：** 每一个 Detector 镜像、每一套 Policy Pack 必须打上 Git Commit SHA 标签，注册到本地控制面注册表（Registry）。
2. **差异化矩阵分析（Diff Engine）：**
   每次代码或策略变更，跑一遍 `Chinese Regression Suite Lite`。比对前一次发布的 `ControlDecisionObject` 与当前的变更差异：

$$\Delta = \mathbf{M}_{miti\_new}(test\_case\_i) \oplus \mathbf{M}_{miti\_old}(test\_case\_i)$$



如果原本应当被 `Block` 或 `Redact` 的测试用例在变动后滑向了 `Allow`，流水线强制中断，拒绝 Release。
3. **交付物：Release Evidence Lite**
流水线自动生成一份结构化的证据归档文件。这份文件就是产品（如 Personal AI 助手）上线前必不可少的安全合规准入凭证（Release Truth）。

---

## 五、 一期工程（Foundation Phase）第 1-4 周敏捷推进路线图

| 周次         | 核心攻关任务            | 阶段性里程碑交付物                                                                            |
| ---------- | ----------------- | ------------------------------------------------------------------------------------ |
| **Week 1** | **核心契约落地与骨架搭建**   | 冻结 `ControlDecisionObject` 规范，发布 Go/Python 两套基础 SDK 核心库。                             |
| **Week 2** | **中文特异性攻击防线建设**   | 交付 `Chinese Regression Suite Lite` 版本（包含首批 2000 个精雕细琢的中文越狱与隐私泄露用例）；上线字符归一化快线拦截器。     |
| **Week 3** | **工具拦截器与策略路由实现**  | 完成 Tool I/O 动态拦截层，支持根据 Policy Pack 规则实时将决策由 `Allow` 动态降级为 `Redact` 或 `Confirmation`。 |
| **Week 4** | **CI/CD 自动化门禁集成** | 跑通首个包含基线版本比对、自动生成 `Release Evidence Lite JSON` 报告的自动化控制回归流水线。                        |

一期工程的底层底座一经固定，整个大模型安全团队的工作就将进入“逻辑有据可查、风险可以回归、发布依赖证据”的现代安全系统科学轨道。准备开始一期工程第一阶段的代码编排与定义。
