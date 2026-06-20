# Skill 映射与调用规则

go-workflow 按需为每个 SubAgent 加载 Skill。采用三级优先级覆盖机制：

```
用户显式指定 (最高优先级)
  ↓ 未指定？
项目 .agent/workflow-skills.yaml
  ↓ 文件不存在？
Skill 内置 assets/workflow-skills.yaml (默认配置)
```

> **默认配置即 `assets/workflow-skills.yaml`。** 项目不需要此文件时，自动使用 Skill 内置的模板配置。

---

## 默认映射

| SubAgent | 默认 Skill | 加载时机 |
|----------|-----------|---------|
| `doc-index-reader` | `go-index` | Observe 阶段 spawn 时 |
| `code-structure-analyzer` | 无 | 代码结构工具通过 MCP 或本地回退调用 |
| `module-code-analyzer` | `go-coding` | Draft 阶段 spawn 时 |
| `test-runner` | `go-diagnose` | Evaluate 阶段 spawn 时 |

---

## 三级覆盖机制

```
用户显式指定 (最高优先级)
  ↓ 未指定？
项目配置文件 .agent/workflow-skills.yaml
  ↓ 不存在？
Skill 内置默认映射 (兜底)
```

### 级别 1：用户显式指定

用户在触发 go-workflow 时通过 `$skill-name` 语法指定：

```
用户: "$go-workflow 用 $go-coding 和 $go-diagnose 修复 PreClose bug"
用户: "$go-workflow --skills doc-index-reader=go-index,module-code-analyzer=go-coding 重构 handler"
```

主 Agent 在 Decide 阶段将用户指定的 Skill 注入 dispatch_plan，**覆盖**对应 SubAgent 的默认值。

解析规则：
- `$skill-name` 语法：提取所有 `$xxx` token，按 SubAgent 角色分类。
- `--skills` 语法：`agent_name=skill1,skill2`，精确指定。
- 未指定的 SubAgent 回退到默认映射。

### 级别 2：项目配置文件（覆盖默认）

仅在需要**不同于默认**的配置时创建。从 Skill 内置模板复制并修改：

```bash
cp ~/.codex/skills/go-workflow/assets/workflow-skills.yaml .agent/workflow-skills.yaml
```

模板包含完整的注释和可选配置项。如不需要自定义，删除此文件即可使用内置默认。

`.agent/workflow-skills.yaml` 示例：

```yaml
# .agent/workflow-skills.yaml
subagents:
  doc-index-reader:
    skills: [go-index]
  code-structure-analyzer:
    skills: []                    # 代码结构工具（MCP 或本地回退），无需额外 Skill
  module-code-analyzer:
    skills: [go-coding, go-diagnose]  # 项目需要额外的诊断能力
  test-runner:
    skills: [go-diagnose]

# 可选：全局开关
options:
  skip_structure_analysis_for_single_file: true   # 单文件任务跳过 Orient
  max_parallel_agents: 3               # Draft 阶段最大并发
```

Observe 阶段 `doc-index-reader` 读取项目索引文件时，同时检查 `.agent/workflow-skills.yaml`。若存在，其配置覆盖默认映射。

### 级别 3：Skill 内置默认配置

项目无 `.agent/workflow-skills.yaml` 时，使用 Skill 自带的 `assets/workflow-skills.yaml`：

```yaml
# 位置: ~/.codex/skills/go-workflow/assets/workflow-skills.yaml
subagents:
  doc-index-reader:
    skills: [go-index]
  code-structure-analyzer:
    skills: []
  module-code-analyzer:
    skills: [go-coding]
  test-runner:
    skills: [go-diagnose]

options:
  skip_structure_analysis_for_single_file: false
  max_parallel_agents: 3
  max_iterations: 3
  enable_log_analysis: true
  log_level: standard
```

这是 Skill 的**出厂配置**。修改此文件会影响所有使用该 Skill 的项目。


## spawn_agent 传参规范

主 Agent 调用 `spawn_agent` 时，通过 `items` 数组传递 Skill：

```json
{
  "message": "读取 domain/anomaly/http_handler.go，在 PreClose 计算失败时添加 warn 日志",
  "items": [
    {"type": "skill", "name": "go-coding", "path": "~/.codex/skills/go-coding"}
  ]
}
```

**传递规则：**

| 规则 | 说明 |
|------|------|
| 每个 Skill 一个 item | `items` 数组中每个 Skill 独立一个元素 |
| path 指向 skill 目录 | Codex 自动发现 `${CODEX_HOME}/skills/<name>` |
| SubAgent 自动加载 | `type: "skill"` 确保 Skill 在 SubAgent 上下文中生效 |
| 不传源码文件 | items 中不传 `.go` 文件——源码由 SubAgent 通过其权限自行读取 |

---

## 与 Decide 阶段的集成

Decide 阶段的 `dispatch_plan` 扩展 `skills` 字段：

```json
{
  "dispatch_plan": [
    {
      "agent": "module-code-analyzer",
      "scope": "domain/anomaly/http_handler.go",
      "task": "修复 PreClose 计算失败时添加 warn 日志",
      "skills": ["go-coding"],
      "skills_source": "default"
    }
  ]
}
```

`skills_source` 标明来源：
- `"user"` — 用户显式指定
- `"project_config"` — 项目 `.agent/workflow-skills.yaml`
- `"default"` — Skill 内置默认映射

此字段同时出现在日志中，用于审计分析：

```
[go-workflow] phase=Decide  skills_source=default  skills_mapped="module-code-analyzer→go-coding"
```

---

## 完整示例

### 场景 1：纯默认（95% 场景）

```
用户: "$go-workflow 修复 PreClose bug"
  → Observe: 检查 .agent/workflow-skills.yaml → 不存在
  → Decide: 读取 assets/workflow-skills.yaml → 使用 Skill 默认配置
  → Draft: spawn_agent 传 go-coding
  → 日志: [go-workflow] phase=Decide  skills_source=skill_asset
```

### 场景 2：用户显式覆盖（最高优先级）

```
用户: "$go-workflow 用 $go-coding 和 $go-diagnose 修复 flaky test"
  → Decide: doc-index-reader 保留默认，module-code-analyzer 覆盖为 [go-coding, go-diagnose]
  → Draft: spawn_agent 传 go-coding + go-diagnose
  → 日志: [go-workflow] phase=Decide  skills_source=user  skills_mapped="module-code-analyzer→go-coding,go-diagnose"
```

### 场景 3：项目配置覆盖默认

```
项目: .agent/workflow-skills.yaml 配置 module-code-analyzer: [go-coding, go-diagnose]
  → Observe: doc-index-reader 读取项目配置文件
  → Decide: 项目配置覆盖 Skill 内置 assets/workflow-skills.yaml
  → Draft: spawn_agent 传 go-coding + go-diagnose
  → 日志: [go-workflow] phase=Decide  skills_source=project_config
```
