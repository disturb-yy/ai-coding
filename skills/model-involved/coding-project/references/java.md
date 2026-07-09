# Java Coding Reference

Load this reference when editing Java code or when the project contains `pom.xml`, `build.gradle`, `settings.gradle`, or relevant `*.java` files.

## Table of Contents

- Orientation Checklist
- Code Conventions
- Testing
- Validation Commands
- Few-Shot Examples

## Orientation Checklist

- Detect the build system: Maven (`pom.xml`) or Gradle (`build.gradle`, `settings.gradle`, wrapper scripts).
- Identify Java version, dependency management, plugins, and test framework from build files.
- Match existing package structure, layering, naming, dependency injection style, and framework conventions.
- Check nearby tests before adding new test utilities or fixtures.
- Prefer project-local services, mappers, validators, clients, and exception types over new abstractions.
- Respect generated sources and annotation processors; regenerate only through the build tool when required.

## Code Conventions

- Keep package names lowercase and aligned with directory structure.
- Use existing framework idioms for dependency injection, transactions, validation, logging, and configuration.
- Prefer constructor injection when the project already uses it.
- Keep public APIs stable unless the task explicitly requires a contract change.
- Use domain-specific exception and error-response patterns already present in the project.
- Avoid swallowing exceptions; add context where local conventions support wrapping or custom exceptions.
- Keep nullability handling consistent with project style: `Optional`, annotations, validation, or explicit checks.
- Keep methods small enough to read, but avoid extracting one-off helpers that obscure simple logic.

## Testing

- Match the existing test stack: JUnit 4, JUnit 5, TestNG, Mockito, AssertJ, Spring test, or other local conventions.
- Put tests in the corresponding package under `src/test/java` unless the project uses another layout.
- Add focused unit tests for business logic and targeted slice/integration tests only when framework behavior is involved.
- Name tests according to nearby patterns.
- Prefer existing fixtures, builders, factories, and test containers over creating new ones.

## Validation Commands

Prefer wrapper scripts and project docs. When absent, choose the relevant command. If wrapper scripts are absent, use the project-installed `mvn` or `gradle` equivalent. On Windows or other constrained environments, use the available platform equivalent and report the substitution:

```bash
./mvnw test
./mvnw -q test
./mvnw -q -DskipTests package
./gradlew test
./gradlew build
./gradlew check
```

Use targeted test filters during iteration when supported, then run the broader relevant command before finishing when practical.


## Few-Shot Examples

### Preserve constructor injection in service change

Input:

```text
Add an audit publisher call after an order is cancelled.
```

Expected behavior:

```text
Check existing dependency injection style before editing.
Add the collaborator through the constructor if the class already uses constructor injection.
Keep transaction, logging, and exception patterns consistent with nearby methods.
```

Example shape:

```diff
- public OrderService(OrderRepository orders) {
+ public OrderService(OrderRepository orders, AuditPublisher auditPublisher) {
      this.orders = orders;
+     this.auditPublisher = auditPublisher;
  }
```

### Keep null handling aligned with project style

Input:

```text
Fix the mapper when upstream status is null.
```

Expected behavior:

```text
Inspect nearby mapper methods for null handling style.
Use existing default, Optional, validation, or exception conventions.
Add a focused regression test in the matching test package.
```

Example shape:

```diff
- return Status.valueOf(source.getStatus());
+ if (source.getStatus() == null) {
+     return Status.UNKNOWN;
+ }
+ return Status.valueOf(source.getStatus());
```

### Preserve domain exception behavior

Input:

```text
Return the existing domain error when payment authorization is rejected.
```

Expected behavior:

```text
Find the project's established exception and error-response mapping.
Do not introduce a generic RuntimeException when a domain exception exists.
Preserve transaction and logging behavior from nearby service methods.
Add a focused test for the rejected authorization path.
```

Example shape:

```diff
- throw new RuntimeException("payment rejected");
+ throw new PaymentRejectedException(paymentId, reason);
```
