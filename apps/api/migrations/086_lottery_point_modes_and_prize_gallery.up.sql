-- 086: moemoepoint payout modes, and a prize carrying more than one image.
--
-- point_mode changes what point_amount means: 'fixed' keeps the original
-- reading (every winner gets that many), while 'split' and 'random' read it as
-- a pool that is fully handed out. A pool has to be recorded per winner, so
-- topic_lottery_entry gains point_awarded — recomputing a share at read time
-- would move every winner's number whenever another winner is banned.
--
-- image_hash becomes image_hashes and is dropped rather than kept as "the first
-- image": two columns holding the same picture is how they drift. Existing rows
-- are folded into a one-element array first. The drop is safe here only because
-- 083 has not been applied in production yet, so no deployed binary selects the
-- old column; a later drop of a live column would have to wait for the deploy.
BEGIN;

ALTER TABLE topic_lottery_prize
  ADD COLUMN IF NOT EXISTS point_mode text NOT NULL DEFAULT 'fixed';

ALTER TABLE topic_lottery_prize
  ADD COLUMN IF NOT EXISTS image_hashes jsonb NOT NULL DEFAULT '[]'::jsonb;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'topic_lottery_prize'
      AND column_name = 'image_hash'
  ) THEN
    UPDATE topic_lottery_prize
       SET image_hashes = jsonb_build_array(image_hash)
     WHERE image_hash <> '' AND image_hashes = '[]'::jsonb;
  END IF;
END $$;

ALTER TABLE topic_lottery_prize DROP COLUMN IF EXISTS image_hash;

COMMENT ON COLUMN topic_lottery_prize.point_mode IS
  'fixed: point_amount per winner. split/random: point_amount is a pool shared by this prize''s winners.';

ALTER TABLE topic_lottery_entry
  ADD COLUMN IF NOT EXISTS point_awarded integer NOT NULL DEFAULT 0;

COMMENT ON COLUMN topic_lottery_entry.point_awarded IS
  'What this winner actually received, stamped at draw time. Zero for every non-point prize.';

COMMIT;
