-- Drops the cache column. The values are catalog's, so nothing is lost: 079 up
-- plus one sync run rebuilds them.

BEGIN;

ALTER TABLE galgame DROP COLUMN IF EXISTS content_limit;

COMMIT;
