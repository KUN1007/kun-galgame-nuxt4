-- 085: give a code prize a claim deadline.
--
-- A winner who never pressed 领取 left the sealed code sitting in
-- topic_lottery_code forever: `forfeited` existed as a state but only the topic
-- author could set it by hand, and nothing told them to. The deadline is
-- stamped on the entry at draw time rather than derived from won_at, so
-- changing the grace period later cannot retroactively expire a win that was
-- promised a different date.
BEGIN;

ALTER TABLE topic_lottery_entry ADD COLUMN IF NOT EXISTS claim_deadline timestamptz;

CREATE INDEX IF NOT EXISTS idx_topic_lottery_entry_claim_due
  ON topic_lottery_entry (claim_deadline)
  WHERE fulfillment = 'pending' AND claim_deadline IS NOT NULL;

COMMIT;
