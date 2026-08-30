-- The merge fold drops a child row when the survivor's peer already has one:
-- a user who liked, rated, or collected both copies cannot keep two rows under
-- the one game that is left. For galgame_rating that DELETE destroys a written
-- review — a dev rehearsal on 2026-08-30 confirmed an 800-word review and
-- another user's like on it disappearing with no trace — and all 3265 ratings
-- in production carry text. Nothing the fold cannot move is deleted without
-- landing here first.
--
-- `row` is the whole source row as jsonb. A galgame_rating document also
-- carries a `likes` array, because galgame_rating_like cascades off the rating
-- and would otherwise be unrecoverable.
CREATE TABLE IF NOT EXISTS galgame_merge_discarded (
  id         bigserial   PRIMARY KEY,
  old_gid    integer     NOT NULL,
  new_gid    integer     NOT NULL,
  table_name text        NOT NULL,
  row        jsonb       NOT NULL,
  created    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_galgame_merge_discarded_old ON galgame_merge_discarded (old_gid);
