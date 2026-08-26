-- 079: cache catalog's editorial display verdict on the local galgame row.
--
-- The local lists (/galgame and the entity pages with a resource filter) page
-- ids in SQL and only then hydrate them from catalog, which drops whatever the
-- reader's content_limit refuses. An SFW reader therefore got ~10 of 24 cards
-- on every page while the pager still counted all 7,781 rows. Filtering needs
-- the verdict in SQL, so keep a copy of it here.
--
-- NULL means "not synced yet" and is NOT filtered out, so this deploys ahead of
-- the sync without changing what any page shows. Catalog stays the source of
-- truth: the hydrate gate is still the real door, this column only decides
-- which ids reach it.
--
-- Values are 'sfw' / 'nsfw', the same two the wire uses. 005 dropped a column of
-- this name that held the wiki's own copy of the field; this one is a cache with
-- a refresh job behind it, not a locally edited value.
--
-- No backfill: the galgame content-limit sync fills every row on its next run.

BEGIN;

ALTER TABLE galgame ADD COLUMN IF NOT EXISTS content_limit text;

COMMENT ON COLUMN galgame.content_limit IS
  'Cache of catalog''s display verdict (079). NULL = not synced yet, and not filtered. Written only by the content-limit sync; catalog owns the value.';

COMMIT;
