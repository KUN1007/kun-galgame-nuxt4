-- 088: pg_trgm GIN indexes for forum search ILIKE '%kw%' queries.
--
-- WHY: SearchTopics (search_repo.go) ORs ILIKE across topic.title / topic.content
-- / topic.category with no usable index — two seq scans per search (~300ms,
-- slow-SQL). SearchReplies and SearchComments ILIKE the content columns on
-- topic_reply and topic_comment. gin_trgm_ops supports leading-wildcard ILIKE.
-- IF NOT EXISTS so this is a no-op where the orchestrator already built the
-- indexes CONCURRENTLY on production.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- SearchTopics: (t.title ILIKE ? OR t.content ILIKE ? OR t.category ILIKE ?)
CREATE INDEX IF NOT EXISTS idx_topic_title_trgm
  ON topic USING gin (title gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_topic_content_trgm
  ON topic USING gin (content gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_topic_category_trgm
  ON topic USING gin (category gin_trgm_ops);

-- SearchReplies: r.content ILIKE ? on topic_reply
CREATE INDEX IF NOT EXISTS idx_topic_reply_content_trgm
  ON topic_reply USING gin (content gin_trgm_ops);

-- SearchComments: c.content ILIKE ? on topic_comment
CREATE INDEX IF NOT EXISTS idx_topic_comment_content_trgm
  ON topic_comment USING gin (content gin_trgm_ops);

COMMIT;
