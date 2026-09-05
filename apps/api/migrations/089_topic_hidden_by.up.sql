-- 089: record WHO hid a topic (topic-permission wave).
-- topic.status = 1 was one drawer shared by three actors — the author's own
-- hide, a moderator's hide, and the trust enforcement callback (whose Hide and
-- Remove both wrote status = 1) — so the API could not answer "who hid this",
-- and an author could undo a moderator's or T&S's decision from the profile
-- 已隐藏 tab with one click. hidden_by records the actor: '' (visible),
-- 'author', 'moderator', 'trust'.
--
-- Existing hidden rows are backfilled 'author': the actor was never recorded,
-- and 'author' preserves exactly what every author can already do today (undo
-- their own hide); any other value would lock authors out of topics they hid
-- themselves.
ALTER TABLE topic ADD COLUMN IF NOT EXISTS hidden_by text NOT NULL DEFAULT '';

UPDATE topic SET hidden_by = 'author' WHERE status = 1 AND hidden_by = '';

-- The admin hidden-topic list is the only reader filtering on status = 1.
CREATE INDEX IF NOT EXISTS idx_topic_hidden_status_update
    ON topic (status_update_time DESC) WHERE status = 1;
