# Test Coverage Audit

## Project Type Detection
- README top does not declare one exact token from: backend, fullstack, web, android, ios, desktop.
- README evidence: repo/README.md:3 contains "full-stack application".
- Inferred type (strict mode): fullstack.

## Backend Endpoint Inventory
- Source: repo/backend/internal/router/router.go
- Total endpoints: 109

### Backend Endpoint Inventory
- Endpoint inventory extracted statically from router groups and route declarations.

### API Test Mapping Table
- Mapping completed with template-aware path matching (e.g., :id -> concrete segment).
- Result: all endpoints mapped to direct HTTP call evidence in repo/API_tests/run_api_tests.sh.

### Coverage Summary
- total endpoints: 109
- endpoints with HTTP tests: 109
- endpoints with TRUE no-mock tests: 109
- HTTP coverage %: 100.0
- True API coverage %: 100.0

### Unit Test Summary
#### Backend Unit Tests
- test files detected: 16
- modules covered include middleware/auth plus selected service/repository/helper contracts.
- important backend modules still not deeply tested directly: broad portions of service/repository internals and runtime job behavior.

#### Frontend Unit Tests (STRICT REQUIREMENT)
- frontend test files detected: 19
- frameworks/tools detected: Vitest, Vue Test Utils
- component/module-level tests present for catalog, item detail, experiments, auth flows, moderation queue, routing, and stores.
- Mandatory Verdict: Frontend unit tests: PRESENT

### API Test Classification
1. True No-Mock HTTP
- repo/API_tests/run_api_tests.sh
2. HTTP with Mocking
- none found in API_tests
3. Non-HTTP (unit/integration without HTTP)
- repo/unit_tests/*.go
- repo/frontend/src/__tests__/*.test.ts

### Mock Detection Rules Output
- API layer: no jest.mock / vi.mock / sinon.stub / DI override indicators in API_tests.
- Frontend tests contain controlled vi.mock usage in several unit tests.

### API Observability Check
- mixed: endpoint/status coverage is explicit; response-contract depth varies by endpoint family.

### Test Quality & Sufficiency
- strengths: complete static route-hit coverage and broad auth/permission/error-path checks.
- gaps: depth of assertions and business-layer unit depth remain uneven.

### Tests Check
- repo/run_tests.sh is Docker-based: OK

### End-to-End Expectations
- fullstack expects FE<->BE E2E; evidence is API-strong with frontend unit presence, while browser-level E2E remains partial.

### Test Coverage Score (0-100)
90

### Score Rationale
- + direct HTTP mapping covers 109/109 endpoints (100.0%).
- + no API-layer mocking detected in API_tests.
- - assertion depth remains mixed.
- - backend business-layer direct unit depth remains partial.

### Key Gaps
1. Increase response-contract and business-invariant assertions on critical admin/moderation/analytics routes.
2. Expand direct service and repository unit tests for core business logic and failure handling.
3. Add stronger browser-level FE<->BE end-to-end workflow coverage.

### Confidence & Assumptions
- confidence: high for static endpoint inventory and call-site mapping.
- confidence: medium for behavioral sufficiency under static-only constraints.
- assumption: repo/backend/internal/router/router.go is canonical route map.

---

# README Audit

## README Location
- found: repo/README.md

### Hard Gate Failures
1. startup literal mismatch
- required: docker-compose up
- found: docker compose up (repo/README.md:14, repo/README.md:261)
- result: FAIL

2. manual DB setup instruction present
- repo/README.md:38 and repo/README.md:41 instruct manual SQL role promotion
- result: FAIL

3. demo credentials for all auth roles missing
- auth exists: repo/README.md:52 and repo/README.md:53
- explicit username/email + password for all roles is not provided
- result: FAIL

### High Priority Issues
1. Missing all-role demo credentials.
2. Manual DB mutation violates Docker-contained setup expectations.
3. Required startup literal docker-compose up is absent.

### Medium Priority Issues
1. Required top token (backend/fullstack/web/android/ios/desktop) is not explicitly declared at the top.
2. Onboarding and operational guidance are mixed, reducing strict runbook clarity.

### Low Priority Issues
1. README can be split into quick-start vs reference for auditability.
2. Credential and auth sections can be separated for clearer operator flow.

### README Verdict (PASS / PARTIAL PASS / FAIL)
FAIL

---

# Final Verdicts
1. Test Coverage Audit: PARTIAL PASS
2. README Audit: FAIL
