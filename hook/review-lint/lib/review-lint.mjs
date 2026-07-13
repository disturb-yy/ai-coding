import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const DEFAULT_POLICY = fileURLToPath(
  new URL("../config/default.review-policy.yaml", import.meta.url),
);

const SKIPPED_DIRECTORIES = new Set([
  ".git",
  ".gradle",
  ".idea",
  "build",
  "dist",
  "node_modules",
  "out",
  "target",
  "vendor",
]);

const CONTROL_WORDS = new Set([
  "catch",
  "do",
  "for",
  "if",
  "new",
  "return",
  "switch",
  "synchronized",
  "throw",
  "while",
]);

function parseScalar(raw) {
  const value = raw.trim();
  if (/^(true|false)$/i.test(value)) return value.toLowerCase() === "true";
  if (/^-?\d+$/.test(value)) return Number(value);
  if (
    (value.startsWith('"') && value.endsWith('"')) ||
    (value.startsWith("'") && value.endsWith("'"))
  ) {
    return value.slice(1, -1);
  }
  return value;
}

function stripYamlComment(line) {
  let quote = null;
  for (let index = 0; index < line.length; index += 1) {
    const char = line[index];
    if ((char === '"' || char === "'") && line[index - 1] !== "\\") {
      quote = quote === char ? null : quote || char;
    }
    if (char === "#" && quote === null) return line.slice(0, index);
  }
  return line;
}

export function parsePolicy(text) {
  const trimmed = text.trim();
  if (trimmed.startsWith("{")) return validatePolicy(JSON.parse(trimmed));

  const policy = { version: 1, languages: {} };
  let language = null;
  let listKey = null;

  for (const [lineIndex, original] of text.split(/\r?\n/).entries()) {
    const line = stripYamlComment(original).replace(/\s+$/, "");
    if (!line.trim()) continue;
    if (line.includes("\t")) {
      throw new Error(`policy line ${lineIndex + 1}: tabs are not supported`);
    }

    const indent = line.length - line.trimStart().length;
    const content = line.trim();

    if (indent === 0 && content.startsWith("version:")) {
      policy.version = parseScalar(content.slice("version:".length));
      continue;
    }
    if (indent === 0 && content.startsWith("structure_scope:")) {
      policy.structure_scope = parseScalar(content.slice("structure_scope:".length));
      continue;
    }
    if (indent === 0 && content === "languages:") continue;

    const languageMatch = content.match(/^(go|java):$/);
    if (indent === 2 && languageMatch) {
      language = languageMatch[1];
      listKey = null;
      policy.languages[language] = {};
      continue;
    }

    if (!language) {
      throw new Error(`policy line ${lineIndex + 1}: expected go or java section`);
    }

    const keyMatch = content.match(/^([a-z_]+):(?:\s*(.*))?$/);
    if (indent === 4 && keyMatch) {
      const [, key, rawValue] = keyMatch;
      if (!rawValue) {
        if (key !== "exclude") {
          throw new Error(`policy line ${lineIndex + 1}: ${key} requires a value`);
        }
        policy.languages[language][key] = [];
        listKey = key;
      } else {
        policy.languages[language][key] = parseScalar(rawValue);
        listKey = null;
      }
      continue;
    }

    if (indent === 6 && content.startsWith("- ") && listKey) {
      policy.languages[language][listKey].push(parseScalar(content.slice(2)));
      continue;
    }

    throw new Error(`policy line ${lineIndex + 1}: unsupported policy syntax`);
  }

  return validatePolicy(policy);
}

function validatePolicy(policy) {
  if (!policy || typeof policy !== "object") throw new Error("policy must be an object");
  if (policy.version !== 1) throw new Error(`unsupported policy version: ${policy.version}`);
  policy.structure_scope ??= "full";
  if (!["full", "changed"].includes(policy.structure_scope)) {
    throw new Error("structure_scope must be 'full' or 'changed'");
  }
  if (!policy.languages || typeof policy.languages !== "object") {
    throw new Error("policy.languages is required");
  }

  for (const language of ["go", "java"]) {
    const config = policy.languages[language];
    if (!config) continue;
    config.enabled ??= true;
    config.exclude ??= [];
    for (const key of ["max_function_lines", "max_cyclomatic_complexity"]) {
      if (!Number.isInteger(config[key]) || config[key] < 1) {
        throw new Error(`languages.${language}.${key} must be a positive integer`);
      }
    }
    config.min_test_coverage ??= 80;
    if (!Number.isInteger(config.min_test_coverage) || config.min_test_coverage < 0 || config.min_test_coverage > 100) {
      throw new Error(`languages.${language}.min_test_coverage must be an integer from 0 to 100`);
    }
    if (!Array.isArray(config.exclude) || config.exclude.some((item) => typeof item !== "string")) {
      throw new Error(`languages.${language}.exclude must be a string list`);
    }
  }
  return policy;
}

export function findPolicy(startDirectory, explicitPath) {
  if (explicitPath) return path.resolve(startDirectory, explicitPath);
  let current = path.resolve(startDirectory);
  while (true) {
    const candidate = path.join(current, ".review-policy.yaml");
    if (fs.existsSync(candidate)) return candidate;
    const parent = path.dirname(current);
    if (parent === current) break;
    current = parent;
  }
  return DEFAULT_POLICY;
}

function globMatches(pattern, relativePath) {
  const normalized = relativePath.replaceAll(path.sep, "/");
  let regex = "^";
  for (let index = 0; index < pattern.length; index += 1) {
    const char = pattern[index];
    if (char === "*" && pattern[index + 1] === "*") {
      if (pattern[index + 2] === "/") {
        regex += "(?:.*/)?";
        index += 2;
      } else {
        regex += ".*";
        index += 1;
      }
    } else if (char === "*") {
      regex += "[^/]*";
    } else if (char === "?") {
      regex += "[^/]";
    } else {
      regex += char.replace(/[|\\{}()[\]^$+?.]/g, "\\$&");
    }
  }
  return new RegExp(`${regex}$`).test(normalized);
}

function collectFiles(rootDirectory) {
  const files = [];
  const visit = (directory) => {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
      if (entry.isSymbolicLink()) continue;
      const absolute = path.join(directory, entry.name);
      if (entry.isDirectory()) {
        if (!SKIPPED_DIRECTORIES.has(entry.name)) visit(absolute);
      } else if (entry.isFile() && (entry.name.endsWith(".go") || entry.name.endsWith(".java"))) {
        files.push(absolute);
      }
    }
  };
  visit(rootDirectory);
  return files;
}

export function sanitizeSource(source) {
  let state = "normal";
  let quote = null;
  let output = "";

  for (let index = 0; index < source.length; index += 1) {
    const char = source[index];
    const next = source[index + 1];

    if (state === "line-comment") {
      if (char === "\n") {
        state = "normal";
        output += "\n";
      } else output += " ";
      continue;
    }
    if (state === "block-comment") {
      if (char === "*" && next === "/") {
        output += "  ";
        index += 1;
        state = "normal";
      } else output += char === "\n" ? "\n" : " ";
      continue;
    }
    if (state === "string") {
      if (char === "\n" && quote !== "`") {
        state = "normal";
        quote = null;
        output += "\n";
      } else if (char === quote && (quote === "`" || source[index - 1] !== "\\")) {
        output += " ";
        state = "normal";
        quote = null;
      } else {
        output += char === "\n" ? "\n" : " ";
      }
      continue;
    }

    if (char === "/" && next === "/") {
      output += "  ";
      index += 1;
      state = "line-comment";
    } else if (char === "/" && next === "*") {
      output += "  ";
      index += 1;
      state = "block-comment";
    } else if (char === '"' || char === "'" || char === "`") {
      output += " ";
      state = "string";
      quote = char;
    } else {
      output += char;
    }
  }
  return output;
}

function closingBrace(source, openingIndex) {
  let depth = 0;
  for (let index = openingIndex; index < source.length; index += 1) {
    if (source[index] === "{") depth += 1;
    if (source[index] === "}") {
      depth -= 1;
      if (depth === 0) return index;
    }
  }
  return -1;
}

function lineNumberAt(source, index) {
  return source.slice(0, index).split("\n").length;
}

function metricFor(source, start, openingBrace, end, language) {
  const span = source.slice(start, end + 1);
  const body = source.slice(openingBrace + 1, end);
  const effectiveLines = span.split("\n").filter((line) => line.trim()).length;
  const keywords = language === "go" ? ["if", "for", "case"] : ["if", "for", "while", "case", "catch"];
  let complexity = 1;
  for (const keyword of keywords) {
    complexity += (body.match(new RegExp(`\\b${keyword}\\b`, "g")) || []).length;
  }
  complexity += (body.match(/&&|\|\|/g) || []).length;
  if (language === "java") complexity += (body.match(/\?/g) || []).length;
  return { effectiveLines, complexity };
}

function goFunctions(source) {
  const functions = [];
  const expression = /\bfunc\s+(?:\([^)]*\)\s*)?([A-Za-z_]\w*)\s*\([^)]*\)[^{]*\{/g;
  for (const match of source.matchAll(expression)) {
    const openingBrace = match.index + match[0].lastIndexOf("{");
    const end = closingBrace(source, openingBrace);
    if (end < 0) continue;
    functions.push({ name: match[1], start: match.index, openingBrace, end });
  }
  return functions;
}

function javaFunctions(source) {
  const functions = [];
  const expression = /(?:^|[;{}]\s*|\n\s*)(?:(?:public|protected|private|static|final|abstract|native|strictfp|default|synchronized)\s+)*(?:<[^>{}]+>\s*)?(?:[A-Za-z_$][\w$<>,.?\[\]\s]*\s+)?([A-Za-z_$][\w$]*)\s*\([^;{}]*\)\s*(?:throws\s+[^{}]+)?\{/gm;
  for (const match of source.matchAll(expression)) {
    const name = match[1];
    if (CONTROL_WORDS.has(name)) continue;
    const declaration = match[0].trim();
    if (new RegExp(`\\bnew\\s+${name}\\s*\\(`).test(declaration)) continue;
    const leadingDelimiter = match[0].match(/^(?:[;{}]\s*|\n\s*)/);
    const start = match.index + (leadingDelimiter?.[0].length || 0);
    const openingBrace = match.index + match[0].lastIndexOf("{");
    const end = closingBrace(source, openingBrace);
    if (end < 0) continue;
    functions.push({ name, start, openingBrace, end });
  }
  return functions;
}

export function analyzeSource(source, language) {
  const sanitized = sanitizeSource(source);
  const functions = language === "go" ? goFunctions(sanitized) : javaFunctions(sanitized);
  return functions.map((entry) => ({
    name: entry.name,
    line: lineNumberAt(sanitized, entry.start),
    ...metricFor(sanitized, entry.start, entry.openingBrace, entry.end, language),
  }));
}

function runCommand(command, args, cwd) {
  const result = spawnSync(command, args, {
    cwd,
    encoding: "utf8",
    shell: false,
  });
  return {
    status: result.status,
    stdout: result.stdout || "",
    stderr: result.stderr || "",
    error: result.error,
  };
}

function temporaryDirectory(prefix) {
  const candidates = process.platform === "win32" ? [os.tmpdir()] : ["/tmp", os.tmpdir()];
  for (const candidate of new Set(candidates)) {
    try {
      return fs.mkdtempSync(path.join(candidate, prefix));
    } catch {
      // Try the next runtime-specific temporary directory.
    }
  }
  throw new Error("could not create a temporary directory for coverage data");
}

function findMarkerDirectories(rootDirectory, markerNames, { stopAtMarker = false } = {}) {
  const directories = [];
  const visit = (directory) => {
    const entries = fs.readdirSync(directory, { withFileTypes: true });
    if (entries.some((entry) => entry.isFile() && markerNames.includes(entry.name))) {
      directories.push(directory);
      if (stopAtMarker) return;
    }
    for (const entry of entries) {
      if (entry.isDirectory() && !SKIPPED_DIRECTORIES.has(entry.name)) {
        visit(path.join(directory, entry.name));
      }
    }
  };
  visit(rootDirectory);
  return directories;
}

function findJaCoCoReports(rootDirectory) {
  const reports = [];
  const visit = (directory) => {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
      const absolute = path.join(directory, entry.name);
      if (entry.isDirectory()) {
        if (![".git", "node_modules", "vendor"].includes(entry.name)) visit(absolute);
      } else if (entry.isFile() && (entry.name === "jacoco.xml" || entry.name === "jacocoTestReport.xml")) {
        reports.push(absolute);
      }
    }
  };
  visit(rootDirectory);
  return reports;
}

export function parseJaCoCoLineCoverage(xml) {
  const counters = [...xml.matchAll(/<counter\s+type="LINE"\s+missed="(\d+)"\s+covered="(\d+)"\s*\/>/g)];
  if (!counters.length) throw new Error("JaCoCo XML has no LINE counter");
  const [, missed, covered] = counters.at(-1);
  const total = Number(missed) + Number(covered);
  if (!total) return 100;
  return (Number(covered) / total) * 100;
}

function commandFailure(result) {
  if (result.error) return result.error.message;
  const output = `${result.stderr}\n${result.stdout}`.trim().replace(/\s+/g, " ");
  return output ? output.slice(0, 500) : `command exited with status ${result.status}`;
}

function goCoverage(moduleDirectory, commandRunner) {
  const temporaryRoot = temporaryDirectory("review-lint-go-coverage-");
  const profile = path.join(temporaryRoot, "coverage.out");
  try {
    const testResult = commandRunner("go", ["test", "./...", `-coverprofile=${profile}`], moduleDirectory);
    if (testResult.error || testResult.status !== 0) {
      return { error: `go test failed: ${commandFailure(testResult)}` };
    }
    const coverResult = commandRunner("go", ["tool", "cover", "-func", profile], moduleDirectory);
    if (coverResult.error || coverResult.status !== 0) {
      return { error: `go tool cover failed: ${commandFailure(coverResult)}` };
    }
    const match = coverResult.stdout.match(/total:\s+\(statements\)\s+(\d+(?:\.\d+)?)%/);
    if (!match) return { error: "could not parse total coverage from go tool cover" };
    return { percent: Number(match[1]) };
  } finally {
    fs.rmSync(temporaryRoot, { recursive: true, force: true });
  }
}

function javaCoverageCommand(moduleDirectory) {
  const windows = process.platform === "win32";
  const mavenWrapper = path.join(moduleDirectory, windows ? "mvnw.cmd" : "mvnw");
  const gradleWrapper = path.join(moduleDirectory, windows ? "gradlew.bat" : "gradlew");
  if (fs.existsSync(mavenWrapper)) return [mavenWrapper, ["test", "jacoco:report"]];
  if (fs.existsSync(path.join(moduleDirectory, "pom.xml"))) return ["mvn", ["test", "jacoco:report"]];
  if (fs.existsSync(gradleWrapper)) return [gradleWrapper, ["test", "jacocoTestReport"]];
  if (
    fs.existsSync(path.join(moduleDirectory, "build.gradle")) ||
    fs.existsSync(path.join(moduleDirectory, "build.gradle.kts"))
  ) return ["gradle", ["test", "jacocoTestReport"]];
  return null;
}

function javaCoverage(moduleDirectory, commandRunner) {
  const command = javaCoverageCommand(moduleDirectory);
  if (!command) return [{ error: "could not find a Maven or Gradle build definition" }];
  const startedAt = Date.now();
  const result = commandRunner(command[0], command[1], moduleDirectory);
  if (result.error || result.status !== 0) {
    return [{ error: `Java coverage command failed: ${commandFailure(result)}` }];
  }
  const reports = findJaCoCoReports(moduleDirectory).filter((report) => fs.statSync(report).mtimeMs >= startedAt - 2000);
  if (!reports.length) {
    return [{ error: "JaCoCo XML report was not generated by the coverage command" }];
  }
  return reports.map((report) => {
    try {
      return { percent: parseJaCoCoLineCoverage(fs.readFileSync(report, "utf8")), report };
    } catch (error) {
      return { error: `${path.relative(moduleDirectory, report)}: ${error.message}` };
    }
  });
}

function coverageViolation(language, moduleDirectory, root, limit, outcome) {
  const module = path.relative(root, moduleDirectory).replaceAll(path.sep, "/") || ".";
  if (outcome.error) {
    return {
      type: "test-coverage-command",
      language,
      path: module,
      limit,
      detail: outcome.error,
    };
  }
  if (outcome.percent < limit) {
    return {
      type: "test-coverage",
      language,
      path: module,
      limit,
      actual: outcome.percent,
      report: outcome.report,
    };
  }
  return null;
}

function extractPaths(value, into) {
  if (typeof value === "string") {
    if (value.endsWith(".go") || value.endsWith(".java")) into.add(value);
    return;
  }
  if (Array.isArray(value)) {
    for (const item of value) extractPaths(item, into);
    return;
  }
  if (value && typeof value === "object") {
    for (const key of Object.keys(value)) {
      if (key === "content" || key === "old_string" || key === "new_string" || key === "command" || key === "reason") continue;
      extractPaths(value[key], into);
    }
  }
}

export function changedGoJavaFiles(input, cwd) {
  const paths = new Set();
  extractPaths(input, paths);
  const resolved = [];
  for (const raw of paths) {
    if (!path.isAbsolute(raw)) continue;
    try {
      const rel = path.relative(cwd, raw);
      if (rel && !rel.startsWith("..")) resolved.push(raw);
    } catch {
      // ignore unresolvable paths
    }
  }
  return resolved;
}

export function reviewProject({ cwd = process.cwd(), policyPath, files, commandRunner = runCommand } = {}) {
  const root = path.resolve(cwd);
  const resolvedPolicy = findPolicy(root, policyPath);
  const policy = parsePolicy(fs.readFileSync(resolvedPolicy, "utf8"));
  const violations = [];
  const languagesWithProductionCode = new Set();

  const allFiles = collectFiles(root);
  const checkSet = files && files.length > 0
    ? new Set(files.map((f) => path.resolve(root, f)))
    : null;
  const shouldCheckStructure = policy.structure_scope === "full" || checkSet !== null;
  const checked = [];

  for (const absolutePath of allFiles) {
    const language = absolutePath.endsWith(".go") ? "go" : "java";
    const config = policy.languages[language];
    if (!config?.enabled) continue;
    const relativePath = path.relative(root, absolutePath).replaceAll(path.sep, "/");
    if (config.exclude.some((pattern) => globMatches(pattern, relativePath))) continue;

    if (checkSet && !checkSet.has(absolutePath)) continue;
    checked.push(relativePath);

    languagesWithProductionCode.add(language);
    if (shouldCheckStructure) {
      const source = fs.readFileSync(absolutePath, "utf8");
      for (const metric of analyzeSource(source, language)) {
        if (metric.effectiveLines > config.max_function_lines) {
          violations.push({
            type: "function-lines",
            language,
            path: relativePath,
            ...metric,
            limit: config.max_function_lines,
          });
        }
        if (metric.complexity > config.max_cyclomatic_complexity) {
          violations.push({
            type: "cyclomatic-complexity",
            language,
            path: relativePath,
            ...metric,
            limit: config.max_cyclomatic_complexity,
          });
        }
      }
    }
  }

  if (languagesWithProductionCode.has("go") && policy.languages.go.min_test_coverage > 0) {
    const modules = findMarkerDirectories(root, ["go.mod"]);
    for (const moduleDirectory of modules.length ? modules : [root]) {
      const violation = coverageViolation(
        "go",
        moduleDirectory,
        root,
        policy.languages.go.min_test_coverage,
        goCoverage(moduleDirectory, commandRunner),
      );
      if (violation) violations.push(violation);
    }
  }

  if (languagesWithProductionCode.has("java") && policy.languages.java.min_test_coverage > 0) {
    const modules = findMarkerDirectories(root, ["pom.xml", "build.gradle", "build.gradle.kts"], { stopAtMarker: true });
    for (const moduleDirectory of modules.length ? modules : [root]) {
      for (const outcome of javaCoverage(moduleDirectory, commandRunner)) {
        const violation = coverageViolation(
          "java",
          moduleDirectory,
          root,
          policy.languages.java.min_test_coverage,
          outcome,
        );
        if (violation) violations.push(violation);
      }
    }
  }

  return { policyPath: resolvedPolicy, root, violations, checked };
}

export function formatReview(result) {
  if (!result.violations.length) return "review-lint: passed";

  const hints = [];
  const seen = {};

  for (const violation of result.violations) {
    if (violation.type === "function-lines") {
      hints.push(
        `${violation.path}:${violation.line} ${violation.language} ${violation.name}: ` +
        `effective lines ${violation.effectiveLines} > ${violation.limit} ` +
        `(exceeds by ${violation.effectiveLines - violation.limit} lines)` +
        `\n  FIX: Split this function into smaller helper functions. ` +
        `Extract logical blocks into separate well-named functions, each doing one thing. ` +
        `Target: each function ≤ ${violation.limit} effective lines.`,
      );
      seen["function-lines"] = true;
    } else if (violation.type === "cyclomatic-complexity") {
      hints.push(
        `${violation.path}:${violation.line} ${violation.language} ${violation.name}: ` +
        `cyclomatic complexity ${violation.complexity} > ${violation.limit} ` +
        `(exceeds by ${violation.complexity - violation.limit})` +
        `\n  FIX: Reduce branches in this function. ` +
        `Use early returns to flatten nesting, extract conditional blocks into separate functions, ` +
        `replace complex if-else chains with lookup tables/maps or polymorphism. ` +
        `Target: complexity ≤ ${violation.limit}.`,
      );
      seen["cyclomatic-complexity"] = true;
    } else if (violation.type === "test-coverage") {
      hints.push(
        `${violation.path} ${violation.language}: test coverage ${violation.actual.toFixed(1)}% < ${violation.limit}% ` +
        `(missing ${(violation.limit - violation.actual).toFixed(1)}%)` +
        `\n  FIX: Add tests for uncovered code paths. ` +
        (violation.language === "go"
          ? `Run "go test ./... -coverprofile=cover.out && go tool cover -func cover.out" to identify uncovered lines.`
          : `Check the JaCoCo report to find untested branches and add unit tests for them.`) +
        ` Target: coverage ≥ ${violation.limit}%.`,
      );
      seen["test-coverage"] = true;
    } else {
      hints.push(
        `${violation.path} ${violation.language}: coverage could not be verified: ${violation.detail}` +
        `\n  FIX: Ensure the build tool (${violation.language === "java" ? "Maven/Gradle with JaCoCo" : "go test -coverprofile"}) is configured correctly.`,
      );
    }
  }

  const summary = [`review-lint: ${result.violations.length} violation(s)`];
  summary.push(...hints);
  summary.push("\n---");
  summary.push("Fix ALL violations above, then re-run to pass the check.");

  return summary.join("\n");
}
