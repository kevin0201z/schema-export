# AGENTS.md

本项目使用 AI Project Guide 作为 AI 辅助开发流程规范。

AI 助手在处理本项目任务时，应先阅读 `project-kit/AI_PROJECT_GUIDE.md`，根据任务类型选择最小阅读集，再按需进入 `project-kit/ai-guide/`、`project-kit/templates/`、`docs/` 和 `project-kit/SECURITY.md`。

## 总原则

1. 遵守 `project-kit/AI_PROJECT_GUIDE.md` 中的 `R1-R5` 核心规则。
2. 先判断任务类型，再读取最小必要文档。
3. 涉及规范、模板或示例调整时，同步检查相关入口文档。
4. 涉及安全敏感变更时，按条件读取 `project-kit/SECURITY.md`，并遵守 `R4`。
5. 完成后按 `R5` 说明修改内容、文档同步情况、验证结果和剩余风险。

## 任务入口

常见任务优先按以下路由读取：

| 任务类型 | 必读 | 条件读取 |
|---|---|---|
| 新增或调整工作流程 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/03-doc-system.md`、`project-kit/ai-guide/04-feature-development.md` | `project-kit/ai-guide/08-checklists.md` |
| 调整契约、API、数据库、字段映射规则 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/05-contract-sync.md` | `project-kit/templates/api-spec.template.md`、`project-kit/templates/data-model.template.md`、`project-kit/templates/contract-map.template.md` |
| 调整编码、测试、验收规则 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/ai-guide/06-coding-validation.md` | `project-kit/ai-guide/08-checklists.md` |
| 调整快速使用说明 | `project-kit/AI_PROJECT_GUIDE.md`、`README.md`、`project-kit/QUICKSTART.md` | 相关 `project-kit/ai-guide/` 子文档 |
| 安全敏感变更 | `project-kit/AI_PROJECT_GUIDE.md`、`project-kit/SECURITY.md` | `project-kit/templates/security-review.template.md` |

更完整的任务路由和阻断分级以 `project-kit/AI_PROJECT_GUIDE.md` 为准。

## 质量门禁

提交前运行：

```sh
npm run docs:check
```

如果检查失败，先修正文档结构、链接或模板问题，再继续交付。

## 禁止事项

- 不要把 `draft/` 作为当前规范依据。
- 不要在多个文档中复制大段规则而不说明主来源。
- 不要新增模板字段后遗漏对应指南或示例。
- 不要弱化 `R4` 的安全确认要求。
- 不要替用户选择许可证、发布渠道或商业授权条款；这些需要项目维护者确认。
