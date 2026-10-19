# Memory routing

Read this when preferences, configuration, usage habits, or connection metadata could change the approach.

## Retrieve

Use the index named by current host instructions. Read only records matching the task's topic and scope, then check their source and verification date. Verify facts likely to drift against the current workspace or service before relying on them. Current user instructions and observed facts take precedence over older records.

An unavailable index or record is missing evidence, not a reason to invent a setting or stop unrelated work. Retrieve the fact from current sources where possible; disclose a memory gap only where it affects the result. Do not create a memory store just to finish a lookup.

## Persist

Write, update, or retire persistent memory only with user authorization and within host storage rules. If available, use the memory-authoring skill that matches the request. Otherwise keep the authorized record focused on a stable fact, its scope, source, and verification date; include a review date for facts likely to drift. Check for an equivalent record before adding one.

Update the configured index in the same change when permitted. If the host instead requires a designated update-note location, use it and report that the index has not been changed. A task checkpoint holds continuation state; it is not automatically a durable preference or project fact.

## Connection records

Store service identifiers, account aliases, and configuration locations, never credentials, tokens, or full connection strings. Keep a credential-file locator in the scoped record rather than the root index. A locator grants neither file access nor login authority.

For an authorized login, verify the intended service and permitted file access, then use a trusted tool that consumes the credential file directly and returns status only. Keep values out of prompts, command arguments, logs, and memory. If secure use or permitted access is unavailable, report that boundary and continue independent work.
