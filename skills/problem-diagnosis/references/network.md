# Network Diagnosis Reference

Use this reference for HTTP, DNS, TLS, proxy, timeout, retry, queue, webhook,
and distributed-call failures. Diagnose only; do not change infrastructure or
configuration.

Collect request path, method, status code, headers when safe, correlation IDs,
timeout values, retry policy, upstream/downstream names, DNS/TLS errors, and
timestamped logs from both sides of the boundary when available.

Prefer read-only probes:

```bash
curl -v https://example.com/health
curl --max-time 10 -v https://example.com/path
nslookup example.com
```

Record whether the issue appears client-side, server-side, network-path,
configuration, authentication, rate-limit, or dependency related.

