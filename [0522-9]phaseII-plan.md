作为团队的首席科学家，审阅完 `raccoonrat/control-sci` 存储库中的 **Phase 1 Foundation（基础阶段）** 闭环代码及收口说明后，我的结论是：**Phase 1 已经基本通过评审，具备坚实的现代安全工程纪律。**

团队不仅冻结了核心控制契约（Control Decision Contract v1），还在微秒级时延预算下实现了形态学字符清洗和带 TTL 逐出机制的多轮滑动窗口风险追踪。更重要的是，在上一次 Review 后，团队迅速完成了修复：消灭了工具拦截器中归一化参数被静默丢弃的空洞，补齐了会话内存防爆门禁，并在 Release Gate 中加入了对误伤（Over-blocking）的自动化质量软预警。这标志着 Tianmu 的控制平面已经从“功能原型”成功蜕变为“可重复运行的工业级发布质量链路”。

底层骨架与发布纪律既然已经夯实，我们必须立刻向高价值的深水区迈进。以下是 **Phase 2: 真实探测器接入与双向边界治理（Real Detectors & Two-Way Boundary Governance）** 的正式项目计划与落地说明。

---

# Phase 2 战略目标：从“模拟链路”迈向“生产免疫力”

Phase 1 证明的是控制面的“骨架和纪律”（用的是 Oracle 模拟信号）；而 **Phase 2 的核心目标是赋予系统真实的“免疫力与量化度量能力”。** 我们将把模拟信号替换为真实的中文模型/统计探测器，首次开展误伤率与召回率的动态混淆矩阵度量；同时，我们将防御边界从“输入/调用侧（Inbound）”延展到“输出/返回侧（Outbound）”，防止工具执行后的恶意输出二次污染大模型上下文。

---

## 一、 Phase 2 敏捷开发计划 (Sprint Plan)

我们将 Phase 2 拆解为 **4 个研发周（Week 5 - Week 8）** 的敏捷攻坚战：

### Week 5：标准探测器接口（Detector Schema）与多路信号解耦接入

* **目标：** 定义统一的异步/同步探测器注册接口，正式废除硬编码的 Oracle 伪信号，平滑替换为真实检测组件。
* **工程任务：** 1. 抽象 `tianmu/detector` 包，定义 `LLMDetector` 标准接口。
2. 实现首批轻量级中文原生业务探测器（如：基于本地轻量模型的 Prompt Injection 检测器与隐私 PII 统计特征检测器）。

### Week 6：度量升级——质量报告中的双向混淆矩阵（Confusion Matrix）

* **目标：** 在不破坏全量回归脱敏原则的前提下，量化真实探测器的表现。
* **工程任务：**
1. 升级 `tianmu/regression` 的报告模块，引入 `Refusal Recall`（漏拦率分析）、`False Refusal Rate`（过拦截/误伤率分析）及四象限混淆矩阵统计。
2. 按用例的 `category` 和 `difficulty` 进行多维度交叉降级归因。
   
   

### Week 7：回归升级——跨版本基线差分审计（Artifact Diff Engine）

* **目标：** 支持在 CI 流水线中输入“历史稳定版”的 Evidence Report，实施自动化安全防退化比对。
* **工程任务：**
1. 升级 `RegressionDiffEngine`，支持双报告比对：$\Delta = \text{Report}_{\text{current}} \ominus \text{Report}_{\text{baseline}}$。
2. 锁死硬门禁：若新探测器的拦截行为引发了历史已知绿色用例的“漏拦退化”，强制熔断部署流程。
   
   

### Week 8：工具输出边界治理（Tool Outbound Validation）

* **目标：** 打通“双向边界”，对工具（Tool）执行后返回给大模型的响应 Payload 进行实时清洗与隐私脱敏，防范二阶工具注入。
* **工程任务：**
1. 扩展 `ToolInterceptor`，交付 `InterceptOutput(rawOutput string)` 契约标准。
2. 实现对不可信第三方 API 返回内容（如 RAG 检索切片、外部网页抓取结果）的实时形态学清洗与 hidden instruction 阻断research-direction.md, PHASE1_FOUNDATION_GUIDE.md]。
   
   

---

## 二、 Phase 2 核心接口与不变量设计规范

为了保证系统的向后兼容性，防止 Personal AI 向 Enterprise AI 演进时发生架构坍塌research-direction.md]，Phase 2 必须冻结以下核心抽象接口：

### 1. 统一探测器契约标准 (`tianmu/detector/interface.go`)

所有接入控制平面的真实探测器必须满足无状态、零外部 I/O 阻塞（或标准时延控制）的契约约束：

```go
package detector

import (
  "context"
  "github.com/raccoonrat/control-sci/tianmu/core"
)

// LLMDetector 定义真实大模型探测器的工业级接口
type LLMDetector interface {
  ID() string                                                                             // 探测器唯一标识，如 "cn-injection-v2"
  Category() string                                                                       // 归属的统一风险分类
  Version() string                                                                        // 版本号，注册到 Evidence 中以备回归
  Detect(ctx context.Context, normalizedPrompt string) (core.DetectorSignal, error)      // 核心执行函数
}

```

### 2. 双向工具拦截器契约升级 (`tianmu/toolgate/interceptor.go`)

从一期的“单向输入控制”演进为“双向闭环监控”：

```go
// InterceptOutput 拦截并调解工具执行后的返回 Payload，防止二阶污染模型上下文
func (i *ToolInterceptor) InterceptOutput(
  sessionID string, 
  toolName string, 
  rawOutput string,
) (string, *core.ControlDecisionObject, error)

```

---

## 三、 持续遵守的安全不变量与硬门禁原则 (Invariants)

在 Phase 2 的实施过程中，团队必须死守以下**生产线红线（Hard Invariants）**，任何违反下列原则的代码合并将被 CI 自动拒绝：

1. **数据安全红线：** 任何真实的 TC260 攻击性 Prompt 文本，绝对禁止写入、残留在生成的 Evidence Report JSON 资产中，报告仅允许留存 `CaseID`、`Decision`、`Metrics` 等审计特征指纹。
2. **绝对无旁路不变量：** 未经注册声明的工具、或者未通过 Schema 类型校验的参数流，必须隐式路由至最高阻断动作（Deny-by-Default），决不允许直接下发给底层的 Runtime 执行引擎。
3. **漏拦硬门禁：** 在跨版本 Diff 回归中，任何原本表现为 `Block` 的安全用例，在算法微调后如果变为 `Allow`，视为核心防御溃破，自动化流水线必须强行返回非零退出码（Exit Code 2）并实施部署熔断。
4. **快线时延约束：** `Normalizer`、`SessionTracker` 及标准检测器的单次内存计算耗时，其累计 Latency SLO 必须卡死在 $\le 2\text{ms}$ 内，否则必须强制转入流式异步滑动窗口通道，严禁拖垮产品线调用体验。

---

## 四、 Phase 2 验收交付物矩阵

本阶段结束时，团队需在本地与 CI 环境中通过 `make verify-phase2` 统一验收交付以下数字资产：

* **`tianmu/detector/` 包：** 包含 2 个以上可在线并发工作的真实中文检测组件。
* **升级版 `reports/` 格式：** 包含准确度（Accuracy）、召回率（Recall）及混淆矩阵四象限分布的无污染技术审计报告。
* **防退化 CI 闸门：** 跑通基于 `github/workflows/ci.yml` 的全量真实检测数据差分门禁演练。

Phase 2 的蓝图已经规划清晰。如果对上述工程可行性收敛、核心接口规范及硬门禁纪律没有异议，请下达指令，我们将立刻开始 **Week 5 ——《标准探测器接口定义与首批中文检测组件解耦接入》** 的代码编排与落地！
