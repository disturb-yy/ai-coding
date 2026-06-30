# /workflow-viz

## Command
`/workflow-viz [artifact_dir|workflow_id] [--open] [--theme=auto|light|dark]`

## Description
为当前 workflow 会话 artifacts 生成或刷新一个可视化 Web 页面，并在可行时启动本地网页服务，返回可直接浏览的地址。

## Usage
- `/workflow-viz`
- `/workflow-viz .agent/workflow-artifacts/<workflow_id>`
- `/workflow-viz /home/jadon/projects/go/go-stock/.agent/workflow-artifacts/feature-industry-compare-20260624-132509`
- `/workflow-viz \\wsl.localhost\Ubuntu-24.04\home\jadon\projects\go\go-stock\.agent\workflow-artifacts\feature-industry-compare-20260624-132509 --open`

## Instructions For The AI
你正在根据 workflow-agent artifacts 生成一个中文可视化工作流报告。

1. 全程使用中文：
   - 与用户沟通时使用中文。
   - 最终回复使用中文。
   - 生成的 HTML 页面中，标题、标签、按钮、状态文案、空状态、错误提示、说明文字都使用中文。
   - 保留真实文件名、命令、路径、JSON 字段名和代码标识符，不要强行翻译这些机器可读内容。
2. 解析 artifact 目录：
   - 如果参数是目录，直接使用该目录。
   - 如果参数像 workflow id，则从当前项目查找 `.agent/workflow-artifacts/<workflow_id>`。
   - 如果没有参数，使用 `.agent/workflow-artifacts/` 下最新的目录。
   - 同时支持 Linux 路径和 Windows UNC WSL 路径。
3. 读取目录中的所有 `*.json` artifact，尤其是：
   - `requirement_analysis.json`
   - `project_understanding.json`
   - `solution_design.json`
   - `implementation_result.json`
   - `test_result.json`
   - `verification_result.json`
   - `workflow_summary.json`
4. 在同一个 artifact 目录中生成 `workflow-visualization.html`。
5. HTML 必须是自包含页面：
   - 不依赖外部网络资源。
   - CSS 和 JavaScript 内联。
   - 将解析后的 artifact 数据嵌入页面。
6. 首屏就是可用报告，不要做落地页或宣传页。
7. 页面内容尽量中文化，并在数据存在时包含：
   - 工作流状态、任务类型、耗时、重试次数、整体置信度。
   - 阶段时间线：状态、技能、执行模式、置信度、说明。
   - 变更文件。
   - 验证证据。
   - 风险和后续事项。
   - 子代理交互记录。
   - 可折叠的原始 artifact JSON 面板。
8. 使用精致但克制的仪表盘布局：
   - 适配桌面和移动端。
   - 对比度可访问。
   - 只对重复条目或明确分组使用卡片/面板。
   - 除非本地已有依赖，否则不要使用外部库。
9. 生成 HTML 后，启动一个可网页浏览的本地地址：
   - 优先在 artifact 目录中启动静态文件服务。
   - 优先使用 `python3 -m http.server <port>`，没有 `python3` 时可尝试 `python -m http.server <port>`。
   - 端口优先使用 `8765`；如果被占用，依次尝试 `8766`、`8767`、`8768`、`8769`。
   - 服务应在后台运行；不要阻塞当前会话。
   - 返回可访问 URL，例如 `http://localhost:8765/workflow-visualization.html`。
   - 如果环境无法启动服务，则明确说明原因，并仍然提供 HTML 文件路径。
10. 验证输出：
   - 确认 HTML 文件存在。
   - 确认 HTML 包含 workflow id 或 artifact 目录名。
   - 确认所有已解析 artifact 文件都在页面中有展示或原始 JSON 面板。
   - 如启动了 Web 服务，尽量用本地 HTTP 请求确认页面返回成功。
11. 最终回复必须使用中文，并提供：
   - 生成的 HTML 文件路径。
   - 可浏览的本地 URL。
   - 页面包含内容的简短摘要。
   - 缺失 artifact 文件或限制说明。

## Output File
`<artifact_dir>/workflow-visualization.html`
