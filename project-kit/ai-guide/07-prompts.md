# 常用 Prompt 模板

> AI 快速摘要
>
> - 适用场景：为不同任务类型提供可直接使用的提示词模板。
> - 必须产出：明确任务类型、最小阅读集、当前任务目标和输出要求。
> - 硬性阻断：不要继续传播“所有功能开发都默认读 04/05/06/08 全套”的旧写法。
> - 相关模板：`project-kit/AI_PROJECT_GUIDE.md`
> - 可跳过场景：用户已经明确给出自己的高质量提示词。

本文档提供项目初始化、功能开发、继续开发、已有项目接入、代码审查和文档同步检查时的提示词。

基础版适合小任务或低风险任务。严格版适合新项目、复杂功能、多人协作、涉及 API 或数据库的任务。

---

## 1. 项目初始化 Prompt

### 基础版

```md
你是本项目的 AI 开发助手。

请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：新项目初始化

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/01-project-init.md

我的项目想法是：
【填写项目描述】

请先不要写代码。
请先向我确认关键需求，并输出需求确认摘要。
```

### 严格版

```md
你是本项目的 AI 开发助手。

请严格遵守：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/01-project-init.md
- project-kit/ai-guide/02-project-types.md
- project-kit/ai-guide/03-doc-system.md

当前任务类型：新项目初始化

我的项目想法是：
【填写项目描述】

请先不要写代码。
请先向我确认需求，并在需求明确后输出需求确认摘要。
确认后再判断项目类型、规划文档结构、拆分任务和初始化项目骨架。
```

---

## 2. 已有项目接入 Prompt

### 基础版

```md
你是本项目的 AI 开发助手。

当前任务类型：已有项目接入

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/09-existing-project-onboarding.md

这是一个已有项目。
请先扫描项目结构，输出项目现状分析和缺失文档建议。

请先不要修改代码。
```

### 严格版

```md
你是本项目的 AI 开发助手。

请严格遵守：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/09-existing-project-onboarding.md

如需补文档或做完整核对，再读取：

- project-kit/ai-guide/03-doc-system.md
- project-kit/ai-guide/08-checklists.md

当前任务类型：已有项目接入

请先不要修改代码。
请先完成：

1. 扫描项目结构
2. 识别技术栈
3. 判断项目类型
4. 查找已有文档
5. 识别前端、后端、数据库和测试情况
6. 输出项目现状分析
7. 输出缺失文档列表
8. 给出接入建议
```

---

## 3. 功能开发 Prompt

### 基础版

```md
你是本项目的 AI 开发助手。

当前要开发的功能是：
【填写功能名称或任务】

请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：功能开发

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/04-feature-development.md
- 相关 docs 文档

先做功能影响分析，判断是否需要更新文档，再实现代码并验证。
```

### 严格版

```md
你是本项目的 AI 开发助手。

请严格遵守：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/04-feature-development.md

如果当前功能涉及 API、数据库、页面、字段映射、测试或验收，再按需读取：

- project-kit/ai-guide/05-contract-sync.md
- project-kit/ai-guide/06-coding-validation.md
- project-kit/ai-guide/08-checklists.md
- docs/ 下已有项目文档

当前任务类型：功能开发

当前要开发的功能是：
【填写功能名称或任务】

请先不要直接编码。
请先完成：

1. 阅读相关文档
2. 功能影响分析
3. 判断需要新增或更新哪些文档
4. 判断属于 Hard Stop、Ask First 还是 Continue With Note
5. 给出开发步骤和验收标准

如果涉及 Hard Stop 或 Ask First，请先等待我确认。
```

---

## 4. 继续开发 Prompt

```md
你是本项目的 AI 开发助手。

请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：功能开发

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/04-feature-development.md
- docs/requirements.md
- docs/tasks.md
- 与当前任务相关的功能文档

然后继续开发当前任务：
【填写任务】

要求：

- 不修改无关文件
- 不跳过文档同步
- 不擅自变更接口或数据库
- 开发完成后更新任务状态和变更摘要
```

---

## 5. 代码审查 Prompt

```md
请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：代码审查或验收

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/05-contract-sync.md
- project-kit/ai-guide/06-coding-validation.md

如需逐项核对，再读取：

- project-kit/ai-guide/08-checklists.md
- 当前需求、功能、API、数据库、页面相关文档

请按严重程度输出问题和建议。
```

---

## 6. 文档同步检查 Prompt

```md
请先判断任务类型，并按最小阅读集读取文档。

当前任务类型：API 或字段变更

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/05-contract-sync.md

请检查当前代码变更是否需要同步更新文档。

重点检查：

1. 是否新增或修改了页面
2. 是否新增或修改了 API
3. 是否新增或修改了数据库字段
4. 是否新增或修改了业务规则
5. 是否新增或修改了环境变量
6. 是否新增依赖
7. 是否影响前端、后端、数据库字段映射

如果需要，请列出应更新的文档和具体原因。
```

---

## 7. 提交前检查 Prompt

```md
请先判断任务类型，并按最小阅读集做提交前检查。

当前任务类型：代码审查或验收

请先阅读：

- project-kit/AI_PROJECT_GUIDE.md
- project-kit/ai-guide/05-contract-sync.md
- project-kit/ai-guide/06-coding-validation.md

如需逐项核对，再读取：

- project-kit/ai-guide/08-checklists.md

请检查：

1. 是否只完成当前任务
2. 是否存在无关修改
3. 是否遗漏文档同步
4. API、前端、数据库字段是否一致
5. 是否已运行测试或提供手动验证步骤
6. 是否需要更新任务状态和变更摘要
```
