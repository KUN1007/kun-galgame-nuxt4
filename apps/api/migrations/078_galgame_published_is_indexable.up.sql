-- 078: `published` stops following catalog claim state.
--
-- 方案③: catalog is the existence layer, so a kungal galgame page is reachable
-- for every work. `published` is now the sticky SEO/index flag: true once a
-- resource has been accepted on this site, and it stays true if those rows are
-- later deleted. The claim-event cron must not flip it (hidden/ban is the
-- remaining unpublish path, in application code).
--
-- Backfill is "has a resource today". Deleted history is gone, so ever-had
-- cannot be reconstructed; rows that were only claim-live with no resource
-- drop out of the sitemap, which is the point.

BEGIN;

UPDATE galgame g
SET published = EXISTS (
  SELECT 1 FROM galgame_resource r WHERE r.galgame_id = g.id
);

COMMENT ON COLUMN galgame.published IS
  'Sticky SEO/index flag (078). True after the first galgame_resource on this row; not cleared when resources are deleted. Independent of catalog claim_state.';

COMMIT;
