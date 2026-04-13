# Local Insights Review & Experimentation Portal API Specification

## 1. Runtime Boundary

- This document describes the HTTP contract exposed by the Go/Gin backend in this repository.
- It is intended to reflect the behavior of the currently checked-in code, not an abstract future design.
- When there is a disagreement between this markdown file and the router/handlers, the code is the ground truth.

## 2. Source of Truth in the Codebase

- Route registration: `repo/backend/internal/router/router.go`
- HTTP handlers: `repo/backend/internal/handler/*.go`
- DTOs and validation: `repo/backend/internal/dto/*`
- Error types: `repo/backend/internal/errs/errors.go`
- Middleware (auth, CSRF, idempotency, IP rules, rate limit): `repo/backend/internal/middleware/*.go`
- Database schema and migrations: `repo/backend/migrations/*.sql`

Use this specification as a high-level contract guide and the code for precise field-level schemas.

## 3. Contract Conventions

- Base URL (default Docker setup): `https://localhost:8443`
- API prefix: `/api/v1`
- Default content type: `application/json` for both requests and responses unless explicitly noted.
- Character encoding: UTF-8.
- IDs:
	- Public-facing identifiers are UUID strings where applicable.
	- Internal numeric IDs may appear only on admin or analytics endpoints.
- Timestamps: RFC 3339 with milliseconds when present (for example, `2026-04-14T12:34:56.789Z`).

### 3.1 Global Headers

- `Authorization: Bearer <access_token>`
	- Required for all authenticated endpoints.
	- Optional for public endpoints; if present and valid, user context is still attached.
- `X-Request-Id: <string>`
	- Optional client-supplied correlation id.
	- If omitted, the backend generates one.
- `X-CSRF-Token: <token>`
	- Required on mutating requests (POST/PUT/PATCH/DELETE) when CSRF protection is enabled.
	- Must match the `csrf_token` cookie set by `GET /api/v1/csrf`.
- `X-Idempotency-Key: <opaque-string>`
	- Required for selected POST endpoints that are decorated with the idempotency middleware (see below).
	- Scoped per user and endpoint.

## 4. Error Envelope

All API errors follow a consistent JSON envelope:

```json
{
	"code": 400,
	"msg": "Invalid input"
}
```

- `code`: numeric HTTP status code repeated in the body.
- `msg`: human-readable, localized-safe message.

Common status codes:

- `400` — malformed JSON, missing fields, or generic bad request.
- `401` — missing or invalid authentication.
- `403` — authenticated but not allowed to perform the action.
- `404` — resource not found.
- `409` — conflict (for example, duplicate review, existing appeal).
- `412` — CAPTCHA required.
- `422` — structured validation errors.
- `429` — rate limit exceeded.
- `500` — unexpected internal error.

The `errs` package defines reusable error types (for example, `CAPTCHA_REQUIRED`, `IMAGE_TYPE_INVALID`, `IMAGE_TOO_LARGE`, `EXPERIMENT_NOT_ACTIVE`); handlers map these to the envelope above.

## 5. Authentication, Session, CSRF & Idempotency

### 5.1 Authentication & Tokens

- Authentication is performed via username/password credentials.
- Successful login issues:
	- short-lived access token (JWT) used in `Authorization: Bearer <token>`.
	- refresh token used to obtain new access tokens.
- Token validation and claim extraction are implemented in `internal/pkg/jwt` and `middleware/auth.go`.

### 5.2 CSRF Protection

- Implemented via the double-submit cookie pattern in `middleware/csrf.go`.
- `GET /api/v1/csrf`:
	- Generates a random token.
	- Sets it in a `csrf_token` cookie (Secure, SameSite=Strict).
	- Returns `{ "csrf_token": "..." }` in the body.
- Mutating requests must send the same token in the `X-CSRF-Token` header.
- If the token is missing or mismatched, the server returns `403` with `CSRF_INVALID`.

### 5.3 Idempotency

- Selected POST endpoints are wrapped with `middleware.Idempotency` and therefore require `X-Idempotency-Key`.
- The key is combined with the authenticated user id and HTTP method+path to derive a unique hash.
- On first use within a configured TTL, the response is stored in the `idempotency_keys` table.
- Subsequent calls with the same key, user, and endpoint replay the stored response.

### 5.4 Rate Limiting & IP Filtering

- `middleware.RateLimit` enforces a per-client request rate cap.
- `middleware.NewIPFilter` enforces allow/deny rules from the database:
	- Requests from denied addresses receive a `FORBIDDEN` error.
	- Admin endpoints manage these rules.

## 6. Shared Query Patterns

- Pagination: most list endpoints support `limit` (default 50, max 200) and `offset` (default 0).
- Sorting and filtering use endpoint-specific query parameters (for example, `status`, `role`, `created_after`).
- Unless documented otherwise, unspecified filters default to "all" records the user is permitted to see.

## 7. Endpoint Catalog

Paths below are relative to `/api/v1`.

### 7.1 Health & CSRF

#### GET /health

- Auth: none.
- Purpose: quick liveness and database connectivity check.
- Response `200`:
	- `{ "status": "healthy" }` when DB health check passes.
- Response `503`:
	- `{ "status": "unhealthy", "error": "..." }` when the DB is unreachable.

#### GET /csrf

- Auth: none.
- Purpose: issue CSRF token and cookie.
- Response `200`:
	- `{ "csrf_token": "<token>" }` and `csrf_token` cookie is set.

---

### 7.2 CAPTCHA

#### GET /captcha/generate

- Auth: none.
- Purpose: issue a new local CAPTCHA challenge (for example, after repeated failed logins).

#### POST /captcha/verify

- Auth: none.
- Body: verification payload defined in `dto/request.CaptchaVerifyRequest`.
- Purpose: verify user-supplied answer and unlock subsequent sensitive actions.

---

### 7.3 Authentication

#### POST /auth/register

- Auth: none.
- Body: `RegisterRequest` (username, email, password, optional profile fields).
- Response `201`:
	- Basic user summary (uuid, username, email, role, active flag, created timestamp).

#### POST /auth/login

- Auth: none.
- Body: `LoginRequest` (username or email plus password, optional CAPTCHA fields).
- Response `200`:
	- `access_token`, `refresh_token`, `expires_in`, and embedded user summary.

#### POST /auth/refresh

- Auth: none.
- Body: `RefreshRequest` (refresh token).
- Response `200`:
	- New access token plus embedded user summary.

#### POST /auth/logout

- Auth: bearer.
- Idempotent: yes (requires `X-Idempotency-Key`).
- Body: `RefreshRequest` (optional); if omitted the call still returns success.
- Response `200`:
	- `{ "msg": "Logged out" }`.

---

### 7.4 Public Browsing & Shared Views

#### GET /items

- Auth: optional.
- Purpose: list catalog items available for review and analytics.
- Supports standard pagination and basic filters (see item handler and service for concrete behavior).

#### GET /items/:id

- Auth: optional.
- Purpose: fetch a single item with core metadata needed for review & Q&A screens.

#### GET /items/:id/reviews

- Auth: optional.
- Purpose: list published reviews for an item, subject to moderation and visibility rules.

#### GET /items/:id/questions

- Auth: optional.
- Purpose: list questions associated with an item.

#### GET /questions/:id/answers

- Auth: optional.
- Purpose: list answers for a single question.

#### GET /images/:hash

- Auth: optional.
- Purpose: serve a stored image by content hash; respects quarantine and access rules.

#### GET /shared/:token

- Auth: bearer.
- Purpose: load metadata for a shared analytics view (read-only link for internal users).

#### GET /shared/:token/data

- Auth: bearer.
- Purpose: fetch the actual analytics payload for the shared view identified by token.

---

### 7.5 Authenticated User Surfaces

These endpoints require a valid JWT (`Authorization: Bearer ...`) and pass through `RequireAuth`.

#### GET /users/me

- Purpose: fetch the current user's profile (id, username, email, role, status, preferences).

#### PUT /users/me

- Purpose: update editable profile attributes (for example, display name, locale).

#### PUT /users/me/preferences

- Purpose: update per-user preferences such as timezone, language, and analytics options.

#### POST /items/:id/reviews

- Idempotent: yes.
- Purpose: create a new review for an item.
- Body: `CreateReviewRequest` (rating, optional text, references to uploaded images, etc.).
- Enforces:
	- one review per user per item.
	- rating and text validation rules.

#### PUT /reviews/:id

- Purpose: edit an existing review owned by the current user, subject to policy.

#### DELETE /reviews/:id

- Purpose: soft-delete or hide a review owned by the current user.

#### POST /images/upload

- Idempotent: yes.
- Purpose: upload a review image.
- Constraints (enforced by image service):
	- MIME sniffing for JPEG, PNG, WebP.
	- Maximum file size as configured.
	- Hash-based duplicate detection and quarantine for suspicious files.

#### POST /items/:id/questions

- Idempotent: yes.
- Purpose: create a new question tied to an item.

#### PUT /questions/:id

- Purpose: update an existing question owned by the caller.

#### DELETE /questions/:id

- Purpose: remove a question owned by the caller, where policy permits.

#### POST /questions/:id/answers

- Idempotent: yes.
- Purpose: add an answer to an existing question.

#### PUT /answers/:id

- Purpose: update an answer authored by the caller.

#### DELETE /answers/:id

- Purpose: delete an answer authored by the caller.

#### GET /favorites

- Purpose: list the caller's favorite items.

#### POST /favorites

- Idempotent: yes.
- Purpose: add an item to favorites.

#### DELETE /favorites/:item_id

- Purpose: remove an item from favorites.

#### GET /wishlists

- Purpose: list wishlists owned by the caller.

#### POST /wishlists

- Idempotent: yes.
- Purpose: create a wishlist.

#### PUT /wishlists/:id

- Purpose: rename or update a wishlist.

#### DELETE /wishlists/:id

- Purpose: delete a wishlist.

#### POST /wishlists/:id/items

- Idempotent: yes.
- Purpose: add an item to a wishlist.

#### DELETE /wishlists/:id/items/:item_id

- Purpose: remove an item from a wishlist.

#### POST /reports

- Idempotent: yes.
- Purpose: file a content report (for example, abusive review, problematic Q&A).

#### GET /reports/mine

- Purpose: list reports previously submitted by the caller.

#### POST /reports/:id/appeal

- Idempotent: yes.
- Purpose: create an appeal for a moderation outcome, where allowed.

#### PUT /reports/:id/appeal

- Purpose: resubmit or update an existing appeal owned by the caller.

#### POST /analytics/events

- Idempotent: yes.
- Purpose: ingest a batch of behavior events (impressions, clicks, dwell, etc.).

#### POST /analytics/sessions

- Idempotent: yes.
- Purpose: create a new analytics session record for client-side tracking.

#### PUT /analytics/sessions/:id/heartbeat

- Purpose: update last-seen timestamps and derived dwell metrics for a session.

#### GET /experiments/assignment/:exp_id

- Purpose: return the caller's variant assignment for an experiment (deterministic and sticky by design).

#### POST /experiments/:id/expose

- Idempotent: yes.
- Purpose: record exposure events tied to experiment assignments.

#### GET /notifications

- Purpose: list notifications for the current user.

#### GET /notifications/:id

- Purpose: fetch a single notification.

#### GET /notifications/unread-count

- Purpose: return unread notification count for badge UIs.

#### PUT /notifications/:id/read

- Purpose: mark a single notification as read.

#### PUT /notifications/read-all

- Purpose: mark all notifications as read.

#### POST /monitoring/frontend-errors

- Idempotent: yes.
- Purpose: capture structured frontend error reports for internal monitoring.

---

### 7.6 Moderator Endpoints

All moderator endpoints live under the `/moderation` prefix and require both authentication and role checks (`moderator` or `admin`).

#### GET /moderation/queue

- Purpose: unified moderation queue across reports, appeals, sensitive-word hits, quarantined content, and fraud-suspected items.

#### PUT /moderation/reports/:id

- Purpose: update report state and apply a moderation decision.

#### POST /moderation/reports/:id/notes

- Idempotent: yes.
- Purpose: add internal notes to a report.

#### GET /moderation/reports/:id/notes

- Purpose: list internal notes associated with a report.

#### GET /moderation/appeals

- Purpose: list appeals requiring review.

#### PUT /moderation/appeals/:id

- Purpose: accept, reject, or otherwise handle an appeal.

#### GET /moderation/quarantine

- Purpose: list quarantined uploads pending review.

#### PUT /moderation/quarantine/:id

- Purpose: approve or reject a quarantined upload.

#### GET /moderation/fraud

- Purpose: list reviews or sessions flagged as suspected fraud.

#### PUT /moderation/fraud/:review_id

- Purpose: confirm or clear fraud status for a review.

#### GET /moderation/word-rules

- Purpose: list configured sensitive-word rules.

#### POST /moderation/word-rules

- Idempotent: yes.
- Purpose: create a new sensitive-word rule.

#### PUT /moderation/word-rules/:id

- Purpose: update an existing rule.

#### DELETE /moderation/word-rules/:id

- Purpose: delete a rule.

---

### 7.7 Product Analyst Endpoints

These endpoints require the `product_analyst` or `admin` role.

#### GET /analytics/dashboard

- Purpose: return aggregated analytics metrics for dashboards (filters passed via query string).

#### GET /analytics/keywords
#### GET /analytics/topics
#### GET /analytics/cooccurrences
#### GET /analytics/sentiment

- Purpose: expose specialized analytics slices for keyword, topic, co-occurrence, and sentiment visualizations.

#### GET /analytics/aggregate-sessions

- Purpose: list aggregated session summaries for exploration.

#### GET /analytics/sessions/:id

- Purpose: fetch a single analytics session record.

#### GET /analytics/sessions/:id/timeline

- Purpose: fetch ordered events for a session.

#### GET /analytics/saved-views
#### POST /analytics/saved-views
#### PUT /analytics/saved-views/:id
#### DELETE /analytics/saved-views/:id

- Purpose: CRUD operations for saved analytics views (filter sets + layout configuration).

#### POST /analytics/saved-views/:id/share

- Idempotent: yes.
- Purpose: create a time-limited internal share link for a saved view.

#### DELETE /analytics/saved-views/:id/share

- Purpose: revoke a previously issued share link.

#### POST /analytics/saved-views/clone

- Idempotent: yes.
- Purpose: clone a shared or existing view into the caller's workspace.

#### GET /experiments
#### POST /experiments
#### GET /experiments/:id
#### PUT /experiments/:id

- Purpose: list, create, fetch, and update A/B experiments.

#### PUT /experiments/:id/traffic

- Purpose: adjust traffic allocation between variants.

#### POST /experiments/:id/start
#### POST /experiments/:id/pause
#### POST /experiments/:id/complete
#### POST /experiments/:id/rollback

- Idempotent: yes on all mutating operations.
- Purpose: control experiment lifecycle.

#### GET /experiments/:id/results

- Purpose: fetch metrics and confidence indicators for an experiment.

#### GET /scoring/weights
#### PUT /scoring/weights
#### GET /scoring/weights/history

- Purpose: view and manage preference scoring weights and their version history.

---

### 7.8 Admin Endpoints

Admin endpoints live under `/admin` and require the `admin` role.

#### GET /admin/users

- Purpose: list users with basic account metadata and filters.

#### PUT /admin/users/:id/role

- Purpose: update a user's primary role (for example, upgrade to moderator or analyst).

#### PUT /admin/users/:id/status

- Purpose: activate/deactivate a user.

#### GET /admin/audit-logs

- Purpose: search and page through audit records.

#### GET /admin/ip-rules
#### POST /admin/ip-rules
#### DELETE /admin/ip-rules/:id

- Purpose: manage IP allowlist/denylist entries.

#### POST /admin/backup/trigger

- Idempotent: yes.
- Purpose: trigger an on-demand backup.

#### GET /admin/recovery-drills
#### POST /admin/recovery-drills/trigger

- Purpose: list historical recovery drills and trigger new ones.

#### POST /admin/analytics/rebuild

- Idempotent: yes.
- Purpose: trigger a rebuild of analytics summary tables.

#### GET /admin/monitoring/performance
#### GET /admin/monitoring/errors
#### GET /admin/monitoring/health

- Purpose: expose internal performance, error, and health metrics for operators.

## 8. Representative Models

For full field lists, see DTOs and models in the codebase. The following are indicative examples of JSON shapes.

### 8.1 Error

```json
{
	"code": 422,
	"msg": "Rating must be between 1 and 5"
}
```

### 8.2 User (summary)

```json
{
	"id": "e5f1c8b2-0d2b-4cce-9e7b-3b6f9f1a9c10",
	"username": "analyst1",
	"email": "analyst1@example.local",
	"role": "product_analyst",
	"is_active": true
}
```

### 8.3 Item (summary)

```json
{
	"id": "fcde9f10-21a4-4a15-9e86-1f5f8c6a1234",
	"title": "Internal Feature X",
	"category": "experiment-target",
	"average_rating": 4.3,
	"review_count": 27
}
```

### 8.4 Review (summary)

```json
{
	"id": "a1b2c3d4-1111-2222-3333-444455556666",
	"item_id": "fcde9f10-21a4-4a15-9e86-1f5f8c6a1234",
	"rating": 5,
	"text": "Great internal experience.",
	"created_at": "2026-04-14T10:01:02.345Z",
	"updated_at": "2026-04-14T10:05:00.000Z"
}
```

### 8.5 Analytics Event (simplified)

```json
{
	"session_id": "s123",
	"type": "item_view",
	"item_id": "fcde9f10-21a4-4a15-9e86-1f5f8c6a1234",
	"occurred_at": "2026-04-14T10:00:00.000Z"
}
```

### 8.6 Experiment (summary)

```json
{
	"id": "exp-landing-layout-1",
	"name": "Landing layout variant test",
	"status": "running",
	"primary_metric": "click_through_rate"
}
```

## 9. Auditing & Masking

- Mutating operations on security-sensitive and operational surfaces are logged to an audit table.
- Examples include:
	- role changes
	- moderation decisions
	- IP rule updates
	- scoring weight changes
	- experiment lifecycle actions
- Audit entries record actor, action, resource type/id, and masked details.
- Personally identifiable or sensitive values are never stored in cleartext in audit details.

## 10. Notes on Completeness

- This document focuses on route-level contracts and high-level payload structures.
- Validation rules, enum values, and full field lists live with the Go DTOs, services, and database migrations.
- When adding or changing endpoints:
	- update `router/router.go` and handlers first,
	- update DTOs and migrations as needed,
	- then refresh this specification to match the implemented behavior.

