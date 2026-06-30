# /review
检查当前分支的代码变更，重点关注安全漏洞和性能问题，检查语法问题，以表格形式输出，不需要任何特殊语法。

## Command
`/review`

## Description
Review the current code, selected code block, or a specific file/section and provide actionable feedback.

## Usage
`/review [target] [--focus=<area>] [--summary] [--style] [--bugs] [--tests]`

### Target examples
- `target service`
- `current file`
- `selected code`
- `file path/to/file.js`
- `section functionName`

### Options
- `--focus=<area>`: focus review on a specific topic such as performance, security, style, or architecture
- `--summary`: return a concise review summary
- `--style`: call out coding-style issues
- `--bugs`: highlight likely bugs and logic problems
- `--tests`: recommend missing tests or test improvements

## Examples
- `/review current file`
- `/review selected code --focus=performance --bugs`
- `/review file skills/commands/review.md --summary`

## Notes
Use this command inside a code review or editor environment where Codex can see the target code and provide review guidance.