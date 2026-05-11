# AI Project Guide Kit

这是可以复制到其他项目中使用的 AI 开发指南包。

## 包含内容

```text
project-kit/
├── README.md
├── AGENTS.md
├── QUICKSTART.md
├── AI_PROJECT_GUIDE.md
├── SECURITY.md
├── ai-guide/
└── templates/
```

## 推荐使用方式

1. 把整个 `project-kit/` 文件夹复制到目标项目根目录。
2. 把 `project-kit/AGENTS.md` 复制一份到目标项目根目录的 `AGENTS.md`。
3. 让 AI 从目标项目根目录开始工作。
4. 新项目按 `project-kit/ai-guide/01-project-init.md` 初始化。
5. 已有项目按 `project-kit/ai-guide/09-existing-project-onboarding.md` 接入。

## 复制后的目标项目结构

```text
your-project/
├── AGENTS.md
├── project-kit/
│   ├── AI_PROJECT_GUIDE.md
│   ├── SECURITY.md
│   ├── ai-guide/
│   └── templates/
└── docs/
```

`docs/` 是目标项目自己的业务文档，由 AI 根据 `project-kit/templates/` 生成和维护。
