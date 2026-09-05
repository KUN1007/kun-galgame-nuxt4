-- Restore the pre-090 trigger functions verbatim (children of hidden topics
-- go back to entering the feed — that was the old behavior, leak included),
-- then drop the scope column and grant table. Swept feed rows are not
-- resurrected; the next write to each child row re-inserts it.
CREATE OR REPLACE FUNCTION public.feed_sync_topic() RETURNS trigger
LANGUAGE plpgsql AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN PERFORM feed_delete('TOPIC_CREATION', OLD.id); RETURN OLD; END IF;
    IF NEW.status = 1 THEN PERFORM feed_delete('TOPIC_CREATION', NEW.id); RETURN NEW; END IF;
    PERFORM feed_upsert('TOPIC_CREATION', NEW.id, NEW.user_id, 0, NEW.title, '/topic/' || NEW.id, NEW.is_nsfw, NEW.created);
    IF TG_OP = 'UPDATE' AND NEW.is_nsfw IS DISTINCT FROM OLD.is_nsfw THEN
        UPDATE feed_activity fa SET is_nsfw = NEW.is_nsfw
          FROM topic_reply r   WHERE fa.type = 'TOPIC_REPLY_CREATION'   AND fa.source_id = r.id AND r.topic_id = NEW.id;
        UPDATE feed_activity fa SET is_nsfw = NEW.is_nsfw
          FROM topic_comment c WHERE fa.type = 'TOPIC_COMMENT_CREATION' AND fa.source_id = c.id AND c.topic_id = NEW.id;
        UPDATE feed_activity fa SET is_nsfw = NEW.is_nsfw
          FROM topic_upvote u  WHERE fa.type = 'TOPIC_UPVOTE'           AND fa.source_id = u.id AND u.topic_id = NEW.id;
        UPDATE feed_activity fa SET is_nsfw = NEW.is_nsfw
          WHERE fa.type IN ('MESSAGE_UPVOTE', 'MESSAGE_SOLUTION') AND fa.link = '/topic/' || NEW.id;
    END IF;
    RETURN NEW;
END $function$;

CREATE OR REPLACE FUNCTION public.feed_sync_topic_reply() RETURNS trigger
LANGUAGE plpgsql AS $function$
DECLARE v_nsfw boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN PERFORM feed_delete('TOPIC_REPLY_CREATION', OLD.id); RETURN OLD; END IF;
    IF NEW.status <> 0 THEN PERFORM feed_delete('TOPIC_REPLY_CREATION', NEW.id); RETURN NEW; END IF;
    SELECT is_nsfw INTO v_nsfw FROM topic WHERE id = NEW.topic_id;
    PERFORM feed_upsert('TOPIC_REPLY_CREATION', NEW.id, NEW.user_id, 0, COALESCE(NEW.content, ''), '/topic/' || NEW.topic_id, COALESCE(v_nsfw, false), NEW.created);
    RETURN NEW;
END $function$;

CREATE OR REPLACE FUNCTION public.feed_sync_topic_comment() RETURNS trigger
LANGUAGE plpgsql AS $function$
DECLARE v_nsfw boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN PERFORM feed_delete('TOPIC_COMMENT_CREATION', OLD.id); RETURN OLD; END IF;
    IF NEW.status <> 0 THEN PERFORM feed_delete('TOPIC_COMMENT_CREATION', NEW.id); RETURN NEW; END IF;
    SELECT is_nsfw INTO v_nsfw FROM topic WHERE id = NEW.topic_id;
    PERFORM feed_upsert('TOPIC_COMMENT_CREATION', NEW.id, NEW.user_id, 0, SUBSTRING(NEW.content, 1, 100), '/topic/' || NEW.topic_id, COALESCE(v_nsfw, false), NEW.created);
    RETURN NEW;
END $function$;

CREATE OR REPLACE FUNCTION public.feed_sync_topic_upvote() RETURNS trigger
LANGUAGE plpgsql AS $function$
DECLARE v_nsfw boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN PERFORM feed_delete('TOPIC_UPVOTE', OLD.id); RETURN OLD; END IF;
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
        IF v_tid IS NOT NULL THEN SELECT is_nsfw INTO v_nsfw FROM topic WHERE id = v_tid; END IF;
        PERFORM feed_upsert(v_type, NEW.id, NEW.sender_id, 0, NEW.content, NEW.link, COALESCE(v_nsfw, false), NEW.created);
    END IF;
    RETURN NEW;
END $function$;

DROP FUNCTION IF EXISTS topic_feed_visible(integer);
DROP TABLE IF EXISTS topic_access_grant;
DROP INDEX IF EXISTS idx_topic_access_scope_nonpublic;
ALTER TABLE topic DROP COLUMN IF EXISTS access_scope;
