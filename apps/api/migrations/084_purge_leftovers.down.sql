-- Only the constraints come back; the orphan rows and the half-dead chat rooms
-- they were added to prevent are gone for good.
BEGIN;

ALTER TABLE topic_reaction DROP CONSTRAINT IF EXISTS topic_reaction_topic_id_fkey;
ALTER TABLE topic_reply_reaction DROP CONSTRAINT IF EXISTS topic_reply_reaction_topic_reply_id_fkey;
ALTER TABLE topic_lottery DROP CONSTRAINT IF EXISTS topic_lottery_topic_id_fkey;

COMMIT;
