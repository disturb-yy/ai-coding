# Java Unit Test Reference

Load this reference when adding or modifying Java unit tests, working under `src/test/java`, or when implementation changes require focused Java test coverage.

## Table of Contents

- Testing Scope
- Test Framework and Style
- Mock Boundaries
- Spring and Framework Tests
- Validation Commands
- Few-Shot Examples

## Testing Scope

- Prefer the smallest class, method, package, or module boundary that proves the changed behavior.
- Add regression tests for bug fixes and boundary tests for validation or branching changes.
- Avoid broad Spring or integration tests when a focused unit test proves the behavior.
- Use slice or integration tests only when framework wiring, transactions, serialization, persistence, or security behavior is part of the contract.
- Keep tests deterministic and independent of execution order.

## Test Framework and Style

- Match the existing test stack: JUnit 4, JUnit 5, TestNG, Mockito, AssertJ, Hamcrest, Spring test, or project-specific conventions.
- Follow nearby naming style for test classes and methods.
- Reuse existing fixtures, builders, factories, test data, and helper assertions.
- Prefer explicit assertions for important fields over comparing large objects when only a few fields matter.
- For exceptions, assert the exception type and the meaningful message, error code, or domain classification used by the project.

## Mock Boundaries

- Mock external systems, repositories, clients, clocks, random sources, and slow or nondeterministic collaborators.
- Do not mock the method under test or simple value transformations that should be asserted directly.
- Prefer constructor-injected dependencies and existing test doubles when available.
- Keep mock setup close to the test case that needs it.
- Verify interactions only when the interaction is part of the behavior contract.

## Spring and Framework Tests

- Prefer plain unit tests over `@SpringBootTest` unless container wiring or framework behavior is required.
- Use existing slice-test annotations such as `@WebMvcTest`, `@DataJpaTest`, or project equivalents when they fit the behavior.
- Avoid loading the full application context for simple mapper, validator, or service logic.
- Keep test configuration local and reuse existing test profiles or fixtures.

## Validation Commands

Prefer project wrapper scripts and documented commands. Use targeted tests during iteration, then broaden when practical. If wrapper scripts are absent, use the project-installed `mvn` or `gradle` equivalent. On Windows or other constrained environments, use the available platform equivalent and report the substitution.

```bash
./mvnw -q -Dtest=ClassNameTest test
./mvnw test
./gradlew test --tests 'com.example.ClassNameTest'
./gradlew test
./gradlew check
```

If a module path is required, use the project's existing Maven module or Gradle subproject command instead of inventing a new layout.


## Few-Shot Examples

### Plain unit test for mapper regression

Input:

```text
Add a regression test for a mapper that returns UNKNOWN when upstream status is null.
```

Expected behavior:

```text
Use a plain unit test, not Spring context loading.
Reuse existing builders or fixtures if present.
Assert the specific mapped status and avoid broad object equality when only one field matters.
```

Example shape:

```java
@Test
void mapsNullStatusToUnknown() {
    var source = OrderSource.builder().status(null).build();

    var result = mapper.map(source);

    assertThat(result.status()).isEqualTo(Status.UNKNOWN);
}
```

### Mockito collaborator boundary

Input:

```text
Test that OrderService publishes an audit event after cancellation.
```

Expected behavior:

```text
Mock the repository and publisher collaborators, not the service method under test.
Assert the returned behavior first, then verify publisher interaction because it is part of the contract.
```

Example shape:

```java
when(orders.findById(orderId)).thenReturn(Optional.of(order));

service.cancel(orderId);

verify(auditPublisher).publish(any(OrderCancelledEvent.class));
```

### Exception classification test

Input:

```text
Add a test that payment authorization rejection maps to the existing domain exception.
```

Expected behavior:

```text
Assert the exception type and stable domain classification, not only that any exception is thrown.
Reuse existing fixtures and assertion style.
Avoid loading Spring context unless framework exception mapping is the behavior under test.
```

Example shape:

```java
var thrown = assertThrows(PaymentRejectedException.class, () ->
    service.authorize(paymentId)
);

assertThat(thrown.reason()).isEqualTo(RejectReason.INSUFFICIENT_FUNDS);
```
