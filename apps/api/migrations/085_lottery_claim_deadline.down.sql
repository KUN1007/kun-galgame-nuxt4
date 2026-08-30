BEGIN;

DROP INDEX IF EXISTS idx_topic_lottery_entry_claim_due;
ALTER TABLE topic_lottery_entry DROP COLUMN IF EXISTS claim_deadline;

COMMIT;
