# 快速开始

本文档用于快速使用这套 AI 项目开发指南。

如果你只想马上开始一个项目，按下面步骤走即可。默认原则是：先让 AI 判断任务类型，再读取最小必要文档。

---

## 1. 通过 Git 地址接入到开发项目

如果目标开发项目还没有 `project-kit/`，可以直接把本项目 Git 地址告诉 AI：

```text
https://github.com/kevin0201z/AI_Project_Guide.git
```

对 AI 说：

```md
请接入项目 https://github.com/kevin0201z/AI_Project_Guide.git，阅读 README.md 进行安装。
```

AI 应根据仓库 README 和 `AI_INSTALL_GUIDE.md` 完成接入。AI 识别入口必须落在目标项目根目录的 `AGENTS.md`，该文件应指向 `project-kit/AI_PROJECT_GUIDE.md`，并说明 AI 需要先判断任务类型，再读取最小必要文档。

---

## 2. 新项目怎么用

把 `project-kit/` 整个复制到新项目根目录：

```text
your-project/
└── project-kit/
    ├── AI_PROJECT_GUIDE.md
    ├── SECURITY.md
    ├── ai-guide/
    └── templates/
```

如果 AI 工具会自动读取根目录 `AGENTS.md`，再把 `project-kit/AGENTS.md` 复制一份到新项目根目录。

然后对 AI 说：

```md
你是本项目的 AI 开发助手。

请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：新项目初始化

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/01-project-init.md
- project-kit/ai-guide/02-project-types.md
- project-kit/ai-guide/03-doc-system.md

我的项目想法是：
【填写项目描述】

请先不要写代码。
请先向我确认需求，并在需求明确后输出需求确认摘要。
```

---

## 3. 已有项目怎么接入

对 AI 说：

```md
你是本项目的 AI 开发助手。

请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：已有项目接入

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/09-existing-project-onboarding.md

如有需要，再读取：

- project-kit/ai-guide/03-doc-system.md
- project-kit/ai-guide/08-checklists.md

这是一个已有项目。
请先扫描项目结构，识别技术栈、前端、后端、数据库和已有文档。

请不要直接修改代码。
请先输出项目现状分析、缺失文档列表和接入建议。
```

---

## 4. 开发具体功能怎么用

对 AI 说：

```md
你是本项目的 AI 开发助手。

请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：功能开发

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/04-feature-development.md

如果当前功能涉及 API、数据库、页面、字段映射、测试或验收，再按需读取：

- project-kit/ai-guide/05-contract-sync.md
- project-kit/ai-guide/06-coding-validation.md
- project-kit/ai-guide/08-checklists.md
- docs/ 下与当前任务相关的文档

当前要开发的功能是：
【填写功能名称】

请先不要直接编码。
请先做功能影响分析，判断是否需要更新 API、数据库、UI、字段映射和测试文档。
```

---

## 5. API 或数据库变更怎么用

对 AI 说：

```md
你是本项目的 AI 开发助手。

请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：API 或字段变更

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/05-contract-sync.md

如需更新文档模板，再读取：

- project-kit/templates/api-spec.template.md
- project-kit/templates/contract-map.template.md
- project-kit/templates/data-model.template.md

当前变更是：
【填写接口、字段或数据库变更】

请先判断这次修改属于 Hard Stop、Ask First 还是 Continue With Note，再给出文档和实现影响分析。
```

---

## 6. 安全敏感变更怎么用

对 AI 说：

```md
你是本项目的 AI 开发助手。

请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：安全敏感变更

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/SECURITY.md

如需安全审查文档，再读取：

- project-kit/templates/security-review.template.md

当前变更是：
【填写认证、权限、密钥、支付、数据删除或部署相关内容】

请先不要直接修改。
请先输出风险点、待确认项和建议验证方式。
```

---

## 7. 如何在 AGENTS.md 中使用本项目

很多 AI 开发工具会优先读取项目根目录下的 `AGENTS.md`。推荐把 `AGENTS.md` 作为本项目指南的入口文件，而不是把所有规则重复复制进去。

在目标项目根目录中放入：

```text
AGENTS.md
project-kit/
```

推荐直接把 `project-kit/AGENTS.md` 复制为目标项目根目录的 `AGENTS.md`。它会先把 AI 路由到 `project-kit/AI_PROJECT_GUIDE.md`，再根据任务类型进入对应子文档和模板。

如果你想手写一个更短的根目录 `AGENTS.md`，可以这样写：

```md
# AGENTS.md

本项目使用 AI Project Guide 作为 AI 辅助开发流程规范。

AI 助手在处理本项目任务时，应先阅读 `project-kit/AI_PROJECT_GUIDE.md`，根据任务类型选择最小阅读集，再按需进入 `project-kit/ai-guide/`、`project-kit/templates/`、`docs/` 和 `project-kit/SECURITY.md`。

## 总原则

1. 遵守 `project-kit/AI_PROJECT_GUIDE.md` 中的 `R1-R5` 核心规则。
2. 先判断任务类型，再读取最小必要文档。
3. 涉及安全敏感变更时，按条件读取 `project-kit/SECURITY.md`。
4. 完成后说明修改内容、验证结果和剩余风险。

## 任务入口

常见任务优先按以下路由读取：

| 任务类型 | 必读 | 条件读取 |
|---|---|---|
| 新项目初始化 | project-kit/AI_PROJECT_GUIDE.md, project-kit/ai-guide/01-project-init.md, project-kit/ai-guide/02-project-types.md, project-kit/ai-guide/03-doc-system.md | 相关 project-kit/templates |
| 功能开发 | project-kit/AI_PROJECT_GUIDE.md, project-kit/ai-guide/04-feature-development.md | project-kit/ai-guide/05-contract-sync.md, project-kit/ai-guide/06-coding-validation.md, project-kit/ai-guide/08-checklists.md, 相关 docs |
| 安全敏感变更 | project-kit/AI_PROJECT_GUIDE.md, project-kit/SECURITY.md | project-kit/templates/security-review.template.md |
```

使用时只需要让 AI 工具进入项目根目录。支持 `AGENTS.md` 的工具会先读取它，再根据里面的入口规则继续读取 `project-kit/AI_PROJECT_GUIDE.md` 和 `project-kit/ai-guide/`。

`AGENTS.md` 只做入口和路由；完整规则仍维护在 `project-kit/AI_PROJECT_GUIDE.md` 和 `project-kit/ai-guide/` 中，避免多处重复导致规则不一致。

---

## 8. 提交前怎么检查

对 AI 说：

```md
请先判断任务类型，并按最小阅读集做提交前检查。

当前任务类型：代码审查或验收

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/05-contract-sync.md
- project-kit/ai-guide/06-coding-validation.md

如需逐项核对，再读取：

- project-kit/ai-guide/08-checklists.md
- 当前需求、功能、API、数据库、页面相关文档

重点检查：

1. 是否有未同步的文档
2. API、前端、数据库字段是否一致
3. 是否修改了无关文件
4. 是否缺少测试或验证说明
5. 是否有高风险变更没有确认
```

---

## 9. 常见误区

1. 不要让 AI 一上来直接写完整项目。
2. 不要让前端先凭空写接口字段。
3. 不要让后端先凭空写数据库字段。
4. 不要每个项目都生成同一套文档。
5. 不要跳过功能影响分析。
6. 不要在文档和代码不一致时继续开发。
7. 不要把安全文档变成所有任务的默认阅读项。

---

## 10. 推荐阅读顺序

新项目：

```text
project-kit/AI_PROJECT_GUIDE.md
→ project-kit/ai-guide/01-project-init.md
→ project-kit/ai-guide/02-project-types.md
→ project-kit/ai-guide/03-doc-system.md
```

开发功能：

```text
project-kit/AI_PROJECT_GUIDE.md
→ project-kit/ai-guide/04-feature-development.md
→ 根据影响面按需进入 05、06、08 和相关 docs
```

已有项目接入：

```text
project-kit/AI_PROJECT_GUIDE.md
→ project-kit/ai-guide/09-existing-project-onboarding.md
→ 需要补文档或做核对时再进入 03、08
```

安全敏感变更：

```text
project-kit/AI_PROJECT_GUIDE.md
→ project-kit/SECURITY.md
→ 需要审查记录时再进入 security-review.template.md
```
