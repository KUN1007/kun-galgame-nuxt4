-- Dropping the column loses which images were marked adult; the images
-- themselves stay in image_hashes and come back unmarked.
BEGIN;

ALTER TABLE topic_lottery_prize DROP COLUMN IF EXISTS nsfw_hashes;

COMMIT;
