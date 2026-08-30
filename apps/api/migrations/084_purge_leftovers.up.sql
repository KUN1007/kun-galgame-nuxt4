-- 084: clean up what the user purge left behind, and make the two holes that
-- produced it impossible to reopen.
--
-- Every other child of `topic` carries ON DELETE CASCADE, but topic_reaction,
-- topic_reply_reaction and topic_lottery were created without one. The admin
-- purge is the only path that hard-deletes a topic, and it never listed those
-- three tables, so purging one account left *other* people's reactions pointing
-- at topics that no longer existed (3 topic_reaction + 1 topic_reply_reaction
-- rows on the 2026-08 production snapshot).
--
-- The second hole is the private chat room. The purge deleted the account's
-- participant row and messages but kept the room, so the counterparty was left
-- with a thread whose peer resolved to nothing — a blank name and avatar in the
-- chat list — and a later DM to the same pair would have collided with the
-- unique index on chat_room.name. A private room needs both sides; the rows
-- below delete the ones that have lost a side, including the surviving party's
-- own messages in them, because the thread is already unreachable.
BEGIN;

DELETE FROM topic_reaction r
WHERE NOT EXISTS (SELECT 1 FROM topic t WHERE t.id = r.topic_id);

DELETE FROM topic_reply_reaction r
WHERE NOT EXISTS (SELECT 1 FROM topic_reply x WHERE x.id = r.topic_reply_id);

DELETE FROM topic_lottery l
WHERE NOT EXISTS (SELECT 1 FROM topic t WHERE t.id = l.topic_id);

DELETE FROM chat_room cr
WHERE cr.type = 'private'
  AND (SELECT count(*) FROM chat_room_participant p WHERE p.chat_room_id = cr.id) < 2;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'topic_reaction_topic_id_fkey'
  ) THEN
    ALTER TABLE topic_reaction
      ADD CONSTRAINT topic_reaction_topic_id_fkey
      FOREIGN KEY (topic_id) REFERENCES topic(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'topic_reply_reaction_topic_reply_id_fkey'
  ) THEN
    ALTER TABLE topic_reply_reaction
      ADD CONSTRAINT topic_reply_reaction_topic_reply_id_fkey
      FOREIGN KEY (topic_reply_id) REFERENCES topic_reply(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'topic_lottery_topic_id_fkey'
  ) THEN
    ALTER TABLE topic_lottery
      ADD CONSTRAINT topic_lottery_topic_id_fkey
      FOREIGN KEY (topic_id) REFERENCES topic(id) ON DELETE CASCADE;
  END IF;
END $$;

COMMIT;
