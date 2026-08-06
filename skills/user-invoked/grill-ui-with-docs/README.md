# Grill UI with Docs 使用指南

`grill-ui-with-docs` 用于把一个模糊的页面或流程想法，收敛成经过用户确认、可交给编码技能实现的设计交接包。

它是用户调用型技能：在 Codex 中显式输入 `$grill-ui-with-docs` 使用。

## 适用场景

- 新页面、仪表盘、管理后台或关键用户流程的设计。
- 现有页面需要重新组织信息、交互或视觉系统。
- 实现前需要低保真原型和明确的用户确认。
- 多个页面需要共享颜色、排版、间距和组件规则。

纯后端改动、只改一行样式、或已经有完整批准设计稿的实现任务，不需要使用此技能。

## 怎么开始

提供页面意图即可；不需要先写完整 PRD。

```text
使用 $grill-ui-with-docs，为个人投资者设计预警中心页面。
```

也可以补充现有项目与约束：

```text
使用 $grill-ui-with-docs，为现有 React 项目新增预警中心。
用户需要查看触发的预警，也要管理规则；优先桌面端，但必须支持手机查看。
```

技能会先检查项目中已有的路由、组件、设计 token、接口和术语。能从项目获得的信息不会重复询问。

## 你需要确认什么

它每次只处理一个设计决策，并附带推荐与理由。通常会依次确认：

1. 目标用户、页面范围和核心任务。
2. 信息架构、入口和任务流。
3. 页面结构、信息层级、密度、表格或图表方式。
4. 加载、空数据、错误、权限和破坏性操作状态。
5. 响应式、键盘操作和动效约束。
6. 视觉方向与可复用组件规则。

示例：

```text
当前决策：预警中心的主要工作方式

推荐：以“查看与处理已触发预警”为主，同时支持管理预警规则。

理由：个人投资者首先需要知道哪些关注标的出现了值得行动的变化；
规则创建与编辑是次级任务，不应占满首屏。

是否接受这个方向？
```

你可以直接接受、选择替代方案，或说明新的约束。技能会把已接受且会影响后续设计的决定写入相应文档。

## Wireframe Gate

在生成可实现的设计系统之前，技能会先给出 Markdown/ASCII 低保真原型。该原型包含：

- 页面目标与主任务；
- 各区域布局、组件树和信息层级；
- 主交互流程；
- 加载、空数据、错误和权限状态；
- 响应式与可访问性约束；
- 已确认决策与待确认假设。

只有你明确批准当前 wireframe 后，技能才会进入设计系统和实现交接阶段。你可以回答“批准 wireframe”或指出具体修改点。

## 最终产物

技能优先沿用项目既有文档结构。项目没有约定时，默认创建：

```text
docs/design/
├── DESIGN.md
└── pages/
    └── <page>.md
```

`DESIGN.md` 是跨页面共享的设计系统，采用 Google `design.md` 风格：YAML frontmatter 存可验证 token，Markdown 正文说明其视觉意图和使用规则。

页面说明文件保存该页面的 wireframe、任务流、页面布局与对全局系统的例外。业务术语仍归 `CONTEXT.md`，不可逆的架构取舍仍归 ADR；不要把它们复制到设计文档。

完成后会得到类似的交接：

```yaml
status: design-approved
design_system: docs/design/DESIGN.md
page_specification: docs/design/pages/alert-center.md
implementation_allowed: true
acceptance_checks:
  - approved wireframe is reflected in the page structure
  - loading, empty, error, and permission states are implemented
  - shared tokens and component rules are respected
```

## DESIGN.md 校验

若项目已安装 `@google/design.md`，或你同意下载该 CLI，技能会运行：

```bash
npx @google/design.md lint docs/design/DESIGN.md
```

它检查 token 引用、组件前景/背景对比度、排版 token 与 Markdown 章节顺序。该规范仍处于 alpha 阶段，升级时应重新审阅格式与 lint 结果。

## 边界

本技能负责设计访谈、文档与交接；生产代码由后续明确调用的前端编码技能实现。若可用的 UI/UX 知识技能已安装，它会用于给出视觉、可访问性、图表与交互建议；未安装时，流程仍可独立完成。
