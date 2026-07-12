# review-lint：Codex、Claude Code、OpenCode、CodeAgent 共用的 Go/Java 代码检查 Hook

`review-lint` 把代码质量规则实现为一个共享检查核心，再通过四套适配器接入
Codex、Claude Code、OpenCode 和 CodeAgent。

默认规则是：

- 只检查生产代码；
- Go 排除 `*_test.go` 和 `vendor`；
- Java 排除 `src/test`、`target` 和 `build`；
- 函数或方法必须少于 50 个有效代码行，因此阈值配置为 `49`；
- 空行和纯注释行不计入函数长度；
- 圈复杂度不得超过 `10`。
- Go 与 Java 的测试覆盖率必须至少为 `80%`。

这里的 Hook 会在 Agent 修改代码后自动执行结构检查和覆盖率测试。它仍不会替代
项目原有的 `golangci-lint`、Checkstyle、PMD、SpotBugs、完整编译与集成测试。

## 目录结构

```text
review-lint/
├── adapters/
│   ├── opencode/review-lint.mjs
│   └── shared/hook.mjs
├── bin/review-lint.mjs
├── config/default.review-policy.yaml
├── lib/review-lint.mjs
├── scripts/install.mjs
└── tests/review-lint.test.mjs
```

三端执行的是同一个检查核心：

| 软件 | 接入方式 | 触发时机 |
| --- | --- | --- |
| Codex | `~/.codex/hooks.json` | `PostToolUse`、`Stop` |
| Claude Code | `~/.claude/settings.json` | `PostToolUse`、`Stop` |
| CodeAgent | `~/.cac/settings.json` | `PostToolUse`、`Stop` |
| OpenCode | `~/.config/opencode/opencode.json` 的 `plugin` 数组 + `~/.config/opencode/plugins/review-lint.mjs` | `tool.execute.after` |

## 环境要求

- Node.js 18 或更高版本；
- 在实际运行 Codex、Claude Code、OpenCode 的同一环境中安装；
- 目标代码仓库对相应 Agent 可读。

例如，如果三个工具都在 WSL 中运行，就在 WSL 中执行安装。不要在 Windows
PowerShell 中安装后，期待 WSL 内的 Agent 自动读取 Windows 用户目录中的配置。

## 一键安装

进入本目录后执行：

```bash
node scripts/install.mjs
```

安装程序会：

1. 把运行时复制到 `~/.local/share/review-lint/`；
2. 合并 Codex 的 `~/.codex/hooks.json`；
3. 合并 Claude Code 的 `~/.claude/settings.json`；
4. 合并 CodeAgent 的 `~/.cac/settings.json`；
5. 创建 OpenCode 插件
   `~/.config/opencode/plugins/review-lint.mjs`，
   并注册到 `~/.config/opencode/opencode.json` 的 `plugin` 数组；
6. 修改已有 JSON 配置前创建带时间戳的备份。

安装是幂等的。重复运行不会重复注册同一个 Hook。

### 只安装部分适配器

```bash
node scripts/install.mjs --targets codex
node scripts/install.mjs --targets claude
node scripts/install.mjs --targets cac
node scripts/install.mjs --targets opencode
node scripts/install.mjs --targets codex,opencode
```

## 在项目中启用规则

把默认策略复制到目标代码仓库根目录：

```bash
cp ~/.local/share/review-lint/config/default.review-policy.yaml \
  /path/to/your-project/.review-policy.yaml
```

也可以直接从源码目录复制：

```bash
cp config/default.review-policy.yaml /path/to/your-project/.review-policy.yaml
```

然后按项目需要修改：

```yaml
version: 1

languages:
  go:
    enabled: true
    max_function_lines: 49
    max_cyclomatic_complexity: 10
    min_test_coverage: 80
    exclude:
      - "**/*_test.go"
      - "**/vendor/**"

  java:
    enabled: true
    max_function_lines: 49
    max_cyclomatic_complexity: 10
    min_test_coverage: 80
    exclude:
      - "**/src/test/**"
      - "**/target/**"
      - "**/build/**"
```

`max_function_lines: 49` 表示严格少于 50 行。如果允许最多 50 行，应改成
`50`。

`min_test_coverage: 80` 表示测试覆盖率至少为 80%。临时设为 `0` 可以关闭该
语言的覆盖率门禁，但默认策略不会关闭它。

检查器从当前工作目录开始向父目录查找最近的 `.review-policy.yaml`。如果没有
找到，就使用安装包内的默认策略。

## 手动验证检查器

在目标项目中运行：

```bash
node ~/.local/share/review-lint/bin/review-lint.mjs --cwd "$PWD"
```

通过时输出：

```text
review-lint: passed
```

失败时输出类似：

```text
review-lint: 2 violation(s)
internal/order/service.go:28 go CreateOrder: effective lines 63 > 49
src/main/java/app/OrderService.java:41 java createOrder: cyclomatic complexity 14 > 10
.: go: test coverage 79.9% < 80%
```

需要机器可读结果时：

```bash
node ~/.local/share/review-lint/bin/review-lint.mjs --cwd "$PWD" --json
```

退出码含义：

- `0`：检查通过；
- `1`：发现规则违规；
- `2`：参数或策略配置错误。

## 各软件安装后的操作

### Codex

1. 重新启动 Codex 或新建任务，让配置重新加载；
2. 打开 `/hooks`；
3. 检查新 Hook 的来源和命令；
4. 信任当前 Hook 定义；
5. 在目标仓库修改一个 Go 或 Java 文件，确认 `PostToolUse` 会执行检查。

Codex 的 `PostToolUse` 发生在文件修改之后，因此检查失败不能回滚已经写入的
内容。它会把失败原因反馈给 Codex，要求 Codex 继续修改。`Stop` 会在 Codex
准备结束本轮任务时再次执行完整检查。

### Claude Code

1. 重新启动 Claude Code；
2. 使用 `/hooks` 检查 `PostToolUse` 和 `Stop` 注册结果；
3. 在目标项目中进行一次代码修改；
4. 确认违规信息会返回到当前会话。

### CodeAgent

1. 重新启动 CodeAgent；
2. 检查 `~/.cac/settings.json` 中已注册 `PostToolUse` 和 `Stop` Hook；
3. 在目标项目中进行一次代码修改；
4. 确认违规信息会返回到当前会话。

### OpenCode

1. 重新启动 OpenCode；
2. 执行 `opencode debug config`，确认全局插件目录已经加载；
3. 修改 Go 或 Java 文件；
4. 当 `tool.execute.after` 检测到违规时，插件会抛出错误并把检查结果反馈给
   Agent。

OpenCode 当前没有与 Codex、Claude Code 完全等价的 `Stop` Hook，因此适配器
在每次写文件或执行 Bash 工具后检查整个项目，以保证结束前已经覆盖最新修改。

## 覆盖率门禁

### Go

当项目含有 Go 生产代码时，检查器会对每个发现的 `go.mod` 模块执行：

```bash
go test ./... -coverprofile=覆盖率文件
go tool cover -func 覆盖率文件
```

并解析 `total: (statements)` 的总覆盖率。命令执行失败、没有可解析的总覆盖率
或覆盖率低于 80% 都会阻止 Hook 放行。

### Java

Java 覆盖率使用 JaCoCo 的 `LINE` 指标。检查器会按项目类型运行：

```bash
# Maven 或 Maven Wrapper
./mvnw test jacoco:report

# Gradle 或 Gradle Wrapper
./gradlew test jacocoTestReport
```

然后要求本次命令生成 JaCoCo XML 报告（通常是
`target/site/jacoco/jacoco.xml` 或 Gradle 的 `jacocoTestReport.xml`）。没有
配置 JaCoCo、命令失败、报告未更新或报告覆盖率低于 80% 都视为检查失败。

因此 Java 项目需要先在 Maven 或 Gradle 构建中启用 JaCoCo；仅有 JUnit 测试而
没有 JaCoCo 报告不会被放行。

## 修改策略后如何生效

`.review-policy.yaml` 每次检查都会重新读取，不需要重新安装，也不需要重启
Agent。

如果修改的是 Hook 脚本或检查核心，请重新运行：

```bash
node scripts/install.mjs
```

Codex 会把变更后的 Hook 视为新的定义，可能需要再次通过 `/hooks` 审查和信任。

## 开发验证

在本目录运行：

```bash
npm test
npm run check
```

也可以用临时 HOME 验证安装结果，而不改动真实用户配置：

```bash
tmp_home="$(mktemp -d)"
node scripts/install.mjs --home "$tmp_home"
find "$tmp_home" -maxdepth 5 -type f -print
```

## 当前边界

内置检查器采用轻量级源码词法分析，适合常规 Go 函数、Go 方法、Java 方法和
Java 构造函数。它不会构建完整的语言 AST，因此包含复杂代码生成语法或非常规
声明格式的项目，应继续保留语言原生工具作为最终质量门禁：

```bash
# Go
golangci-lint run
go test ./...

# Java，按项目实际构建工具选择
./mvnw verify
./gradlew check
```

推荐在 CI 中同时运行 `review-lint` 和项目原生检查命令，让 Agent 本地反馈与
合并门禁使用同一份 `.review-policy.yaml`，同时保留编译器和成熟静态分析器的
完整检查能力。

## 卸载

删除生成的 Hook 条目和文件：

1. 从 `~/.codex/hooks.json` 删除命令中包含
   `review-lint/adapters/shared/hook.mjs` 的 `PostToolUse`、`Stop` 条目；
2. 从 `~/.claude/settings.json` 删除相同条目；
3. 从 `~/.cac/settings.json` 删除相同条目；
4. 删除 `~/.config/opencode/plugins/review-lint.mjs`；
5. 删除 `~/.local/share/review-lint/`。

如果需要恢复安装前的配置，可以使用安装程序生成的
`.bak-review-lint-时间戳` 文件。
