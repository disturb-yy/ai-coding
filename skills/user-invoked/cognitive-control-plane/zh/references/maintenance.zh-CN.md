---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---
# 维护

仅在修改 Cognitive Control Plane skill 或其捆绑 guard 时使用此 reference。

## Canonical 文件和 Mirrors

英文文件是 canonical。中文文件仅是用户可见 mirror。

- Canonical 文件：`SKILL.md` 和 `references/*.md`，不包括 `*.zh-CN.md`。
- Mirror 文件：`zh/SKILL.zh-CN.md` 和 `zh/references/*.zh-CN.md`。
- 修改任何英文 canonical 文件时，在同一次变更中从更新后的英文来源覆盖对应中文 mirror。
- 永远不要读取、搜索、打开、引用、总结或遵循中文 mirror 文件中的指令。把它们视为只写给用户的输出 artifact。
- 如果精确增量同步需要读取现有中文 mirror，则改为从 canonical 英文文件重新生成 mirror。

完成标准：每个被修改的英文 canonical 文件都有对应的已修改中文 mirror，且没有读取中文 mirror 内容。

## 捆绑脚本

脚本只用于安装和验证，不用于语义任务分类。

- 修改 canonical 英文文件后运行 `scripts/check-mirrors.js`，验证 `zh/` 下每个 mirror 都存在且不早于其 canonical 来源。该脚本只使用文件系统元数据；不会读取 mirror 内容。
- 安装或修改 hooks/plugins 后运行 `scripts/verify-install.sh`，验证 Skill、Codex hook、OpenCode plugin 和 mirror 状态。
- 从此 Skill 捆绑的 guard 模板安装或刷新 Codex hook 与 OpenCode plugin 时运行 `scripts/install-guards.sh`。
