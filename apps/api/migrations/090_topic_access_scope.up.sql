-- 090: per-topic access scope (topic-permission wave, 读法二).
-- topic.access_scope: 'public' (default, everyone), 'login' (any signed-in
-- user), 'role' / 'users' (only subjects granted in topic_access_grant, plus
-- the author and topic.view_restricted holders). Existing rows stay 'public',
-- so behavior is unchanged until the new write faces deploy — safe to run
-- before the code ships.
ALTER TABLE topic ADD COLUMN IF NOT EXISTS access_scope text NOT NULL DEFAULT 'public';

CREATE INDEX IF NOT EXISTS idx_topic_access_scope_nonpublic
    ON topic (access_scope) WHERE access_scope <> 'public';

-- Grant rows exist only for scope 'role' (subject_value = role name) and
-- 'users' (subject_value = user id as text). The author is never stored —
-- authors always see their own topic.
CREATE TABLE IF NOT EXISTS topic_access_grant (
    topic_id      integer NOT NULL REFERENCES topic (id) ON DELETE CASCADE,
    subject_type  text    NOT NULL,
    subject_value text    NOT NULL,
    PRIMARY KEY (topic_id, subject_type, subject_value)
);

-- Feed maintenance. Before this migration only the topic's own TOPIC_CREATION
-- row was removed on hide; feed_sync_topic_reply/comment/upvote and
-- feed_sync_message never looked at the parent topic, so replies, comments,
-- upvotes and message activity of HIDDEN topics stayed in the shared home
-- feed (content excerpt visible to everyone, link 404ing since wave 1's read
-- gates). Access scopes would widen that hole, so all activity of a topic
-- that is hidden OR non-public now leaves the feed together, and the leaked
-- rows are swept at the bottom of this file.

CREATE OR REPLACE FUNCTION topic_feed_visible(p_tid integer) RETURNS boolean
LANGUAGE sql STABLE AS $$
    SELECT COALESCE(
        (SELECT status = 0 AND access_scope = 'public' FROM topic WHERE id = p_tid),
        false);
$$;

CREATE OR REPLACE FUNCTION public.feed_sync_topic() RETURNS trigger
LANGUAGE plpgsql AS $function$
DECLARE v_visible boolean; v_was boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN PERFORM feed_delete('TOPIC_CREATION', OLD.id); RETURN OLD; END IF;
    v_visible := NEW.status = 0 AND NEW.access_scope = 'public';
    IF v_visible THEN
        PERFORM feed_upsert('TOPIC_CREATION', NEW.id, NEW.user_id, 0, NEW.title, '/topic/' || NEW.id, NEW.is_nsfw, NEW.created);
    ELSE
        PERFORM feed_delete('TOPIC_CREATION', NEW.id);
    END IF;
    IF TG_OP = 'UPDATE' THEN
        v_was := OLD.status = 0 AND OLD.access_scope = 'public';
        IF v_was AND NOT v_visible THEN
            -- Children leave the feed with the topic. Message rows are keyed
            -- by link: exact or '?...' only — LIKE '/topic/<id>%' would also
            -- match topic 41209 when topic 4120 goes dark.
            DELETE FROM feed_activity fa USING topic_reply r
              WHERE fa.type = 'TOPIC_REPLY_CREATION' AND fa.source_id = r.id AND r.topic_id = NEW.id;
            DELETE FROM feed_activity fa USING topic_comment c
              WHERE fa.type = 'TOPIC_COMMENT_CREATION' AND fa.source_id = c.id AND c.topic_id = NEW.id;
            DELETE FROM feed_activity fa USING topic_upvote u
              WHERE fa.type = 'TOPIC_UPVOTE' AND fa.source_id = u.id AND u.topic_id = NEW.id;
            DELETE FROM feed_activity
              WHERE type IN ('MESSAGE_UPVOTE', 'MESSAGE_SOLUTION')
                AND (link = '/topic/' || NEW.id OR link LIKE '/topic/' || NEW.id || '?%');
        ELSIF NOT v_was AND v_visible THEN
            -- Children rejoin from the live tables when the topic becomes
            -- visible and public again.
            PERFORM feed_upsert('TOPIC_REPLY_CREATION', r.id, r.user_id, 0, COALESCE(r.content, ''), '/topic/' || NEW.id, NEW.is_nsfw, r.created)
              FROM topic_reply r WHERE r.topic_id = NEW.id AND r.status = 0;
            PERFORM feed_upsert('TOPIC_COMMENT_CREATION', c.id, c.user_id, 0, SUBSTRING(c.content, 1, 100), '/topic/' || NEW.id, NEW.is_nsfw, c.created)
              FROM topic_comment c WHERE c.topic_id = NEW.id AND c.status = 0;
            PERFORM feed_upsert('TOPIC_UPVOTE', u.id, u.user_id, 0, COALESCE(u.description, ''), '/topic/' || NEW.id, NEW.is_nsfw, u.created)
              FROM topic_upvote u WHERE u.topic_id = NEW.id;
            PERFORM feed_upsert(CASE m.type WHEN 'upvoted' THEN 'MESSAGE_UPVOTE' ELSE 'MESSAGE_SOLUTION' END,
                                m.id, m.sender_id, 0, m.content, m.link, NEW.is_nsfw, m.created)
              FROM message m
              WHERE m.type IN ('upvoted', 'solution')
                AND (m.link = '/topic/' || NEW.id OR m.link LIKE '/topic/' || NEW.id || '?%');
        END IF;
        -- A change to the topic's NSFW flag re-flags everything that hangs off it.
        IF NEW.is_nsfw IS DISTINCT FROM OLD.is_nsfw THEN
            UPDATE feed_activity fa SET is_nsfw = NEW.is_nsfw
              FROM topic_reply r   WHERE fa.type = 'TOPIC_REPLY_CREATION'   AND fa.source_id = r.id AND r.topic_id = NEW.id;
            UPDATE feed_activity fa SET is_nsfw = NEW.is_nsfw
              FROM topic_comment c WHERE fa.type = 'TOPIC_COMMENT_CREATION' AND fa.source_id = c.id AND c.topic_id = NEW.id;
            UPDATE feed_activity fa SET is_nsfw = NEW.is_nsfw
              FROM topic_upvote u  WHERE fa.type = 'TOPIC_UPVOTE'           AND fa.source_id = u.id AND u.topic_id = NEW.id;
            UPDATE feed_activity fa SET is_nsfw = NEW.is_nsfw
              WHERE fa.type IN ('MESSAGE_UPVOTE', 'MESSAGE_SOLUTION') AND fa.link = '/topic/' || NEW.id;
        END IF;
    END IF;
    RETURN NEW;
END $function$;

CREATE OR REPLACE FUNCTION public.feed_sync_topic_reply() RETURNS trigger
LANGUAGE plpgsql AS $function$
DECLARE v_nsfw boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN PERFORM feed_delete('TOPIC_REPLY_CREATION', OLD.id); RETURN OLD; END IF;
    IF NEW.status <> 0 OR NOT topic_feed_visible(NEW.topic_id) THEN
        PERFORM feed_delete('TOPIC_REPLY_CREATION', NEW.id); RETURN NEW;
    END IF;
    SELECT is_nsfw INTO v_nsfw FROM topic WHERE id = NEW.topic_id;
    PERFORM feed_upsert('TOPIC_REPLY_CREATION', NEW.id, NEW.user_id, 0, COALESCE(NEW.content, ''), '/topic/' || NEW.topic_id, COALESCE(v_nsfw, false), NEW.created);
    RETURN NEW;
END $function$;

CREATE OR REPLACE FUNCTION public.feed_sync_topic_comment() RETURNS trigger
LANGUAGE plpgsql AS $function$
DECLARE v_nsfw boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN PERFORM feed_delete('TOPIC_COMMENT_CREATION', OLD.id); RETURN OLD; END IF;
    IF NEW.status <> 0 OR NOT topic_feed_visible(NEW.topic_id) THEN
        PERFORM feed_delete('TOPIC_COMMENT_CREATION', NEW.id); RETURN NEW;
    END IF;
    SELECT is_nsfw INTO v_nsfw FROM topic WHERE id = NEW.topic_id;
    PERFORM feed_upsert('TOPIC_COMMENT_CREATION', NEW.id, NEW.user_id, 0, SUBSTRING(NEW.content, 1, 100), '/topic/' || NEW.topic_id, COALESCE(v_nsfw, false), NEW.created);
    RETURN NEW;
END $function$;

CREATE OR REPLACE FUNCTION public.feed_sync_topic_upvote() RETURNS trigger
LANGUAGE plpgsql AS $function$
DECLARE v_nsfw boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN PERFORM feed_delete('TOPIC_UPVOTE', OLD.id); RETURN OLD; END IF;
    IF NOT topic_feed_visible(NEW.topic_id) THEN
        PERFORM feed_delete('TOPIC_UPVOTE', NEW.id); RETURN NEW;
    END IF;
    SELECT is_nsfw INTO v_nsfw FROM topic WHERE id = NEW.topic_id;
    PERFORM feed_upsert('TOPIC_UPVOTE', NEW.id, NEW.user_id, 0, COALESCE(NEW.description, ''), '/topic/' || NEW.topic_id, COALESCE(v_nsfw, false), NEW.created);
    RETURN NEW;
END $function$;

CREATE OR REPLACE FUNCTION public.feed_sync_message() RETURNS trigger
LANGUAGE plpgsql AS $function$
DECLARE v_type text; v_tid int; v_nsfw boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM feed_delete('MESSAGE_UPVOTE', OLD.id);
        PERFORM feed_delete('MESSAGE_SOLUTION', OLD.id);
        RETURN OLD;
    END IF;
    v_type := CASE NEW.type WHEN 'upvoted' THEN 'MESSAGE_UPVOTE' WHEN 'solution' THEN 'MESSAGE_SOLUTION' ELSE NULL END;
    IF v_type IS NOT NULL THEN
        v_tid := (regexp_match(NEW.link, '^/topic/([0-9]+)'))[1]::int;
        -- Non-topic links (v_tid NULL) keep the old unconditional path.
        IF v_tid IS NOT NULL AND NOT topic_feed_visible(v_tid) THEN
            PERFORM feed_delete(v_type, NEW.id); RETURN NEW;
        END IF;
        IF v_tid IS NOT NULL THEN SELECT is_nsfw INTO v_nsfw FROM topic WHERE id = v_tid; END IF;
        PERFORM feed_upsert(v_type, NEW.id, NEW.sender_id, 0, NEW.content, NEW.link, COALESCE(v_nsfw, false), NEW.created);
    END IF;
    RETURN NEW;
END $function$;

-- Sweep the rows the old child triggers leaked for already-hidden topics
-- (dev showed ~842 reply rows alone). Idempotent: deleting nothing is fine.
DELETE FROM feed_activity fa USING topic_reply r, topic t
  WHERE fa.type = 'TOPIC_REPLY_CREATION' AND fa.source_id = r.id
    AND r.topic_id = t.id AND (t.status <> 0 OR t.access_scope <> 'public');
DELETE FROM feed_activity fa USING topic_comment c, topic t
  WHERE fa.type = 'TOPIC_COMMENT_CREATION' AND fa.source_id = c.id
    AND c.topic_id = t.id AND (t.status <> 0 OR t.access_scope <> 'public');
DELETE FROM feed_activity fa USING topic_upvote u, topic t
  WHERE fa.type = 'TOPIC_UPVOTE' AND fa.source_id = u.id
    AND u.topic_id = t.id AND (t.status <> 0 OR t.access_scope <> 'public');
DELETE FROM feed_activity fa USING topic t
  WHERE fa.type IN ('MESSAGE_UPVOTE', 'MESSAGE_SOLUTION')
    AND (fa.link = '/topic/' || t.id OR fa.link LIKE '/topic/' || t.id || '?%')
    AND (t.status <> 0 OR t.access_scope <> 'public');
