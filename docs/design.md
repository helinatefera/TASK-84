````md
# Design Document: Local Insights Review & Experimentation Portal

## 1. Overview

The Local Insights Review & Experimentation Portal is an offline-first internal platform for collecting user-generated feedback, moderating community content, analyzing product behavior, and running controlled A/B experiments without internet dependency.

The system is designed for a product team operating inside a restricted local network. Regular Users browse internal items, submit reviews, ask and answer questions, and manage favorites and wishlists. Moderators review reports, appeals, quarantined files, and sensitive-word violations through a unified moderation queue. Product Analysts explore feedback and behavioral analytics, configure scoring weights, save analytics views, and run offline canary experiments with rollback controls. Administrators manage system-wide roles, permissions, security controls, and operational settings.

The frontend is built with Vue.js and communicates with a Gin-based REST API. MySQL serves as the system of record for transactional and reporting data. The platform emphasizes reliability, safety, and auditability under a fully offline deployment model.

---

## 2. Goals

### 2.1 Primary Goals

- Support internal user registration and role-based access
- Allow users to review internal catalog items with ratings, text, and images
- Support Q&A, favorites, and wishlists with device sync through the local server
- Provide a unified moderation workflow for reports, appeals, sensitive-word checks, and quarantined files
- Deliver offline analytics dashboards with drill-down capability from aggregates to session-level detail
- Support offline A/B experiments with canary rollout, exposure tracking, and rollback decisions
- Enforce strong local-network security and auditable moderation, permission, and configuration changes
- Keep analytics responsive using scheduled sync and summary refresh jobs
- Operate fully without internet access or external services

### 2.2 Non-Goals

- Public internet access
- External machine learning or cloud moderation services
- External CAPTCHA providers
- Online payment or e-commerce features
- Third-party authentication providers
- Real-time internet-based analytics platforms

---

## 3. User Roles

### 3.1 Regular User

Regular Users can:
- Register and sign in with username and password
- Browse internal items
- Submit ratings, reviews, and images
- Ask and answer item-specific Q&A
- Manage favorites and wishlists
- Report content
- Appeal moderation outcomes where allowed

### 3.2 Moderator

Moderators can:
- Review flagged or reported content
- Apply sensitive-word moderation rules
- Review quarantined files
- Process appeals
- Record moderation decisions with required notes

### 3.3 Product Analyst

Product Analysts can:
- View analytics dashboards
- Filter feedback and behavior data
- Save and share read-only analytics views
- Configure preference scoring weights
- Create and monitor A/B experiments
- Review confidence indicators and decide whether to keep or roll back variants

### 3.4 Administrator

Administrators can:
- Manage roles and permissions
- Manage item catalog lifecycle
- Manage IP allowlists and denylists
- Oversee system operations and backup status
- Access administrative audit and operational dashboards
- Configure security and deployment settings

---

## 4. System Context

The system operates entirely inside a local network and consists of:

- A Vue.js frontend
- A Gin REST API backend
- A MySQL database
- Local file storage for uploaded images and quarantined files
- Background scheduled jobs for aggregation, backups, sync, and recovery drills
- Internal dashboards for observability and moderation

No runtime dependency may rely on external internet connectivity. All application assets, scripts, charting libraries, CAPTCHA resources, and static files are bundled locally.

---

## 5. High-Level Architecture

## 5.1 Component View

```text
+------------------------------------------------------+
| Vue.js Frontend                                      |
| - Auth UI                                            |
| - Item Catalog UI                                    |
| - Review & Q&A UI                                    |
| - Favorites/Wishlist UI                              |
| - Moderation Queue UI                                |
| - Analytics Dashboard UI                             |
| - Experiment Management UI                           |
| - Admin & Monitoring UI                              |
+------------------------------+-----------------------+
                               |
                               | HTTPS REST API
                               v
+------------------------------------------------------+
| Gin Backend                                          |
| - Authentication & Session Management                |
| - RBAC / Permission Checks                           |
| - Item Service                                       |
| - Review Service                                     |
| - Q&A Service                                        |
| - Favorites/Wishlist Sync Service                    |
| - Report & Appeal Service                            |
| - Moderation Queue Service                           |
| - Sensitive Word Service                             |
| - File Upload & Quarantine Service                   |
| - Behavior Tracking Service                          |
| - Analytics Aggregation & Reporting Service          |
| - Experiment Service                                 |
| - Idempotency Token Service                          |
| - Audit Log Service                                  |
| - Backup / Recovery / Monitoring Jobs                |
+-------------------------+----------------------------+
                          |
                          v
+------------------------------------------------------+
| MySQL                                                |
| - Users / Roles / Sessions                           |
| - Items / Reviews / Images                           |
| - Q&A / Favorites / Wishlists                        |
| - Reports / Appeals / Moderation                     |
| - Experiments / Exposures / Variants                 |
| - Event Logs / Fraud Flags                           |
| - Analytics Summary Tables                           |
| - Audit Logs / Config Tables                         |
+------------------------------------------------------+

+------------------------------------------------------+
| Local File Storage                                   |
| - Review Images                                      |
| - Quarantined Uploads                                |
| - Backup Files                                       |
| - Generated Reports                                  |
+------------------------------------------------------+
````

## 5.2 Architectural Style

The application uses a layered architecture:

* Presentation layer: Vue.js SPA
* API layer: Gin route handlers and middleware
* Service layer: business logic and workflows
* Persistence layer: MySQL repositories and file storage
* Background processing layer: scheduled jobs and sync tasks

This structure supports maintainability, testability, and offline deployment.

---

## 6. Design Principles

### 6.1 Offline-First

All primary functions must work entirely inside the local network without external APIs or cloud services.

### 6.2 Consistency for Critical Writes

Reviews, reports, and experiment exposures must use idempotency controls and database transactions where consistency matters.

### 6.3 Moderation Before Trust

Reported, suspicious, or rule-violating content must flow through moderation and quarantine mechanisms before being trusted in analytics or user-visible flows.

### 6.4 Analytics Without Sacrificing Performance

Transactional data is stored in normalized form, while background jobs build optimized reporting tables for dashboard performance.

### 6.5 Auditable Security

Security-sensitive operations, role changes, moderation actions, and configuration changes are logged with masked sensitive fields.

### 6.6 Local Security Hardening

All traffic uses HTTPS on the local network. Authentication, CAPTCHA, rate limiting, IP controls, and input validation are handled entirely within local infrastructure.

---

## 7. Functional Design

## 7.1 Authentication and Account Management

Users authenticate with username and password. Passwords are stored using Argon2id. After repeated failed logins, a local CAPTCHA challenge is required.

### Responsibilities

* User registration
* Login/logout
* Password verification
* Failed login tracking
* CAPTCHA challenge enforcement
* Role lookup and authorization
* Session creation and expiration
* Audit logging for login attempts and permission-sensitive actions

### Security Notes

* Passwords are never stored in plaintext
* CAPTCHA is generated locally
* Login thresholds are enforced per account and/or IP according to policy
* IP allowlist/denylist checks apply before protected actions where configured

---

## 7.2 Item Catalog

The catalog contains internal items such as songs, features, or content bundles.

### Responsibilities

* Create and manage items
* Categorize and archive items
* List items for browsing
* Support item detail pages with reviews, Q&A, and saved-state actions

Each item acts as the central anchor for feedback, analytics, and experiments.

---

## 7.3 Reviews and Images

Users can submit reviews that include:

* A required star rating from 1 to 5
* Optional text
* Up to 6 images

The frontend provides immediate validation and autosaves review drafts locally every 10 seconds. The UI locks the submit button with a visible “Submitting…” state for 3 seconds to reduce accidental double-submission.

### Responsibilities

* Validate ratings, text, and upload limits
* Store reviews transactionally
* Associate uploaded images in deterministic order
* Prevent duplicate submissions with idempotency tokens
* Update rating aggregates
* Support review moderation and reporting

### Image Handling

* Allowed formats: JPEG, PNG, WebP
* Maximum size: 5 MB per image
* File type is checked using format sniffing
* Hash-based duplicate detection is applied
* Suspicious files are quarantined for moderator review

---

## 7.4 Q&A

Users may ask questions about an item and post answers.

### Responsibilities

* Create questions
* Create answers linked to an item question
* Support owner edits and moderation rules
* Report Q&A content
* Include Q&A activity in analytics and moderation workflows

The design uses a flat question-and-answer structure rather than deep nested discussion threads.

---

## 7.5 Favorites and Wishlists

Users may save items into favorites and wishlists. These lists sync through the local server across devices once connected.

### Responsibilities

* Add/remove favorite items
* Add/remove wishlist items
* Sync device changes to server state
* Resolve conflicts using a deterministic sync rule

A last-write-wins model is acceptable for this lightweight list synchronization.

---

## 7.6 Reporting, Appeals, and Moderation

Moderators work from a unified queue that includes:

* Reported reviews
* Reported Q&A content
* Appeals
* Sensitive-word matches
* Quarantined files
* Fraud-suspected content requiring review

### Moderation Outcomes

* Approved
* Rejected
* Needs Edit

Each moderation decision requires notes.

### Responsibilities

* Accept reports
* Route appeals
* Apply moderation rules
* Record internal and user-visible notes where needed
* Update queue state
* Expose resolution history to authorized roles

Queue prioritization is based on severity, number of reports, quarantine status, fraud suspicion, and age.

---

## 7.7 Sensitive-Word Rules

Sensitive-word filtering is a configurable rule-based moderation layer.

### Responsibilities

* Store rule lists and versions
* Scan supported text-bearing content
* Trigger moderation actions or queue entries
* Track flagged terms
* Allow authorized updates without redeploy

These rules apply consistently across reviews, Q&A, report descriptions, and other supported user-generated text.

---

## 7.8 Analytics Dashboard

Product Analysts use an interactive dashboard to analyze review and behavior data.

### Supported Filters

* Item
* Time window
* Sentiment label
* Keywords

### Visualizations

* Trends
* Heatmaps
* Word clouds
* Topic distributions
* Co-occurrence relationships

### Drill-Down

Analysts can navigate from aggregate views to individual review sessions and event trails.

### Responsibilities

* Query transactional and reporting tables
* Apply role-based access to analytics views
* Load saved filter/view definitions
* Support efficient filtering and drill-down
* Exclude fraud-suspected data by default unless authorized otherwise

---

## 7.9 Saved Views and Internal Sharing

Analysts can save dashboard configurations and share them as read-only internal links that expire after 7 days.

### Responsibilities

* Save filter state and visualization state
* Generate internal signed share links
* Enforce read-only access
* Expire links automatically
* Support early revocation
* Allow recipients to copy a shared view into their own workspace without modifying the original

Shared links should still require authenticated internal access.

---

## 7.10 A/B Experimentation

Product Analysts can create experiments for UI or wording changes and deploy them using offline canary rollout.

### Experiment Capabilities

* Define variants
* Configure rollout percentage
* Assign users or sessions consistently
* Record exposures and outcomes
* Show confidence indicators
* Keep or roll back variants

### Responsibilities

* Manage experiment lifecycle
* Persist user/session assignment
* Record experiment exposures idempotently
* Display experiment health and decision support metrics
* Preserve history even after rollback

Rollbacks stop future exposures while retaining prior experiment data for analysis.

---

## 7.11 Behavior Tracking

The system records:

* Impressions
* Clicks
* Dwell time
* Favorites
* Shares
* Comments

### Tracking Rules

* Session attribution is required
* Duplicate events within 2 seconds collapse
* Dwell time is stored in 1-second buckets
* Dwell time is capped at 10 minutes per session
* Fraud rules flag accounts exceeding 300 events per hour
* Fraud rules also flag repeated identical sequences across sessions

### Responsibilities

* Track behavioral events efficiently
* Deduplicate noisy events
* Attribute events to sessions
* Exclude suspicious activity from analytics until reviewed
* Support configurable preference scoring

Only active engagement should count toward dwell time.

---

## 7.12 Preference Scoring

Preference scoring uses configurable per-event weights editable by Product Analysts without redeploy.

### Responsibilities

* Store scoring weight versions
* Allow updates through the UI
* Recompute downstream metrics as required
* Audit all configuration changes
* Support rollback to prior scoring versions

---

## 8. Data Design

## 8.1 Core Entities

### User

Stores account credentials and role membership.

Key fields:

* id
* username
* passwordHash
* roleId
* status
* createdAt
* updatedAt

### Role

Defines permission groups.

Key fields:

* id
* name
* description

### Item

Represents an internal product object under review.

Key fields:

* id
* title
* type
* category
* description
* status
* createdAt

### Review

Represents a user review for an item.

Key fields:

* id
* itemId
* userId
* rating
* text
* status
* createdAt
* updatedAt

### ReviewImage

Stores image metadata for review uploads.

Key fields:

* id
* reviewId
* filePath
* fileHash
* mimeType
* displayOrder
* quarantineStatus

### Question

Represents an item-specific question.

Key fields:

* id
* itemId
* userId
* text
* status
* createdAt

### Answer

Represents an answer to a question.

Key fields:

* id
* questionId
* userId
* text
* status
* createdAt

### Favorite

Represents a favorited item.

Key fields:

* id
* userId
* itemId
* createdAt
* updatedAt

### Wishlist

Represents a wishlisted item.

Key fields:

* id
* userId
* itemId
* createdAt
* updatedAt

### Report

Represents a user report against content.

Key fields:

* id
* reporterId
* targetType
* targetId
* category
* description
* status
* createdAt

### Appeal

Represents an appeal against a moderation decision.

Key fields:

* id
* targetType
* targetId
* requesterId
* reason
* status
* createdAt
* resolvedAt

### ModerationDecision

Stores moderation resolution details.

Key fields:

* id
* targetType
* targetId
* moderatorId
* outcome
* internalNotes
* visibleNotes
* createdAt

### SensitiveWordRule

Stores moderation rules.

Key fields:

* id
* term
* severity
* version
* isActive

### Session

Represents a browsing or review session for tracking.

Key fields:

* id
* userId
* startedAt
* endedAt
* deviceKey

### EventLog

Stores behavior and experiment-related events.

Key fields:

* id
* sessionId
* userId
* itemId
* eventType
* payload
* dedupeFingerprint
* createdAt
* fraudFlag

### FraudFlag

Stores suspicious activity cases.

Key fields:

* id
* userId
* sessionId
* reason
* status
* reviewedBy
* reviewedAt

### Experiment

Stores experiment definitions.

Key fields:

* id
* name
* targetType
* status
* rolloutPercent
* startAt
* endAt

### ExperimentVariant

Stores experiment variants.

Key fields:

* id
* experimentId
* key
* configPayload

### ExperimentExposure

Stores assigned variant exposures.

Key fields:

* id
* experimentId
* variantId
* sessionId
* userId
* createdAt

### SavedAnalyticsView

Stores saved dashboard views.

Key fields:

* id
* ownerId
* name
* filterConfig
* chartConfig
* createdAt

### SharedViewLink

Stores expiring share links.

Key fields:

* id
* savedViewId
* token
* expiresAt
* revokedAt

### PreferenceWeightConfig

Stores event-weight configuration.

Key fields:

* id
* version
* config
* editedBy
* effectiveAt

### AuditLog

Stores masked, append-only audit events.

Key fields:

* id
* actorId
* action
* resourceType
* resourceId
* maskedDetails
* createdAt

### IdempotencyToken

Stores request-token usage for critical operations.

Key fields:

* id
* token
* requestType
* requestFingerprint
* responseSnapshot
* expiresAt

### IpRule

Stores allowlist and denylist entries.

Key fields:

* id
* ipOrRange
* ruleType
* reason
* createdBy

### BackupRun

Stores backup execution history.

Key fields:

* id
* startedAt
* completedAt
* status
* location
* notes

### RecoveryDrill

Stores weekly recovery drill history.

Key fields:

* id
* startedAt
* completedAt
* status
* notes

---

## 9. Security Design

## 9.1 Authentication Security

* Argon2id password hashing
* Local-only CAPTCHA after repeated login failures
* Rate limiting for API access
* Role-based access checks for protected operations

## 9.2 Transport Security

* HTTPS required for all traffic on the local network
* Certificates are provisioned and trusted locally

## 9.3 Input Security

* CSRF protection for state-changing requests
* XSS-safe rendering and output encoding
* SQL injection prevention through parameterized queries
* Request validation for all API inputs

## 9.4 Network Access Security

* IP allowlist/denylist support
* Denylist takes precedence
* Administrative changes to network rules are audited

## 9.5 File Upload Security

* MIME and signature validation
* Size checks
* Duplicate detection using file hashes
* Suspicious file quarantine
* No external scanning dependency

## 9.6 Audit Security

* Permission and configuration changes are logged
* Sensitive fields are masked prior to log persistence
* Audit data is append-only from the application perspective

---

## 10. Data Consistency and Idempotency

Critical operations include:

* Posting a review
* Filing a report
* Recording an experiment exposure

These operations must be idempotent for 10 minutes using server-side request token storage.

### Design

* Client sends an idempotency token
* Backend stores token usage and request fingerprint
* Duplicate request within 10 minutes returns the original response
* Database transactions protect rating and related aggregate consistency

This avoids duplicate writes caused by retries or UX race conditions.

---

## 11. Analytics and Reporting Performance

MySQL is the source of truth, but dashboard responsiveness requires optimized reporting structures.

### Design

* Transactional tables store canonical events and content
* Scheduled ETL-style jobs update analytics summary tables
* Rating summaries refresh within 60 seconds
* Report generation jobs prepare precomputed datasets
* Fraud-suspected records are excluded by default until cleared

This balances correctness with dashboard performance.

---

## 12. Scheduled Jobs

The system includes background jobs for:

* Rating summary refresh
* Analytics table sync
* Shared-link expiration
* Nightly backups at 2:00 AM
* Weekly recovery drills
* Monitoring snapshots and operational rollups

All scheduled jobs run locally and are visible through operational dashboards.

---

## 13. Reliability and Operations

## 13.1 Monitoring

The platform provides internal dashboards for:

* Exception counts
* API latency
* Error rate
* Queue depth
* Job runtimes
* Backup status
* Recovery drill results
* Storage usage
* Sync freshness

## 13.2 Backups

Backups are:

* Performed nightly at 2:00 AM
* Stored locally
* Encrypted at rest
* Rotated according to retention policy

## 13.3 Recovery Drills

Recovery drills run weekly in an isolated local environment to verify that:

* Database backups restore correctly
* Uploaded files restore correctly
* Core services start successfully after restore

---

## 14. API Design Overview

The backend exposes REST-style APIs grouped by domain:

* Auth
* Users and Roles
* Items
* Reviews and Images
* Q&A
* Favorites and Wishlists
* Reports and Appeals
* Moderation Queue
* Sensitive Word Rules
* Analytics
* Saved Views
* Experiments
* Event Tracking
* Audit and Operations
* Backups and Recovery status

### API Characteristics

* JSON-based request and response bodies
* Multipart upload support for images
* Strong request validation
* Role-aware authorization
* Idempotency support for critical writes
* Consistent error responses

---

## 15. Error Handling

Errors are categorized into:

* Validation errors
* Authentication errors
* Authorization errors
* Moderation state errors
* Conflict and idempotency errors
* Upload validation errors
* Quarantine and fraud-review restrictions
* Operational failures

### Principles

* Do not leak internal implementation details
* Return actionable validation messages
* Log security-relevant failures
* Distinguish user-correctable errors from system failures

---

## 16. Key Design Decisions

### 16.1 MySQL as Source of Truth

Chosen because:

* Strong transactional guarantees
* Suitable relational structure for users, content, moderation, and analytics metadata
* Familiar operational model for local deployment

Tradeoff:

* Rich analytics may require additional summary tables and scheduled sync jobs

### 16.2 Offline Charting and Analysis

Chosen because:

* Internet-independent analytics are required
* Analysts need rich local dashboards

Tradeoff:

* All scoring, sentiment, and confidence logic must be implemented locally

### 16.3 Idempotency Tokens for Critical Writes

Chosen because:

* Prevents duplicate writes caused by retries
* Aligns with required 10-minute duplicate protection window

Tradeoff:

* Requires token storage and response replay logic

### 16.4 Quarantine Instead of External Scanning

Chosen because:

* External scanning is disallowed
* Moderator review is required for suspicious files

Tradeoff:

* Moderator workload may increase

### 16.5 Reporting Tables for Fast Analytics

Chosen because:

* Transactional tables alone may be too slow for rich dashboard filters and drill-down

Tradeoff:

* Requires scheduled sync and reconciliation monitoring

---

## 17. Risks

### 17.1 Product Risks

* Analysts may interpret confidence indicators inconsistently without clear documentation
* Draft autosave expectations may expand beyond the initial review scope
* Shared view access rules may be misunderstood if read-only behavior is not clear

### 17.2 Security Risks

* Improper local certificate trust setup could weaken HTTPS usage
* Weak CAPTCHA design could allow brute-force automation
* Overly broad internal network access could bypass intended isolation

### 17.3 Operational Risks

* Analytics summary jobs could lag behind transactional truth
* Recovery drills may not cover all media restoration scenarios unless automated well
* Moderation queue volume could grow if fraud and quarantine rules are too aggressive

---

## 18. Future Enhancements

Potential future extensions include:

* Advanced local NLP for sentiment and topic extraction
* Better experiment decision-support tooling
* More granular item taxonomy and tagging
* Moderator workload balancing tools
* Richer synchronization support for additional offline-capable client devices
* Exportable internal analytics reports for local archival use

---

## 19. Open Assumptions

This design assumes:

* Elevated roles are assigned internally by administrators
* Shared analytics links still require authenticated access
* Fraud-suspected data is excluded by default from dashboards
* All runtime dependencies are bundled locally
* Comments either map to Q&A/review-related interactions or are later specified as a distinct feature
* Session-based analytics can be implemented without violating local privacy 