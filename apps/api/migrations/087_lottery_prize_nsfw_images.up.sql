-- 087: let a prize mark which of its images are adult content.
--
-- nsfw_hashes is a subset of image_hashes rather than a per-prize boolean: a
-- partner giving away ten photos of one item usually has both safe product
-- shots and R18 ones in the same gallery, and a whole-prize flag would hide the
-- safe ones too. Existing rows have nothing marked, which is the right default
-- for images that were uploaded before the author could say.
BEGIN;

ALTER TABLE topic_lottery_prize
  ADD COLUMN IF NOT EXISTS nsfw_hashes jsonb NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON COLUMN topic_lottery_prize.nsfw_hashes IS
  'Subset of image_hashes the author marked as adult content. Their URLs are withheld from readers whose content setting is SFW.';

COMMIT;
