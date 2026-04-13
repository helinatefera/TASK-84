## Open Questions & Assumptions

### 1. Account Registration, Identity, and Roles

– Question: The prompt allows users to register with username and password and defines four roles, but it does not specify username rules, password policy, account activation, whether users can hold multiple roles, or who assigns elevated roles.
– Assumption: Regular Users self-register, while Administrator, Moderator, and Product Analyst roles are assigned by Administrators; usernames must be unique and normalized for login; each account has one primary role.
– Solution: Enforce username format and uniqueness, require a minimum password policy, hash passwords with Argon2id, allow self-registration only for Regular Users, and implement administrator-only role assignment with audit logging.

### 2. Item Catalog Ownership and Scope

– Question: Users can browse internal items, but the prompt does not specify who creates items, how items are categorized, or whether items can be archived or hidden.
– Assumption: Items are managed internally by Administrators or Product Analysts and may need lifecycle states.
– Solution: Add item management with create, edit, categorize, publish, and archive actions, plus role-based permissions.

### 3. Review Submission, Editing, and Validation Rules

– Question: Reviews support a 1–5 star rating, optional text, and up to 6 images, but the prompt does not define review edit/delete rules, text length limits, or whether ratings are integers only.
– Assumption: Reviews use whole-number ratings only, text is optional with a reasonable maximum length, and users can edit or delete their own reviews within policy limits.
– Solution: Restrict ratings to integers 1–5, enforce text length limits in Vue and Gin, and support controlled review edit/delete with audit history.

### 4. Review Images, Ordering, and Duplicate Handling

– Question: Reviews may include up to 6 images, but the prompt does not specify image order, cover-image behavior, or whether duplicate images are blocked globally, per review, or per user.
– Assumption: Upload order should be preserved, and duplicate detection should work system-wide with moderation override if needed.
– Solution: Store ordered image positions, use the first image as default preview, compute file hashes, and warn or block exact duplicates based on policy.

### 5. Draft Autosave and Submission Lock Behavior

– Question: The interface autosaves draft text every 10 seconds locally and prevents accidental double-posting with a visible 3-second “Submitting…” lock, but it does not define what parts of the draft are saved, how long drafts persist, or whether backend duplicate prevention is also required.
– Assumption: Autosave applies to the current review draft in the browser, and UI locking must be reinforced by server-side idempotency.
– Solution: Save review draft text, rating, and item context to local storage until submit or discard, disable submit for 3 seconds, and require backend idempotency keys for posting reviews.

### 6. Q&A and Comment Model

– Question: Users can ask and answer Q&A entries, and behavior tracking includes comments, but the prompt does not define whether questions/answers can be edited or deleted, whether nested replies are allowed, or whether comments are a separate feature from Q&A.
– Assumption: A simple one-question, many-answers structure is intended, with no deep nesting, and comments should map to a clearly defined supported interaction.
– Solution: Implement flat Q&A threads with owner edit/delete rules and either define a separate comment feature explicitly or map “comments” to existing Q&A/review interactions in the analytics model.

### 7. Favorites, Wishlists, and Cross-Device Sync

– Question: Users can maintain favorites and wishlists that sync across devices, but the prompt does not define whether favorites and wishlists are distinct, or how sync conflicts are handled when offline changes occur on multiple devices.
– Assumption: Favorites and wishlists are separate saved-item collections, and simple conflict handling is acceptable.
– Solution: Model favorites and wishlists as distinct lists and use timestamp-based last-write-wins sync when devices reconnect to the local server.

### 8. Reports, Appeals, and Moderator Outcomes

– Question: Moderators process reports and appeals with outcomes of approved, rejected, or needs edit plus required moderator notes, but the prompt does not define report categories, appeal eligibility, appeal limits, decision finality, or whether moderator notes are internal or user-visible.
– Assumption: Controlled categories are needed, content owners may appeal once, and moderation should support both internal notes and user-facing explanations.
– Solution: Define report categories, allow one appeal per eligible moderation case, treat resolved appeals as final unless reopened by an Administrator, and store both internal notes and optional user-visible resolution notes.

### 9. Unified Moderator Queue and Prioritization

– Question: Moderators work from a unified queue, but the prompt does not define prioritization, queue filtering, or how quarantined files and fraud-suspected content are mixed into review workflows.
– Assumption: Higher-risk content should be reviewed first and the queue must support multiple content types.
– Solution: Build a queue sorted by severity, report count, fraud flags, quarantine status, and age, with filters by content type and moderation state.

### 10. Sensitive-Word Rules and Moderation Scope

– Question: Moderators apply sensitive-word rules, but the prompt does not define who can edit these rules, whether they are versioned, or which content types they apply to.
– Assumption: Sensitive-word rules must be centrally managed and consistently applied across user-generated text.
– Solution: Version and audit the rule list, allow authorized admins/moderators to manage it, and apply it to reviews, Q&A, reports, and any other text-bearing content.

### 11. Analytics Dimensions, Sessions, and Drill-Down

– Question: Analysts can filter by item, time window, sentiment label, and keywords, then drill down from aggregate views to individual review sessions, but the prompt does not define session boundaries, time-window granularities, sentiment labeling, or exact chart computations.
– Assumption: Sessions are first-class analytics entities, sentiment is generated locally, and chart definitions must be explicitly documented.
– Solution: Define session attribution rules, support preset and custom time ranges, implement local sentiment labeling with override support, and document exact metric formulas for trends, heatmaps, word clouds, topic distributions, and co-occurrence views.

### 12. Saved Analytics Views and Expiring Share Links

– Question: Custom analytics views can be saved and shared as read-only links expiring after 7 days, but the prompt does not define whether links require login, whether they can be revoked early, or whether recipients can duplicate them into personal views.
– Assumption: Shared links remain internal and authenticated, should be revocable, and may be cloned without editing the source view.
– Solution: Generate authenticated signed read-only links with 7-day expiry, support early revocation, and allow recipients to copy a shared view into their own workspace.

### 13. A/B Testing, Canary Rollout, and Confidence Indicators

– Question: Analysts can run A/B tests with offline canary rollout and on-screen confidence indicators, but the prompt does not define assignment method, rollout unit, stickiness, statistical method, or rollback effects on historical data.
– Assumption: Stable local assignment and explainable offline statistics are required.
– Solution: Use deterministic assignment by user or session, support percentage-based canary rollout, define a local confidence method such as Bayesian or confidence-interval reporting, preserve historical data after rollback, and stop only future exposures when a variant is rolled back.

### 14. Idempotency Token Design

– Question: Posting a review, filing a report, and recording an experiment exposure must be idempotent for 10 minutes using server-side request tokens, but token generation, storage, and response replay behavior are not specified.
– Assumption: Clients submit unique request keys and identical retries should return the original successful result.
– Solution: Store idempotency keys with request fingerprints and response metadata for 10 minutes and replay the original response for duplicates.

### 15. Rating Aggregates, Transactions, and Fraud Exclusion

– Question: Database writes must keep rating aggregates consistent, summaries refresh within 60 seconds, and suspected fraud must be excluded until cleared, but the refresh mechanism and fraud-clearing workflow are not defined.
– Assumption: Transactional writes plus background summary refresh are required, and fraud review is moderator- or admin-driven.
– Solution: Use MySQL transactions for rating writes, refresh denormalized summary tables within 60 seconds, exclude flagged data by default, and add fraud-review states with reviewer notes and authorized override views.

### 16. Behavior Tracking, Deduplication, and Anti-Fraud Rules

– Question: The system records impressions, clicks, dwell time, favorites, shares, and comments with session attribution and deduplication, but the prompt does not define deduplication keys, session expiration, whether inactive tabs count toward dwell time, or what happens after anti-fraud thresholds are exceeded.
– Assumption: Sessions expire after inactivity, only active engagement counts toward dwell time, and suspicious behavior should be flagged rather than automatically blocked.
– Solution: Define a session model, count dwell time only while the page is active, deduplicate events using a fingerprint of event type plus target plus actor within 2 seconds, and flag suspicious activity for moderator/admin review while excluding it from default analytics.

### 17. Preference Scoring Configuration

– Question: Preference scoring uses configurable per-event weights editable by Product Analysts without redeploy, but the prompt does not specify versioning, approval, rollback, or audit requirements for these changes.
– Assumption: Weight changes materially affect analytics and must be traceable.
– Solution: Store scoring configurations with version history, effective timestamps, editor identity, rollback support, and audit logs.

### 18. Authentication Hardening and Login Abuse Controls

– Question: Security includes Argon2id password hashing and a local-only CAPTCHA after 5 failed logins in 15 minutes, but the prompt does not define CAPTCHA type, failure-counter scope, or password reset behavior.
– Assumption: The CAPTCHA must be generated locally, and failures should be tracked by both account and IP.
– Solution: Implement a local arithmetic or image CAPTCHA, trigger it when either account or IP exceeds the threshold, and provide a secure offline password reset workflow.

### 19. Network Access Controls and Web Security Protections

– Question: The system requires IP allowlist/denylist support plus CSRF, XSS, SQL injection protections, and HTTPS across the local network, but precedence rules and implementation details are not specified.
– Assumption: Denylist takes precedence, all unsafe input must be sanitized or parameterized, and frontend/backend session protection must be explicit.
– Solution: Enforce denylist-before-allowlist logic, use parameterized SQL and output encoding, implement CSRF protection appropriate for the chosen auth/session model, and deploy with locally trusted HTTPS certificates.

### 20. File Upload Validation and Quarantine Workflow

– Question: Uploads are restricted to JPEG/PNG/WebP up to 5 MB with format sniffing and suspicious-file quarantine, but the prompt does not define quarantine access rules or user-visible behavior while a file is quarantined.
– Assumption: Quarantined files must be hidden from regular users until reviewed.
– Solution: Validate MIME type and file signature, hash each file, move suspicious files into a separate quarantine area, and expose them only to moderators and administrators with clear review states.

### 21. Audit Logs and Sensitive Field Masking

– Question: Permission audit logs with masked sensitive fields are required, but the exact audited events and masking rules are not specified.
– Assumption: All security-sensitive and configuration-changing actions must be logged, and masking must be centralized.
– Solution: Log authentication, role changes, moderation actions, analytics configuration changes, IP list changes, experiment changes, and permission checks, while applying a single masking policy to protected values before storage.

### 22. Reliability Monitoring and Operational Dashboards

– Question: Reliability requires structured exception and performance monitoring dashboards, but the exact metrics, retention, and operational alerts are not specified.
– Assumption: Offline observability is required for both application health and background jobs.
– Solution: Track API latency, error rates, queue depth, job runtimes, storage use, backup status, and analytics sync health in local dashboards with retention settings.

### 23. Backups, Recovery Drills, and Disaster Readiness

– Question: The system requires nightly backups at 2:00 AM with weekly recovery drills, but backup destination, encryption, retention policy, and drill success criteria are not specified.
– Assumption: Backups must remain local, encrypted, and verifiable.
– Solution: Write encrypted backups to local storage with retention rotation, automate weekly restore drills in an isolated environment, and record results for operators.

### 24. Analytics Reporting Tables and Scheduled Data Sync

– Question: Scheduled report generation and data sync between tables must keep analytics views fast and recoverable, but the target schema, sync direction, and consistency approach are not defined.
– Assumption: Separate reporting tables or materialized summary tables are intended.
– Solution: Build scheduled ETL-style jobs that populate optimized analytics tables from transactional MySQL tables with checkpoints and reconciliation logs.

### 25. Offline Runtime Dependency Policy

– Question: The portal must run entirely within an offline network, but the prompt does not specify whether runtime dependencies such as fonts, scripts, chart assets, or CAPTCHA resources may come from external CDNs or services.
– Assumption: All runtime assets and dependencies must be locally bundled.
– Solution: Package all frontend libraries, static assets, charts, and security components within the deployment and prohibit any external network calls at runtime.
