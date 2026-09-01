-- 088 down: drop the search trgm indexes. Keep pg_trgm; other objects may use it.

BEGIN;

DROP INDEX IF EXISTS idx_topic_title_trgm;
DROP INDEX IF EXISTS idx_topic_content_trgm;
DROP INDEX IF EXISTS idx_topic_category_trgm;
DROP INDEX IF EXISTS idx_topic_reply_content_trgm;
DROP INDEX IF EXISTS idx_topic_comment_content_trgm;

COMMIT;
