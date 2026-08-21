BEGIN;

DROP INDEX IF EXISTS idx_galgame_website_status;
DROP INDEX IF EXISTS idx_galgame_website_tag_group_id;
ALTER TABLE galgame_website_tag DROP COLUMN IF EXISTS group_id;
DROP TABLE IF EXISTS galgame_website_tag_group;
ALTER TABLE galgame_website_category DROP COLUMN IF EXISTS sort_order;
ALTER TABLE galgame_website DROP COLUMN IF EXISTS status;

COMMIT;
