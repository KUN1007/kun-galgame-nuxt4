-- Folds the gallery back to its first image; the rest are lost, which is why
-- the up migration is the one to trust.
BEGIN;

ALTER TABLE topic_lottery_prize ADD COLUMN IF NOT EXISTS image_hash text NOT NULL DEFAULT '';

UPDATE topic_lottery_prize
   SET image_hash = COALESCE(image_hashes ->> 0, '')
 WHERE jsonb_array_length(image_hashes) > 0;

ALTER TABLE topic_lottery_prize DROP COLUMN IF EXISTS image_hashes;
ALTER TABLE topic_lottery_prize DROP COLUMN IF EXISTS point_mode;
ALTER TABLE topic_lottery_entry DROP COLUMN IF EXISTS point_awarded;

COMMIT;
