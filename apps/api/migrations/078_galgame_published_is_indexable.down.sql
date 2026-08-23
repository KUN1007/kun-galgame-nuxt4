-- Restores the column comment only. The backfill cannot reconstruct
-- claim-driven published values; rolling this forward again is 078 up.

COMMENT ON COLUMN galgame.published IS
  'Public visibility (180). Row existence = local interaction container; this flag = listable. Written only by the claim-event cron.';
