# Local Insights Review & Experimentation Portal

A full-stack application for managing user-generated feedback, moderating content, analyzing behavior, and running A/B experiments — designed to run entirely within an offline network with zero external dependencies.

## Tech Stack

- **Backend**: Go 1.25 + Gin framework + MySQL 8.0 (via sqlx)
- **Frontend**: Vue 3 + TypeScript + Vite + Pinia + Vue Router + ECharts
- **Infrastructure**: Docker Compose (MySQL + Go backend + Nginx-served frontend)

## Quick Start

```bash
docker-compose up
```

This single command builds and starts all three services:

| Service    | URL                           | Description                              |
|------------|-------------------------------|------------------------------------------|
| Frontend   | https://localhost:3000        | Vue.js SPA served by Nginx (HTTPS)       |
| Backend API| https://localhost:8443/api/v1 | Go/Gin REST API (HTTPS, self-signed TLS) |
| MySQL      | localhost:3306                | MySQL 8.0 database                       |

All traffic is encrypted with HTTPS using auto-generated self-signed certificates. Browsers will show a certificate warning — this is expected for local self-signed certs.

The backend automatically runs all database migrations on startup. No manual setup required.

## Default Credentials

### Application demo accounts

The very first `docker-compose up` seeds one account for every role so you can
exercise the full role matrix without any manual SQL. All four accounts are
active and ready to use immediately:

| Role              | Username         | Email                    | Password          |
|-------------------|------------------|--------------------------|-------------------|
| admin             | `admin`          | `admin@local.test`       | `AdminPass1`      |
| moderator         | `moderator`      | `moderator@local.test`   | `ModeratorPass1`  |
| product_analyst   | `product_analyst`| `analyst@local.test`     | `AnalystPass1`    |
| regular_user      | `demo_user`      | `demo@local.test`        | `DemoPass1`       |

Rotate these before any shared or production deployment. New end-user accounts
are self-registered via `POST /api/v1/auth/register` and receive the
`regular_user` role by default; use the `admin` account above to change roles
through `PUT /api/v1/admin/users/:id/role`.

### Database credentials

| Field             | Value         |
|-------------------|---------------|
| MySQL root pass   | rootpassword  |
| MySQL user        | appuser       |
| MySQL password    | apppassword   |
| MySQL database    | local_insights|

## API Endpoints

All endpoints are prefixed with `/api/v1`.

### Public (No Authentication)

| Method | Path                          | Description                    |
|--------|-------------------------------|--------------------------------|
| POST   | /auth/register                | Register a new user            |
| POST   | /auth/login                   | Login, returns JWT tokens      |
| POST   | /auth/refresh                 | Refresh access token           |
| GET    | /captcha/generate             | Generate CAPTCHA challenge     |
| POST   | /captcha/verify               | Verify CAPTCHA answer          |
| GET    | /items                        | List published items           |
| GET    | /items/:id                    | Get item detail                |
| GET    | /items/:id/reviews            | List reviews for an item       |
| GET    | /items/:id/questions          | List questions for an item     |
| GET    | /questions/:id/answers        | List answers for a question    |
| GET    | /images/:hash                 | Serve image by content hash    |
`| GET    | /csrf                         | Get CSRF token (set in cookie) |
| GET    | /health                       | Health check                   |

### Authenticated (Any Role)

Include `Authorization: Bearer <token>` header.

| Method | Path                          | Description                    |
|--------|-------------------------------|--------------------------------|
| POST   | /auth/logout                  | Logout (revoke refresh token)  |
| GET    | /users/me                     | Get current user profile       |
| PUT    | /users/me                     | Update profile                 |
| PUT    | /users/me/preferences         | Update locale/timezone prefs   |
| POST   | /items/:id/reviews            | Create review (idempotent)     |
| PUT    | /reviews/:id                  | Update own review              |
| DELETE | /reviews/:id                  | Delete own review              |
| POST   | /images/upload                | Upload image (JPEG/PNG/WebP)   |
| POST   | /items/:id/questions          | Ask a question                 |
| PUT    | /questions/:id                | Edit own question              |
| DELETE | /questions/:id                | Delete own question            |
| POST   | /questions/:id/answers        | Post an answer                 |
| PUT    | /answers/:id                  | Edit own answer                |
| DELETE | /answers/:id                  | Delete own answer              |
| GET    | /favorites                    | List favorites                 |
| POST   | /favorites                    | Add to favorites               |
| DELETE | /favorites/:item_id           | Remove from favorites          |
| GET    | /wishlists                    | List wishlists                 |
| POST   | /wishlists                    | Create wishlist                |
| PUT    | /wishlists/:id                | Rename wishlist                |
| DELETE | /wishlists/:id                | Delete wishlist                |
| POST   | /wishlists/:id/items          | Add item to wishlist           |
| DELETE | /wishlists/:id/items/:item_id | Remove item from wishlist      |
| POST   | /reports                      | File a report (idempotent)     |
| GET    | /reports/mine                 | List own reports               |
| POST   | /reports/:id/appeal           | Appeal a moderation decision   |
| POST   | /analytics/events             | Ingest behavior events (batch) |
| POST   | /analytics/sessions           | Create analytics session       |
| PUT    | /analytics/sessions/:id/heartbeat | Session heartbeat          |
| GET    | /experiments/assignment/:id   | Get experiment assignment      |
| GET    | /notifications                | List notifications             |
| GET    | /notifications/:id            | Get notification detail        |
| GET    | /notifications/unread-count   | Get unread count               |
| PUT    | /notifications/:id/read       | Mark notification read         |
| PUT    | /notifications/read-all       | Mark all read                  |

### Moderator (moderator + admin)

| Method | Path                              | Description                    |
|--------|-----------------------------------|--------------------------------|
| GET    | /moderation/queue                 | Moderation queue               |
| PUT    | /moderation/reports/:id           | Update report status           |
| POST   | /moderation/reports/:id/notes     | Add moderation note            |
| GET    | /moderation/reports/:id/notes     | List notes for report          |
| PUT    | /moderation/appeals/:id           | Handle appeal                  |
| GET    | /moderation/quarantine            | List quarantined images        |
| PUT    | /moderation/quarantine/:id        | Approve/reject quarantined     |
| GET    | /moderation/fraud                 | List suspected fraud reviews   |
| PUT    | /moderation/fraud/:review_id      | Confirm/clear fraud            |
| GET    | /moderation/word-rules            | List sensitive word rules      |
| POST   | /moderation/word-rules            | Create word rule               |
| PUT    | /moderation/word-rules/:id        | Update word rule               |
| DELETE | /moderation/word-rules/:id        | Delete word rule               |

### Product Analyst (product_analyst + admin)

| Method | Path                              | Description                    |
|--------|-----------------------------------|--------------------------------|
| GET    | /analytics/dashboard              | Analytics dashboard data       |
| GET    | /analytics/sessions/:id           | Session drill-down             |
| GET    | /analytics/sessions/:id/timeline  | Session event timeline         |
| GET    | /analytics/saved-views            | List saved views               |
| POST   | /analytics/saved-views            | Save a view                    |
| PUT    | /analytics/saved-views/:id        | Update saved view              |
| DELETE | /analytics/saved-views/:id        | Delete saved view              |
| POST   | /analytics/saved-views/:id/share  | Create share link (7-day exp.) |
| DELETE | /analytics/saved-views/:id/share  | Revoke share link              |
| POST   | /analytics/saved-views/clone      | Clone shared view              |
| GET    | /experiments                      | List experiments               |
| POST   | /experiments                      | Create experiment              |
| PUT    | /experiments/:id                  | Update experiment              |
| POST   | /experiments/:id/start            | Start experiment               |
| POST   | /experiments/:id/pause            | Pause experiment               |
| POST   | /experiments/:id/complete         | Complete experiment            |
| POST   | /experiments/:id/rollback         | Rollback experiment            |
| GET    | /experiments/:id/results          | Get results + confidence       |
| POST   | /experiments/:id/expose           | Record exposure (idempotent)   |
| GET    | /scoring/weights                  | Get scoring weights            |
| PUT    | /scoring/weights                  | Update scoring weights         |
| GET    | /scoring/weights/history          | Scoring weight version history |

### Admin Only

| Method | Path                              | Description                    |
|--------|-----------------------------------|--------------------------------|
| GET    | /admin/users                      | List all users                 |
| PUT    | /admin/users/:id/role             | Change user role               |
| PUT    | /admin/users/:id/status           | Activate/deactivate user       |
| GET    | /admin/audit-logs                 | View audit logs                |
| GET    | /admin/ip-rules                   | List IP allowlist/denylist     |
| POST   | /admin/ip-rules                   | Add IP rule                    |
| DELETE | /admin/ip-rules/:id               | Remove IP rule                 |
| POST   | /admin/backup/trigger             | Trigger manual backup          |
| GET    | /admin/recovery-drills            | List recovery drill results    |
| POST   | /admin/recovery-drills/trigger    | Trigger recovery drill         |
| POST   | /admin/analytics/rebuild          | Rebuild analytics aggregates   |
| GET    | /admin/monitoring/performance     | API latency metrics            |
| GET    | /admin/monitoring/errors          | Error metrics                  |
| GET    | /admin/monitoring/health          | System health status           |

## Error Response Format

All API errors return consistent JSON:

```json
{
  "code": 400,
  "msg": "Invalid email format"
}
```

No stack traces, internal details, or framework-specific information is ever exposed.

## Security Features

- **Authentication**: JWT access tokens (15min) + refresh tokens (7 days)
- **Password Hashing**: Argon2id with random salt
- **CAPTCHA**: Local arithmetic CAPTCHA after 5 failed logins in 15 minutes
- **CSRF Protection**: Double-submit cookie pattern enabled by default (`CSRF_ENABLED=true` in docker-compose.yml). Frontend calls `GET /api/v1/csrf` before the first mutating request; the response is `{"csrf_token":"<hex>"}` and a `csrf_token` cookie is set. Mutating requests must include the token in the `X-CSRF-Token` header matching the cookie.
- **Rate Limiting**: 60 requests/minute per user (token bucket)
- **IP Filtering**: Allowlist/denylist with deny-first precedence
- **Idempotency**: All authenticated POST endpoints require an `X-Idempotency-Key` header (10 min TTL). Public auth routes (register/login/refresh) are exempt. The key is scoped per user and endpoint to prevent cross-endpoint replay.
- **Input Validation**: All inputs validated on both frontend and backend
- **XSS Protection**: Vue auto-escaping + DOMPurify for any HTML rendering
- **Security Headers**: X-Content-Type-Options, X-Frame-Options, HSTS, CSP, Referrer-Policy
- **Audit Logging**: All security-sensitive actions logged with masked sensitive fields
- **Image Validation**: Magic-byte format sniffing, SHA-256 dedup, suspicious content quarantine

## Secure Deployment Defaults

The Docker Compose configuration ships with security enabled by default:
- **HTTPS**: Self-signed TLS certificates auto-generated at build time
- **CSRF**: Enabled (`CSRF_ENABLED=true`) — frontend bootstraps token automatically
- **Rate Limiting**: 60 requests/minute per IP
- **JWT Secret**: Must be changed from the default in production

To run API tests with CSRF enabled, the test harness automatically obtains a CSRF token and includes it in all mutating requests.

## Background Jobs

The backend runs scheduled background jobs:

| Job               | Schedule    | Purpose                                           |
|-------------------|-------------|---------------------------------------------------|
| Rating Refresh    | Every 30s   | Recompute item rating aggregates                  |
| NLP Processing    | Every 5min  | Sentiment analysis, keyword extraction             |
| Analytics ETL     | Every 5min  | Aggregate behavior events into daily summaries     |
| Fraud Scan        | Every 10min | Rate-based + sequence-based fraud detection        |
| Idempotency Cleanup| Every 2min | Purge expired idempotency keys                    |
| Monitoring        | Every 1min  | Record performance metrics                         |
| Nightly Backup    | 2:00 AM     | MySQL dump with 30-day retention                   |
| Recovery Drill    | Sunday 3AM  | Restore backup to temp DB, verify integrity        |

## Running Tests

```bash
./run_tests.sh
```

This script:
1. Tears down any existing containers
2. Builds and starts all services via Docker Compose
3. Waits for MySQL and backend to be healthy
4. Runs **unit tests** inside a Go container (tests pure logic: hashing, JWT, CAPTCHA, audit masking, image processing, locale formatting, config, monitoring)
5. Runs **API tests** inside a curl container on the Docker network (tests all API endpoints: registration, login, token refresh, profile, favorites, notifications, CAPTCHA, error format validation, security)
6. Prints PASS/FAIL summary
7. Tears down all containers

### Test Structure

```
unit_tests/          # Go unit tests (no DB required)
  hash_test.go       # Argon2id password hashing
  jwt_test.go        # JWT generation/validation
  captcha_test.go    # CAPTCHA generation/verification
  audit_test.go      # Audit log field masking
  imagepro_test.go   # Image processing and validation
  locale_test.go     # Timezone/locale formatting
  config_test.go     # Configuration loading
  monitor_test.go    # Performance metric collection

API_tests/           # Functional API tests (requires running services)
  run_api_tests.sh   # Curl-based API test suite

run_tests.sh         # Master test orchestrator
```

## Verification

After running `docker-compose up`, verify the system is working:

1. **Health check** (use `-k` for self-signed cert):
   ```bash
   curl -k https://localhost:8443/api/v1/health
   # {"status":"healthy"}
   ```

2. **Register a user**:
   ```bash
   curl -k -X POST https://localhost:8443/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"username":"demo","email":"demo@local.test","password":"DemoPass1"}'
   ```

3. **Login**:
   ```bash
   curl -k -X POST https://localhost:8443/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"demo","password":"DemoPass1"}'
   # Returns {access_token, refresh_token, user}
   ```

4. **Access protected route**:
   ```bash
   curl -k https://localhost:8443/api/v1/users/me \
     -H "Authorization: Bearer <access_token>"
   ```

5. **Frontend**:
   Open https://localhost:3000 in a browser (accept the self-signed certificate warning). You should see the login page.

6. **Generate CAPTCHA**:
   ```bash
   curl -k https://localhost:8443/api/v1/captcha/generate
   # Returns {captcha_id, captcha_image} with base64 PNG
   ```

## Project Structure

```
repo/
├── docker-compose.yml          # Orchestrates all services
├── README.md                   # This file
├── run_tests.sh                # Test orchestrator
├── unit_tests/                 # Go unit tests
├── API_tests/                  # API functional tests
├── backend/
│   ├── Dockerfile
│   ├── cmd/server/main.go      # Entry point with DI wiring
│   ├── internal/
│   │   ├── config/             # Environment-based configuration
│   │   ├── middleware/         # 10 middleware layers
│   │   ├── handler/            # 15 HTTP handler files
│   │   ├── service/            # Business logic
│   │   ├── repository/         # Data access (22 interfaces, 18 impls)
│   │   ├── model/              # Domain models
│   │   ├── dto/                # Request/response DTOs
│   │   ├── router/             # Route registration
│   │   ├── job/                # Background job scheduler
│   │   ├── pkg/                # Shared packages
│   │   └── errs/               # Error types
│   └── migrations/             # 52 SQL migration files (up + down)
└── frontend/
    ├── Dockerfile
    ├── nginx.conf              # Nginx config with API proxy
    └── src/
        ├── api/                # Axios client + endpoint modules
        ├── components/         # Vue components
        ├── composables/        # Vue composables
        ├── layouts/            # Auth, Default, Blank layouts
        ├── pages/              # Route-level page components
        ├── router/             # Vue Router with guards
        ├── stores/             # Pinia stores
        ├── types/              # TypeScript types and enums
        └── utils/              # Utility functions
```
