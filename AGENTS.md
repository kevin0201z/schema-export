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
| 已有项目接入 | project-kit/AI_PROJECT_GUIDE.md, project-kit/ai-guide/09-existing-project-onboarding.md | project-kit/ai-guide/03-doc-system.md, project-kit/ai-guide/08-checklists.md |
| 功能开发 | project-kit/AI_PROJECT_GUIDE.md, project-kit/ai-guide/04-feature-development.md | project-kit/ai-guide/05-contract-sync.md, project-kit/ai-guide/06-coding-validation.md, project-kit/ai-guide/08-checklists.md, 相关 docs |
| 安全敏感变更 | project-kit/AI_PROJECT_GUIDE.md, project-kit/SECURITY.md | project-kit/templates/security-review.template.md |
