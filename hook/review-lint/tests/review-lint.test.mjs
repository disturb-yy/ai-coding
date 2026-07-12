import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import test from "node:test";
import { analyzeSource, parseJaCoCoLineCoverage, parsePolicy, reviewProject } from "../lib/review-lint.mjs";
import { install } from "../scripts/install.mjs";

function temporaryDirectory(t) {
  const base = process.platform === "win32" ? os.tmpdir() : "/tmp";
  const directory = fs.mkdtempSync(path.join(base, "review-lint-"));
  t.after(() => fs.rmSync(directory, { recursive: true, force: true }));
  return directory;
}

const TEST_POLICY = `
version: 1
languages:
  go:
    enabled: true
    max_function_lines: 4
    max_cyclomatic_complexity: 2
    min_test_coverage: 0
    exclude:
      - "**/*_test.go"
  java:
    enabled: true
    max_function_lines: 4
    max_cyclomatic_complexity: 2
    min_test_coverage: 0
    exclude:
      - "**/src/test/**"
`;

test("parses the supported policy schema", () => {
  const policy = parsePolicy(TEST_POLICY);
  assert.equal(policy.languages.go.max_function_lines, 4);
  assert.deepEqual(policy.languages.java.exclude, ["**/src/test/**"]);
});

test("counts effective lines and cyclomatic complexity", () => {
  const metrics = analyzeSource(`
func createOrder(ok bool, retry bool) {
  // ignored comment

  if ok && retry {
    println("ok")
  }
}
`, "go");
  assert.equal(metrics.length, 1);
  assert.equal(metrics[0].effectiveLines, 5);
  assert.equal(metrics[0].complexity, 3);
});

test("reviews Go and Java production files and excludes tests", (t) => {
  const project = temporaryDirectory(t);
  fs.writeFileSync(path.join(project, ".review-policy.yaml"), TEST_POLICY);
  fs.writeFileSync(path.join(project, "order.go"), `
package order
func Create(ok bool, retry bool) {
  if ok && retry {
    println("ok")
  }
}
`);
  fs.writeFileSync(path.join(project, "order_test.go"), `
package order
func TestVeryLong() {
  if true { if true { if true { println("ignored") } } }
}
`);
  fs.mkdirSync(path.join(project, "src", "main", "java"), { recursive: true });
  fs.writeFileSync(path.join(project, "src", "main", "java", "Order.java"), `
class Order {
  void create(boolean ok, boolean retry) {
    if (ok || retry) {
      System.out.println("ok");
    }
  }
}
`);

  const result = reviewProject({ cwd: project });
  assert.equal(result.violations.length, 4);
  assert.ok(result.violations.every((violation) => !violation.path.includes("_test.go")));
  assert.deepEqual(
    new Set(result.violations.map((violation) => violation.language)),
    new Set(["go", "java"]),
  );
});

test("requires Go coverage to meet the configured threshold", (t) => {
  const project = temporaryDirectory(t);
  fs.writeFileSync(path.join(project, ".review-policy.yaml"), TEST_POLICY.replaceAll("min_test_coverage: 0", "min_test_coverage: 80"));
  fs.writeFileSync(path.join(project, "go.mod"), "module example.com/coverage\n\ngo 1.22\n");
  fs.writeFileSync(path.join(project, "coverage.go"), "package coverage\nfunc Value() int { return 1 }\n");
  const runner = (_command, args) => args[0] === "tool"
    ? { status: 0, stdout: "example.com/coverage/coverage.go:1:\tValue\t100.0%\ntotal:\t(statements)\t79.9%\n" }
    : { status: 0, stdout: "ok\n" };

  const result = reviewProject({ cwd: project, commandRunner: runner });
  const coverage = result.violations.find((violation) => violation.type === "test-coverage");
  assert.equal(coverage.actual, 79.9);
  assert.equal(coverage.limit, 80);
});

test("runs the real Go coverage toolchain when Go is available", {
  skip: spawnSync("go", ["version"], { stdio: "ignore" }).status !== 0,
}, (t) => {
  const project = temporaryDirectory(t);
  fs.writeFileSync(path.join(project, ".review-policy.yaml"), TEST_POLICY.replaceAll("min_test_coverage: 0", "min_test_coverage: 80"));
  fs.writeFileSync(path.join(project, "go.mod"), "module example.com/coverage\n\ngo 1.22\n");
  fs.writeFileSync(path.join(project, "calculator.go"), "package coverage\nfunc Add(left, right int) int { return left + right }\n");
  fs.writeFileSync(path.join(project, "calculator_test.go"), "package coverage\nimport \"testing\"\nfunc TestAdd(t *testing.T) { if Add(1, 2) != 3 { t.Fatal(\"unexpected sum\") } }\n");

  const result = reviewProject({ cwd: project });
  const coverageViolations = result.violations.filter((violation) => violation.type.startsWith("test-coverage"));
  if (coverageViolations.some((violation) => /spawnSync go EPERM/.test(violation.detail || ""))) {
    t.skip("the current sandbox blocks Node from starting the Go toolchain");
    return;
  }
  assert.deepEqual(coverageViolations, [], JSON.stringify(result));
});

test("requires a fresh JaCoCo report and enforces its line coverage", (t) => {
  const project = temporaryDirectory(t);
  fs.writeFileSync(path.join(project, ".review-policy.yaml"), TEST_POLICY.replaceAll("min_test_coverage: 0", "min_test_coverage: 80"));
  fs.writeFileSync(path.join(project, "pom.xml"), "<project/>\n");
  fs.mkdirSync(path.join(project, "src", "main", "java"), { recursive: true });
  fs.writeFileSync(path.join(project, "src", "main", "java", "Sample.java"), "class Sample { int value() { return 1; } }\n");
  const reportDirectory = path.join(project, "target", "site", "jacoco");
  fs.mkdirSync(reportDirectory, { recursive: true });
  fs.writeFileSync(path.join(reportDirectory, "jacoco.xml"), "<report><counter type=\"LINE\" missed=\"25\" covered=\"75\"/></report>");

  const result = reviewProject({ cwd: project, commandRunner: () => ({ status: 0, stdout: "ok\n" }) });
  const coverage = result.violations.find((violation) => violation.type === "test-coverage");
  assert.equal(coverage.actual, 75);
  assert.equal(coverage.limit, 80);
  assert.equal(parseJaCoCoLineCoverage("<counter type=\"LINE\" missed=\"0\" covered=\"4\"/>"), 100);
});

test("does not pass Java coverage when JaCoCo XML is missing", (t) => {
  const project = temporaryDirectory(t);
  fs.writeFileSync(path.join(project, ".review-policy.yaml"), TEST_POLICY.replaceAll("min_test_coverage: 0", "min_test_coverage: 80"));
  fs.writeFileSync(path.join(project, "pom.xml"), "<project/>\n");
  fs.mkdirSync(path.join(project, "src", "main", "java"), { recursive: true });
  fs.writeFileSync(path.join(project, "src", "main", "java", "Sample.java"), "class Sample { int value() { return 1; } }\n");

  const result = reviewProject({ cwd: project, commandRunner: () => ({ status: 0, stdout: "ok\n" }) });
  const coverage = result.violations.find((violation) => violation.type === "test-coverage-command");
  assert.match(coverage.detail, /JaCoCo XML report was not generated/);
});

test("installer registers all adapters idempotently", (t) => {
  const home = temporaryDirectory(t);
  install({ home });
  install({ home });

  const codex = JSON.parse(fs.readFileSync(path.join(home, ".codex", "hooks.json"), "utf8"));
  const claude = JSON.parse(fs.readFileSync(path.join(home, ".claude", "settings.json"), "utf8"));
  assert.equal(codex.hooks.PostToolUse.length, 1);
  assert.equal(codex.hooks.Stop.length, 1);
  assert.equal(claude.hooks.PostToolUse.length, 1);
  assert.equal(claude.hooks.Stop.length, 1);
  assert.ok(fs.existsSync(path.join(home, ".config", "opencode", "plugins", "review-lint.mjs")));
  assert.ok(fs.existsSync(path.join(home, ".local", "share", "review-lint", "bin", "review-lint.mjs")));
});
