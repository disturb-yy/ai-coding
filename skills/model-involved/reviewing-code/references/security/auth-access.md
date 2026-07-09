# Auth And Access Control

Use this checklist for permissions, tenancy, identity, and protected operations.

## Checks

- Authentication is required on new or changed protected paths.
- Authorization checks happen on the server or trusted boundary, not only in UI or client code.
- Object-level access control prevents reading or mutating another user's, tenant's, organization’s, or project’s data.
- Admin, service-account, internal, feature-flag, debug, and maintenance paths are not exposed to ordinary users.
- Security decisions do not trust user-controlled identifiers, headers, cookies, query params, role names, or client-side state without verification.
- New background jobs, webhooks, queues, cron tasks, and callbacks preserve authorization and tenant context.
- Error messages and logs do not disclose protected resource existence across boundaries.

## Evidence

Trace request or job entry to auth middleware, policy checks, data filters, and storage queries. Use tests or existing policy helpers as comparison points.

## Report

State the actor, protected asset, missing or bypassed check, and exploit path. Mark as `Critical` or `High` when unauthorized data access or mutation is plausible.
