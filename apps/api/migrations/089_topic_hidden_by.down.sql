DROP INDEX IF EXISTS idx_topic_hidden_status_update;
ALTER TABLE topic DROP COLUMN IF EXISTS hidden_by;
