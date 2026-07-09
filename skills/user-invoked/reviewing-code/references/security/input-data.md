# Input And Data Handling

Use this checklist for external input, data integrity, and injection risk.

## Checks

- External inputs are validated, normalized, bounded, and encoded at the correct boundary.
- SQL, NoSQL, shell, template, path, LDAP, XML, GraphQL, and command construction uses safe APIs rather than string concatenation.
- File upload, archive extraction, path handling, MIME parsing, image/document processing, and downloads prevent traversal, overwrite, decompression, and content-type abuse.
- SSRF, open redirect, CORS, CSRF, clickjacking, and request smuggling risks are handled when network, redirect, or browser boundaries change.
- Serialization and deserialization do not permit unsafe classes, prototype pollution, confused types, or untrusted code execution.
- Sensitive data is minimized, encrypted or hashed where appropriate, redacted from logs, and not returned to clients unnecessarily.
- Database migrations and data transforms preserve integrity, uniqueness, foreign keys, and rollback expectations.

## Evidence

Follow the untrusted value from entry to sink. Identify validation, encoding, query parameterization, sanitizer, or escaping points. Verify the sink is actually reachable.

## Report

Name the source, sink, missing control, payload class, and impact. Avoid reporting hypothetical injection when safe APIs or hard boundaries are verified.
