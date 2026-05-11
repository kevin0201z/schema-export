# AI 项目开发总控指南

本文档是 AI 参与项目开发时的总控入口。

它只保留最高层规则、任务路由和阻断分级；具体流程拆分到 `ai-guide/` 子文档中，模板和完整示例按需读取，避免每次处理任务都加载过长上下文。

---

## 1. 使用目标

本指南用于确保 AI 开发过程：

1. 需求先确认，避免一开始就跑偏。
2. 项目类型先判断，避免所有项目套同一种文档结构。
3. 文档按需生成，避免无意义堆文档。
4. 前端、后端、数据库通过契约保持一致。
5. 每个功能开发前有分析，开发后有验证和沉淀。
6. 后续 AI 可以根据已有文档继续开发。

---

## 2. 核心规则

AI 必须遵守以下核心规则：

| 编号 | 名称 | 含义 |
|---|---|---|
| R1 | 当前任务边界 | 只做当前任务，不扩展未确认功能或范围。 |
| R2 | 文档先行 | 涉及需求、接口、数据、页面或业务规则变化时，先更新文档，再实现代码。 |
| R3 | 契约一致 | 前端、API、后端、数据库字段必须一致，或在映射文档中明确说明。 |
| R4 | 高风险确认 | 高风险、破坏性或安全敏感变更必须先说明风险并获得用户确认。 |
| R5 | 验证闭环 | 完成后必须说明修改内容、验证结果、未验证风险和下一步建议。 |

---

## 3. 文档结构

```text
AGENTS.md                       # AI 工具入口
README.md                       # 指南包说明
QUICKSTART.md                   # 人类快速使用入口
AI_PROJECT_GUIDE.md             # 总控入口
SECURITY.md                     # 安全敏感变更规则
ai-guide/
├── 01-project-init.md          # 项目初始化流程
├── 02-project-types.md         # 项目类型与文档结构判断
├── 03-doc-system.md            # 文档体系与生成规则
├── 04-feature-development.md   # 功能模块开发流程
├── 05-contract-sync.md         # 前端/后端/数据库契约同步
├── 06-coding-validation.md     # 编码约束、测试与验收
├── 07-prompts.md               # 常用 Prompt 模板
├── 08-checklists.md            # 检查清单
└── 09-existing-project-onboarding.md # 已有项目接入流程
templates/                      # 项目文档模板
examples/                       # 长示例和落地样例
```

---

## 4. 文档分工

`ai-guide/` 是方法论文档，用于约束 AI 的工作流程。

`project-kit/templates/` 是项目文档模板，用于真正生成 `docs/` 下的项目文档。

`examples/` 用于提供完整样例和落地参考，不作为每次任务的默认阅读内容。

`SECURITY.md` 是安全敏感变更的统一入口，用于约束认证、权限、密钥、用户数据、支付、数据删除和部署安全。

`docs/` 是具体项目文档，应随着需求、功能、接口、数据库、页面和测试持续更新。

---

## 5. 任务类型与最小阅读集

AI 应先判断任务类型，再读取最小必要文档。

| 任务类型 | 必读 | 条件读取 |
|---|---|---|
| 文档错字、链接或格式修复 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/03-doc-system.md` | `project-kit/ai-guide/08-checklists.md` |
| 新项目初始化 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/01-project-init.md`、`project-kit/ai-guide/02-project-types.md`、`project-kit/ai-guide/03-doc-system.md` | 相关 `project-kit/templates/` 模板 |
| 已有项目接入 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/09-existing-project-onboarding.md` | `project-kit/ai-guide/03-doc-system.md`、`project-kit/ai-guide/08-checklists.md` |
| 功能开发 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/04-feature-development.md` | `project-kit/ai-guide/05-contract-sync.md`、`project-kit/ai-guide/06-coding-validation.md`、`project-kit/ai-guide/08-checklists.md`、相关 `docs/` |
| API 或字段变更 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/05-contract-sync.md` | `project-kit/templates/api-spec.template.md`、`project-kit/templates/contract-map.template.md`、相关 `docs/` |
| 数据库变更 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/05-contract-sync.md` | `project-kit/templates/data-model.template.md`、相关 `docs/`、`project-kit/SECURITY.md` |
| 编码、测试、验收规则调整 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/06-coding-validation.md` | `project-kit/ai-guide/08-checklists.md` |
| 代码审查或验收 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/05-contract-sync.md`、`project-kit/ai-guide/06-coding-validation.md` | `project-kit/ai-guide/08-checklists.md`、相关 `docs/` |
| 安全敏感变更 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/SECURITY.md` | `project-kit/templates/security-review.template.md`、相关 `docs/` |

普通任务不默认读取 `project-kit/SECURITY.md`。只有涉及认证、权限、密钥、支付、用户数据、数据删除、部署安全等场景时，才进入安全文档。

---

## 6. AI 阅读规则

### 6.1 新项目初始化

AI 必须阅读：

1. `project-kit/AI_PROJECT_GUIDE.md`
2. `project-kit/ai-guide/01-project-init.md`
3. `project-kit/ai-guide/02-project-types.md`
4. `project-kit/ai-guide/03-doc-system.md`

然后先向用户确认需求，不直接写代码。

用户确认需求并判断项目类型后，再读取 `project-kit/templates/` 中与项目类型相关的模板。

### 6.2 开发具体功能

AI 必须先阅读：

1. `project-kit/AI_PROJECT_GUIDE.md`
2. `project-kit/ai-guide/04-feature-development.md`

如果当前功能涉及 API、数据库、页面、字段映射、测试或验收，再按需读取：

1. `project-kit/ai-guide/05-contract-sync.md`
2. `project-kit/ai-guide/06-coding-validation.md`
3. `project-kit/ai-guide/08-checklists.md`
4. 项目内 `docs/` 下与当前功能相关的文档

然后先做功能影响分析，再决定是否更新 API、数据库、UI、字段映射等文档。

### 6.3 已有项目接入

AI 必须阅读：

1. `project-kit/AI_PROJECT_GUIDE.md`
2. `project-kit/ai-guide/09-existing-project-onboarding.md`

如需补文档或做现状检查，再按需读取：

1. `project-kit/ai-guide/03-doc-system.md`
2. `project-kit/ai-guide/08-checklists.md`

然后先扫描项目现状，不直接修改代码。

### 6.4 代码审查或验收

AI 必须阅读：

1. `project-kit/AI_PROJECT_GUIDE.md`
2. `project-kit/ai-guide/05-contract-sync.md`
3. `project-kit/ai-guide/06-coding-validation.md`

如需逐项核对，再按需读取：

1. `project-kit/ai-guide/08-checklists.md`
2. 当前需求、功能、API、数据库、页面相关文档

重点检查代码是否和需求、接口、数据库、页面文档一致。

---

## 7. 文档优先级

当文档、代码或用户指令之间出现冲突时，AI 必须按以下优先级判断：

1. 用户当前明确指令
2. `project-kit/AI_PROJECT_GUIDE.md`
3. `project-kit/ai-guide/` 子文档
4. 项目内 `docs/requirements.md`
5. 当前功能文档
6. `api-spec.md`、`data-model.md`、`ui-spec.md`、`contract-map.md`
7. 代码现状

如果文档和代码不一致，AI 不能直接按其中一方继续开发，必须先指出冲突并建议修正文档或代码。

如果用户当前明确指令会影响 API、数据库、权限、支付、部署或数据删除等高风险内容，AI 仍必须按 `R4` 先说明风险并请求确认。

---

## 8. 三阶段工作流

### 阶段一：项目初始化

```text
用户提供项目想法
→ AI 提问确认需求
→ AI 输出需求确认摘要
→ 用户确认
→ AI 判断项目类型
→ AI 生成匹配的文档结构
→ AI 拆分总体任务
→ AI 初始化项目骨架
→ AI 验证项目可运行
```

详见：`project-kit/ai-guide/01-project-init.md`

### 阶段二：已有项目接入

```text
用户提供已有项目
→ AI 扫描项目结构
→ AI 识别技术栈和项目类型
→ AI 查找已有文档和运行方式
→ AI 输出现状分析
→ AI 输出缺失文档和接入建议
→ 用户确认
→ AI 补齐必要项目文档
```

详见：`project-kit/ai-guide/09-existing-project-onboarding.md`

### 阶段三：功能模块开发

```text
用户指定功能
→ AI 判断任务类型并读取最小文档集
→ AI 做功能影响分析
→ AI 判断需更新哪些文档
→ AI 先更新文档
→ AI 按文档开发代码
→ AI 添加或更新测试
→ AI 运行验证
→ AI 更新任务状态和变更摘要
```

详见：`project-kit/ai-guide/04-feature-development.md`

---

## 9. 阻断分级

AI 遇到风险或不一致时，不再统一使用“必须拦截”这一种表达，而是按以下等级处理：

| 等级 | 处理方式 | 常见场景 |
|---|---|---|
| Hard Stop | 停止执行，先输出风险与待确认项，等待用户确认。 | 认证、权限、支付、数据删除、真实密钥、生产部署、破坏性删除、明确契约冲突 |
| Ask First | 先给出推荐方案和替代方案，编码前确认。 | 新增核心依赖、数据库结构修改、API 响应结构修改、大范围重构 |
| Continue With Note | 可以继续，但必须在变更摘要中说明原因、影响和剩余风险。 | 纯文案修正、注释修正、fixture/mock 小范围调整、无业务含义的格式改动 |

固定输出格式：

### Hard Stop

```md
该任务属于 Hard Stop：

- 变更目标：
- 风险点：
- 待确认项：
- 建议验证方式：

请确认后继续。
```

### Ask First

```md
该任务属于 Ask First：

- 变更内容：
- 影响范围：
- 推荐方案：
- 替代方案：

请确认采用哪种方案。
```

### Continue With Note

```md
本次变更属于 Continue With Note：

- 原因：
- 影响范围：
- 不需要额外确认的理由：
```

---

## 10. 一句话原则

先确认，再设计；先文档，再代码；先契约，再联调；按任务类型读取最小文档集；小步开发，持续沉淀。
