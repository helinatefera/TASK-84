# Issue Fix Verification (Re-check)

Date: 2026-04-13
Source report reviewed: `.tmp/audit_report-1.md.md`
Method: Static code inspection only (no runtime/test execution)

## Summary
- Fixed: 6
- Partially fixed: 0
- Not fixed: 0

## Per-Issue Verification

### FND-001 - Event deduplication used one batch-wide timestamp
- Previous finding: dedup key could collapse distinct events within a batch.
- Current status: **Fixed**
- Evidence:
  - `backend/internal/handler/analytics_handler.go`: dedup hash now uses per-event `clientTS` (with parse fallback), not one batch-wide timestamp.
  - `computeDedupHash(...)` includes a time bucket derived from event timestamp.
- Notes:
  - This resolves the specific prior defect from the original report.

### FND-002 - Anti-fraud flagged reviews but not accounts
- Previous finding: account-level flagging requirement mismatch.
- Current status: **Fixed**
- Evidence:
  - `backend/internal/job/jobs.go`: fraud scan now updates `users.fraud_status` for both rate-based and sequence-based detections.
  - `backend/internal/model/user.go`: `FraudStatus` field added to user model.
  - `backend/migrations/049_add_users_fraud_status.up.sql`: schema migration adds indexed `users.fraud_status`.

### FND-003 - Security-sensitive action audit/masking only partial
- Previous finding: role/status/IP-rule mutation auditing not fully wired through masked audit logger.
- Current status: **Fixed**
- Evidence:
  - `backend/internal/handler/admin_handler.go`: `UpdateUserRole`, `UpdateUserStatus`, `CreateIPRule`, and `DeleteIPRule` all call `auditLogger.Log(...)`.
  - `backend/internal/pkg/audit/logger.go`: logger applies `MaskDetails(...)` before persistence.

### FND-004 - HTTPS could fall back to HTTP
- Previous finding: strict HTTPS intent was weakened by automatic plaintext fallback.
- Current status: **Fixed**
- Evidence:
  - `backend/internal/config/config.go`: `RequireTLS` defaults to `true`.
  - `backend/cmd/server/main.go`: startup now `log.Fatalf(...)` when `RequireTLS=true` and cert/key are missing; HTTP fallback only when TLS is explicitly not required.

### FND-005 - API tests lacked missing idempotency-key rejection checks
- Previous finding: no negative contract tests for required `X-Idempotency-Key`.
- Current status: **Fixed**
- Evidence:
  - `API_tests/run_api_tests.sh`: explicit checks added for 400 responses when POST requests omit `X-Idempotency-Key` (e.g., `/reports`, `/items/:id/reviews`, `/experiments/:id/expose`).

### FND-006 - Frontend analytics tests did not cover shared-token integration flow
- Previous finding: tests were helper-level and missed mounted shared/read-only integration behavior.
- Current status: **Fixed**
- Evidence:
  - `frontend/src/__tests__/AnalyticsDashboard.test.ts`: mounted component tests cover shared-token mode, `/shared/:token` + `/shared/:token/data` calls, readonly UI restrictions, and normal-mode endpoint behavior.

## Final Conclusion
All previously reported findings in `.tmp/audit_report-1.md.md` are fixed based on current static evidence.

## Boundary Reminder
This verification is static-only. Runtime behavior, deployment correctness, and end-to-end production characteristics still require execution-based validation.
