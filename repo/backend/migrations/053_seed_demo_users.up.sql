-- Seed demo users for each role so a fresh `docker-compose up` can be used
-- immediately without any manual SQL setup. Credentials are documented in
-- repo/README.md (Default Credentials section). Passwords are Argon2id hashes
-- pre-computed with internal/pkg/hash.HashPassword.
--
-- admin            / AdminPass1
-- moderator        / ModeratorPass1
-- product_analyst  / AnalystPass1
-- demo_user        / DemoPass1
--
-- INSERT IGNORE keeps this migration idempotent if the accounts were created
-- by an earlier run or manually. UUIDs are stable so tests can rely on them.

INSERT IGNORE INTO users (uuid, username, email, password_hash, role, is_active)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'admin', 'admin@local.test',
     '$argon2id$v=19$m=65536,t=3,p=2$NnN+hH7gNKk9wCrVWpaHHg$GKMTG39mUdhUIs+YRqrmjHoPWQz5dw0uqQ1kGufnwy4',
     'admin', 1),
    ('22222222-2222-2222-2222-222222222222', 'moderator', 'moderator@local.test',
     '$argon2id$v=19$m=65536,t=3,p=2$tnqmiohEBvRyPryZvo06CQ$VwMWEi5s2sC9pedkpziVDT0SPXnt65ZqkXyyGUXTTKc',
     'moderator', 1),
    ('33333333-3333-3333-3333-333333333333', 'product_analyst', 'analyst@local.test',
     '$argon2id$v=19$m=65536,t=3,p=2$Jqb9qae+v7LoCi6ERNU5xQ$0Byn0TzuSDD9bihIp9bRy9HsgtySL8dFSgDolOVdh88',
     'product_analyst', 1),
    ('44444444-4444-4444-4444-444444444444', 'demo_user', 'demo@local.test',
     '$argon2id$v=19$m=65536,t=3,p=2$G1QmxUA65Aphevw93pVm+A$mC4gZqVVqW4nZFx50jR0yJRQfp6FaEa+EuByoeDNoOE',
     'regular_user', 1);
