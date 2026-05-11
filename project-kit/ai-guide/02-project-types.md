# 项目类型与文档结构判断

> AI 快速摘要
>
> - 适用场景：新项目初始化时判断项目类型和文档结构。
> - 必须产出：项目类型判断结果、推荐文档结构、暂不需要的文档。
> - 硬性阻断：项目类型未判断清楚前，不要机械生成固定文档结构。
> - 相关模板：`requirements.template.md`、`architecture.template.md`、`tasks.template.md`
> - 可跳过场景：已有项目接入后的小功能开发、代码审查。

本文档用于指导 AI 根据项目需求判断项目类型，并生成匹配的文档结构。

## 1. 判断原则

AI 不应给所有项目套同一套文档。

必须根据是否存在以下部分来判断：

1. 前端页面
2. 后端服务
3. 数据库或持久化
4. 外部 API
5. 命令行或自动化任务

## 2. 纯前端项目

适用条件：

- 只有页面展示或本地交互。
- 不需要后端服务。
- 不需要服务端数据库。
- 数据来自静态文件、本地存储或用户输入。

推荐结构：

```text
docs/
├── requirements.md
├── frontend/
│   └── ui-spec.md
├── tasks.md
├── coding-rules.md
├── dev-workflow.md
└── decision-log.md
```

## 3. 前端 + 后端项目

适用条件：

- 有前端页面。
- 有后端 API。
- 暂不需要复杂数据库，或只使用文件、内存、第三方服务。

推荐结构：

```text
docs/
├── requirements.md
├── architecture.md
├── frontend/
│   └── ui-spec.md
├── api/
│   └── api-spec.md
├── contracts/
│   └── contract-map.md
├── tasks.md
├── coding-rules.md
├── dev-workflow.md
└── decision-log.md
```

## 4. 前端 + 后端 + 数据库项目

适用条件：

- 有前端页面。
- 有后端 API。
- 有持久化数据。
- 需要表、字段、关系、索引或迁移。

推荐结构：

```text
docs/
├── requirements.md
├── architecture.md
├── frontend/
│   └── ui-spec.md
├── api/
│   └── api-spec.md
├── database/
│   └── data-model.md
├── contracts/
│   └── contract-map.md
├── tasks.md
├── coding-rules.md
├── dev-workflow.md
├── test-plan.md
└── decision-log.md
```

## 5. 后端/API 项目

适用条件：

- 不需要前端页面。
- 主要提供 API、服务、任务处理或集成能力。

推荐结构：

```text
docs/
├── requirements.md
├── architecture.md
├── api/
│   └── api-spec.md
├── tasks.md
├── coding-rules.md
├── dev-workflow.md
├── test-plan.md
└── decision-log.md
```

## 6. 脚本/工具类项目

适用条件：

- 不需要 Web 页面。
- 不需要长期运行的后端服务。
- 主要用于命令行、自动化、数据处理、文件转换。

推荐结构：

```text
docs/
├── requirements.md
├── usage-spec.md
├── tasks.md
├── coding-rules.md
├── dev-workflow.md
├── test-plan.md
└── decision-log.md
```

## 7. 判断结果输出格式

AI 判断项目类型后，必须输出：

```md
## 项目类型判断

### 判断结果

### 判断理由

### 需要生成的文档

### 暂不需要的文档
```
