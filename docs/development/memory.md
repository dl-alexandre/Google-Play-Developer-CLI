# Run Memory

## 2026-02-19

- Remote ASC verification is pending because the GitHub commits page could not be opened (cache miss).
- Highest-ROI gaps remain: golden output snapshots, bulk tester lifecycle parity decision (implement CSV workflows or mark non-goal), Play OpenAPI/spec snapshot policy, strict auth/config conflict detection and precedence docs, destructive-op safety flag normalization, wider output-format golden coverage, and cookbook troubleshooting completeness.
- Evidence updates: GPD now has mock API testing infrastructure and expanded unit tests (`internal/apitest/mock_client.go`, `internal/api/client_test.go`), plus config normalization/validation improvements in `internal/config/config.go`.
- Issue scope candidate: add golden output snapshots using the new mock API infra with stable assertions for JSON envelope and list/table/markdown outputs.
