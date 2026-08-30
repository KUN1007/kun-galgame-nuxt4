-- 081: where a forum gid went when catalog merged its work away.
--
-- Catalog merges duplicate works. `retireSource` nulls the source row's site and
-- product_work_id and soft-deletes it, so a forum gid bound to that work stops
-- resolving and its page answers 404 with no trail: on 2026-08-29 a batch of
-- 3,242 work merges killed 30 forum pages, one of them holding a download
-- resource posted 9 minutes 45 seconds before the merge ran.
--
-- Catalog announces the survivor on /v2/catalog/redirects, but only in catalog
-- ids, and the dead work's claim -- the only thing that named the gid -- is
-- erased by the same statement. So the mapping cannot be recovered later; it has
-- to be written down when the merge sync folds the row, and this is where.
--
-- old_gid carries no foreign key: the fold deletes the galgame row it names.
-- new_gid does not either, so a later merge of the survivor can be chased by
-- rewriting new_gid rather than by deleting the chain.

BEGIN;

CREATE TABLE IF NOT EXISTS galgame_redirect (
  old_gid integer     PRIMARY KEY,
  new_gid integer     NOT NULL,
  created timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_galgame_redirect_new ON galgame_redirect (new_gid);

COMMENT ON TABLE galgame_redirect IS
  'Forum-side merge ledger (081). old_gid no longer exists in galgame; /galgame/:gid answers moved_to new_gid so the browser can 301.';

COMMIT;
