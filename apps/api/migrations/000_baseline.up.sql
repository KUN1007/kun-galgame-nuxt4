-- 000_baseline.up.sql
--
-- Pre-001 schema baseline — the schema as it stood when the Go backend
-- took over from the legacy Nuxt + Nitro + Prisma stack. Captured by
-- `pg_restore -s` from nextmoe-infra/scripts/kungalgame_backup.dump
-- (the production snapshot used by the cross-project migration
-- runbook, see nextmoe-infra/scripts/reset_all.sh).
--
-- After this baseline applies, migrations 001-011 layer on the changes
-- Go has made since the cut-over (new tables, jsonb conversions, count
-- denorm columns, comment nesting, user identity columns dropped, …).
-- The end-state of `000_baseline + 001-011` equals the live database.
--
-- All statements are idempotent so this is safe to re-run:
--   • Every CREATE TABLE/SEQUENCE/INDEX uses `IF NOT EXISTS`
--   • Every ALTER TABLE ADD CONSTRAINT is wrapped in a `DO $$` block
--     that swallows `duplicate_object` / `duplicate_table` /
--     `invalid_table_definition` errors and emits a NOTICE instead.
--
-- DEPLOYMENT NOTE for existing databases:
--   The cross-project runbook restores kungalgame from the dump in
--   reset_all.sh — that dump IS the same content as this baseline, so
--   after restore there is nothing for the migrator to do here. Mark
--   the baseline as already applied so the runner doesn't replay it:
--
--     INSERT INTO _migrations (name) VALUES ('000_baseline')
--     ON CONFLICT (name) DO NOTHING;
--
-- For fresh databases (CI, DR rehearsal, new dev machine without a
-- dump):
--   The migrator picks this up first and creates the pre-001 schema.
--   001-011 then run normally — they were written assuming this is the
--   starting state.
--
-- Regeneration:
--   Re-extract from the same dump:
--     pg_restore -s --no-owner --no-acl --no-comments \
--       -f /tmp/raw.sql nextmoe-infra/scripts/kungalgame_backup.dump
--   Then run the sanitizer (Python script committed alongside this
--   migration in the original PR description).
--

--
--

--
-- Name: chat_message; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.chat_message (
    id integer NOT NULL,
    chatroom_name text NOT NULL,
    content character varying(1000) NOT NULL,
    is_recall boolean DEFAULT false NOT NULL,
    recall_time timestamp(3) without time zone,
    edit_time timestamp(3) without time zone,
    chat_room_id integer NOT NULL,
    sender_id integer NOT NULL,
    receiver_id integer,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: chat_message_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.chat_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: chat_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chat_message_id_seq OWNED BY public.chat_message.id;

--
-- Name: chat_message_reaction; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.chat_message_reaction (
    id integer NOT NULL,
    reaction text NOT NULL,
    chat_message_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: chat_message_reaction_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.chat_message_reaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: chat_message_reaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chat_message_reaction_id_seq OWNED BY public.chat_message_reaction.id;

--
-- Name: chat_message_read_by; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.chat_message_read_by (
    id integer NOT NULL,
    read_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    chat_message_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: chat_message_read_by_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.chat_message_read_by_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: chat_message_read_by_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chat_message_read_by_id_seq OWNED BY public.chat_message_read_by.id;

--
-- Name: chat_room; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.chat_room (
    id integer NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    avatar text DEFAULT ''::text NOT NULL,
    type text NOT NULL,
    last_message_content text DEFAULT ''::text NOT NULL,
    last_message_time timestamp(3) without time zone,
    last_message_sender_id integer,
    last_message_sender_name text DEFAULT ''::text NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: chat_room_admin; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.chat_room_admin (
    id integer NOT NULL,
    chat_room_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: chat_room_admin_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.chat_room_admin_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: chat_room_admin_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chat_room_admin_id_seq OWNED BY public.chat_room_admin.id;

--
-- Name: chat_room_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.chat_room_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: chat_room_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chat_room_id_seq OWNED BY public.chat_room.id;

--
-- Name: chat_room_participant; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.chat_room_participant (
    id integer NOT NULL,
    chat_room_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: chat_room_participant_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.chat_room_participant_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: chat_room_participant_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chat_room_participant_id_seq OWNED BY public.chat_room_participant.id;

--
-- Name: doc_article; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.doc_article (
    id integer NOT NULL,
    title character varying(233) NOT NULL,
    slug character varying(233) NOT NULL,
    path character varying(255) NOT NULL,
    description character varying(777) NOT NULL,
    banner character varying(777) DEFAULT ''::character varying NOT NULL,
    status integer DEFAULT 1 NOT NULL,
    is_pin boolean DEFAULT false NOT NULL,
    view integer DEFAULT 0 NOT NULL,
    published_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    edited_time timestamp(3) without time zone,
    content_markdown character varying(100007) NOT NULL,
    category_id integer NOT NULL,
    author_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: doc_article_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.doc_article_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: doc_article_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.doc_article_id_seq OWNED BY public.doc_article.id;

--
-- Name: doc_article_tag_relation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.doc_article_tag_relation (
    doc_article_id integer NOT NULL,
    doc_tag_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: doc_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.doc_category (
    id integer NOT NULL,
    slug character varying(128) NOT NULL,
    title character varying(233) NOT NULL,
    description character varying(777) DEFAULT ''::character varying NOT NULL,
    icon character varying(128) DEFAULT ''::character varying NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: doc_category_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.doc_category_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: doc_category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.doc_category_id_seq OWNED BY public.doc_category.id;

--
-- Name: doc_tag; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.doc_tag (
    id integer NOT NULL,
    slug character varying(128) NOT NULL,
    title character varying(128) NOT NULL,
    description character varying(255) DEFAULT ''::character varying NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: doc_tag_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.doc_tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: doc_tag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.doc_tag_id_seq OWNED BY public.doc_tag.id;

--
-- Name: galgame; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame (
    id integer NOT NULL,
    vndb_id character varying(10) NOT NULL,
    name_en_us character varying(1000) DEFAULT ''::character varying NOT NULL,
    name_ja_jp character varying(1000) DEFAULT ''::character varying NOT NULL,
    name_zh_cn character varying(1000) DEFAULT ''::character varying NOT NULL,
    name_zh_tw character varying(1000) DEFAULT ''::character varying NOT NULL,
    banner character varying(233) DEFAULT ''::character varying NOT NULL,
    intro_en_us text DEFAULT ''::text NOT NULL,
    intro_ja_jp text DEFAULT ''::text NOT NULL,
    intro_zh_cn text DEFAULT ''::text NOT NULL,
    intro_zh_tw text DEFAULT ''::text NOT NULL,
    content_limit character varying(10) DEFAULT 'sfw'::character varying NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    view integer DEFAULT 0 NOT NULL,
    resource_update_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    original_language text DEFAULT 'ja-jp'::text NOT NULL,
    age_limit text DEFAULT 'r18'::text NOT NULL,
    user_id integer NOT NULL,
    series_id integer,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_alias; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_alias (
    id integer NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    galgame_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_alias_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_alias_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_alias_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_alias_id_seq OWNED BY public.galgame_alias.id;

--
-- Name: galgame_comment; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_comment (
    id integer NOT NULL,
    content character varying(1007) NOT NULL,
    galgame_id integer NOT NULL,
    user_id integer NOT NULL,
    target_user_id integer,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_comment_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_comment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_comment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_comment_id_seq OWNED BY public.galgame_comment.id;

--
-- Name: galgame_comment_like; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_comment_like (
    id integer NOT NULL,
    user_id integer NOT NULL,
    galgame_comment_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_comment_like_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_comment_like_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_comment_like_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_comment_like_id_seq OWNED BY public.galgame_comment_like.id;

--
-- Name: galgame_contributor; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_contributor (
    id integer NOT NULL,
    galgame_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_contributor_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_contributor_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_contributor_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_contributor_id_seq OWNED BY public.galgame_contributor.id;

--
-- Name: galgame_engine; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_engine (
    id integer NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    alias text[] DEFAULT ARRAY[]::text[],
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_engine_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_engine_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_engine_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_engine_id_seq OWNED BY public.galgame_engine.id;

--
-- Name: galgame_engine_relation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_engine_relation (
    galgame_id integer NOT NULL,
    engine_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_favorite; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_favorite (
    id integer NOT NULL,
    galgame_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_favorite_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_favorite_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_favorite_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_favorite_id_seq OWNED BY public.galgame_favorite.id;

--
-- Name: galgame_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_history (
    id integer NOT NULL,
    action text DEFAULT ''::text NOT NULL,
    type text DEFAULT ''::text NOT NULL,
    content character varying(1007) DEFAULT ''::character varying NOT NULL,
    galgame_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_history_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_history_id_seq OWNED BY public.galgame_history.id;

--
-- Name: galgame_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_id_seq OWNED BY public.galgame.id;

--
-- Name: galgame_like; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_like (
    id integer NOT NULL,
    galgame_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_like_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_like_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_like_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_like_id_seq OWNED BY public.galgame_like.id;

--
-- Name: galgame_link; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_link (
    id integer NOT NULL,
    name character varying(107) DEFAULT ''::character varying NOT NULL,
    link character varying(233) DEFAULT ''::character varying NOT NULL,
    galgame_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_link_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_link_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_link_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_link_id_seq OWNED BY public.galgame_link.id;

--
-- Name: galgame_official; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_official (
    id integer NOT NULL,
    link text DEFAULT ''::text NOT NULL,
    name text NOT NULL,
    category text NOT NULL,
    lang text DEFAULT ''::text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_official_alias; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_official_alias (
    id integer NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    galgame_official_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_official_alias_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_official_alias_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_official_alias_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_official_alias_id_seq OWNED BY public.galgame_official_alias.id;

--
-- Name: galgame_official_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_official_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_official_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_official_id_seq OWNED BY public.galgame_official.id;

--
-- Name: galgame_official_relation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_official_relation (
    galgame_id integer NOT NULL,
    official_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_pr; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_pr (
    id integer NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    index integer DEFAULT 0 NOT NULL,
    note character varying(1007) DEFAULT ''::character varying NOT NULL,
    completed_time timestamp(3) without time zone,
    old_data jsonb,
    new_data jsonb,
    user_id integer NOT NULL,
    galgame_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_pr_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_pr_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_pr_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_pr_id_seq OWNED BY public.galgame_pr.id;

--
-- Name: galgame_rating; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_rating (
    id integer NOT NULL,
    recommend text NOT NULL,
    overall integer NOT NULL,
    view integer DEFAULT 0 NOT NULL,
    galgame_type text[] DEFAULT ARRAY[]::text[],
    play_status text DEFAULT 'not_started'::text NOT NULL,
    short_summary character varying(1314) DEFAULT ''::character varying NOT NULL,
    spoiler_level text DEFAULT 'none'::text NOT NULL,
    art integer DEFAULT 0 NOT NULL,
    story integer DEFAULT 0 NOT NULL,
    music integer DEFAULT 0 NOT NULL,
    "character" integer DEFAULT 0 NOT NULL,
    route integer DEFAULT 0 NOT NULL,
    system integer DEFAULT 0 NOT NULL,
    voice integer DEFAULT 0 NOT NULL,
    replay_value integer DEFAULT 0 NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL,
    user_id integer NOT NULL,
    galgame_id integer NOT NULL
);

--
-- Name: galgame_rating_comment; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_rating_comment (
    id integer NOT NULL,
    content character varying(1314) DEFAULT ''::character varying NOT NULL,
    galgame_rating_id integer NOT NULL,
    user_id integer NOT NULL,
    target_user_id integer,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_rating_comment_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_rating_comment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_rating_comment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_rating_comment_id_seq OWNED BY public.galgame_rating_comment.id;

--
-- Name: galgame_rating_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_rating_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_rating_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_rating_id_seq OWNED BY public.galgame_rating.id;

--
-- Name: galgame_rating_like; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_rating_like (
    id integer NOT NULL,
    galgame_rating_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_rating_like_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_rating_like_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_rating_like_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_rating_like_id_seq OWNED BY public.galgame_rating_like.id;

--
-- Name: galgame_resource; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_resource (
    id integer NOT NULL,
    type text DEFAULT ''::text NOT NULL,
    language text DEFAULT ''::text NOT NULL,
    platform text DEFAULT ''::text NOT NULL,
    size character varying(107) DEFAULT ''::character varying NOT NULL,
    code character varying(1007) DEFAULT ''::character varying NOT NULL,
    password character varying(1007) DEFAULT ''::character varying NOT NULL,
    note character varying(1007) DEFAULT ''::character varying NOT NULL,
    update_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP CONSTRAINT galgame_resource_update_time_not_null1 NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    galgame_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL,
    download integer DEFAULT 0 NOT NULL,
    provider text[] DEFAULT ARRAY[]::text[],
    edited timestamp(3) without time zone,
    view integer DEFAULT 0 NOT NULL
);

--
-- Name: galgame_resource_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_resource_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_resource_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_resource_id_seq OWNED BY public.galgame_resource.id;

--
-- Name: galgame_resource_like; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_resource_like (
    id integer NOT NULL,
    user_id integer NOT NULL,
    galgame_resource_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_resource_like_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_resource_like_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_resource_like_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_resource_like_id_seq OWNED BY public.galgame_resource_like.id;

--
-- Name: galgame_resource_link; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_resource_link (
    id integer NOT NULL,
    url text NOT NULL,
    galgame_resource_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_resource_link_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_resource_link_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_resource_link_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_resource_link_id_seq OWNED BY public.galgame_resource_link.id;

--
-- Name: galgame_series; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_series (
    id integer NOT NULL,
    name character varying(1000) DEFAULT ''::character varying NOT NULL,
    description character varying(2000) DEFAULT ''::character varying NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_series_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_series_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_series_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_series_id_seq OWNED BY public.galgame_series.id;

--
-- Name: galgame_tag; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_tag (
    id integer NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    category text NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_tag_alias; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_tag_alias (
    id integer NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    galgame_tag_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_tag_alias_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_tag_alias_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_tag_alias_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_tag_alias_id_seq OWNED BY public.galgame_tag_alias.id;

--
-- Name: galgame_tag_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_tag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_tag_id_seq OWNED BY public.galgame_tag.id;

--
-- Name: galgame_tag_relation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_tag_relation (
    galgame_id integer NOT NULL,
    tag_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL,
    spoiler_level integer DEFAULT 0 NOT NULL
);

--
-- Name: galgame_toolset; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_toolset (
    id integer NOT NULL,
    name character varying(500) DEFAULT ''::character varying NOT NULL,
    description character varying(2000) DEFAULT ''::character varying NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    view integer DEFAULT 0 NOT NULL,
    type text DEFAULT ''::text NOT NULL,
    language text DEFAULT ''::text NOT NULL,
    platform text DEFAULT ''::text NOT NULL,
    homepage text[] DEFAULT ARRAY[]::text[],
    resource_update_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    edited timestamp(3) without time zone,
    version character varying(233) DEFAULT ''::character varying NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_toolset_alias; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_toolset_alias (
    id integer NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    toolset_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_toolset_alias_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_toolset_alias_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_toolset_alias_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_toolset_alias_id_seq OWNED BY public.galgame_toolset_alias.id;

--
-- Name: galgame_toolset_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_toolset_category (
    id integer NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    alias text[] DEFAULT ARRAY[]::text[],
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_toolset_category_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_toolset_category_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_toolset_category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_toolset_category_id_seq OWNED BY public.galgame_toolset_category.id;

--
-- Name: galgame_toolset_category_relation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_toolset_category_relation (
    toolset_id integer NOT NULL,
    category_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_toolset_comment; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_toolset_comment (
    id integer NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    edited timestamp(3) without time zone,
    user_id integer NOT NULL,
    toolset_id integer NOT NULL,
    parent_id integer,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_toolset_comment_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_toolset_comment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_toolset_comment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_toolset_comment_id_seq OWNED BY public.galgame_toolset_comment.id;

--
-- Name: galgame_toolset_contributor; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_toolset_contributor (
    id integer NOT NULL,
    toolset_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_toolset_contributor_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_toolset_contributor_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_toolset_contributor_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_toolset_contributor_id_seq OWNED BY public.galgame_toolset_contributor.id;

--
-- Name: galgame_toolset_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_toolset_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_toolset_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_toolset_id_seq OWNED BY public.galgame_toolset.id;

--
-- Name: galgame_toolset_practicality; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_toolset_practicality (
    id integer NOT NULL,
    rate integer DEFAULT 1 NOT NULL,
    user_id integer NOT NULL,
    toolset_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_toolset_practicality_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_toolset_practicality_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_toolset_practicality_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_toolset_practicality_id_seq OWNED BY public.galgame_toolset_practicality.id;

--
-- Name: galgame_toolset_resource; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_toolset_resource (
    id integer NOT NULL,
    content character varying(1007) DEFAULT ''::character varying NOT NULL,
    type text DEFAULT ''::text NOT NULL,
    code character varying(1007) DEFAULT ''::character varying NOT NULL,
    password character varying(1007) DEFAULT ''::character varying NOT NULL,
    size character varying(107) DEFAULT ''::character varying NOT NULL,
    note character varying(1007) DEFAULT ''::character varying NOT NULL,
    download integer DEFAULT 0 NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    edited timestamp(3) without time zone,
    toolset_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_toolset_resource_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_toolset_resource_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_toolset_resource_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_toolset_resource_id_seq OWNED BY public.galgame_toolset_resource.id;

--
-- Name: galgame_website; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_website (
    id integer NOT NULL,
    name text NOT NULL,
    url text NOT NULL,
    create_time text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    icon text DEFAULT ''::text NOT NULL,
    view integer DEFAULT 0 NOT NULL,
    language text DEFAULT 'JA'::text NOT NULL,
    age_limit text DEFAULT 'all'::text NOT NULL,
    domain text[] DEFAULT ARRAY[]::text[],
    category_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL,
    user_id integer DEFAULT 2 NOT NULL
);

--
-- Name: galgame_website_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_website_category (
    id integer CONSTRAINT galgame_website_category_id_not_null1 NOT NULL,
    name text NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_website_category_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_website_category_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_website_category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_website_category_id_seq OWNED BY public.galgame_website_category.id;

--
-- Name: galgame_website_comment; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_website_comment (
    id integer NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    edited timestamp(3) without time zone,
    user_id integer NOT NULL,
    website_id integer NOT NULL,
    parent_id integer,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_website_comment_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_website_comment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_website_comment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_website_comment_id_seq OWNED BY public.galgame_website_comment.id;

--
-- Name: galgame_website_favorite; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_website_favorite (
    user_id integer NOT NULL,
    website_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_website_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_website_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_website_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_website_id_seq OWNED BY public.galgame_website.id;

--
-- Name: galgame_website_like; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_website_like (
    user_id integer NOT NULL,
    website_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_website_tag; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_website_tag (
    id integer NOT NULL,
    level integer NOT NULL,
    name text NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: galgame_website_tag_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.galgame_website_tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: galgame_website_tag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.galgame_website_tag_id_seq OWNED BY public.galgame_website_tag.id;

--
-- Name: galgame_website_tag_relation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.galgame_website_tag_relation (
    galgame_website_id integer NOT NULL,
    galgame_website_tag_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: message; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.message (
    id integer NOT NULL,
    content character varying(233) DEFAULT ''::character varying NOT NULL,
    link character varying(100) DEFAULT ''::character varying NOT NULL,
    status text DEFAULT 'unread'::text NOT NULL,
    type text NOT NULL,
    sender_id integer NOT NULL,
    receiver_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: message_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.message_id_seq OWNED BY public.message.id;

--
-- Name: report; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.report (
    id integer NOT NULL,
    reason character varying(1007) NOT NULL,
    type text DEFAULT ''::text NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: report_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.report_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: report_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.report_id_seq OWNED BY public.report.id;

--
-- Name: system_message; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.system_message (
    id integer NOT NULL,
    content_en_us text DEFAULT ''::text NOT NULL,
    content_ja_jp text DEFAULT ''::text NOT NULL,
    content_zh_cn text DEFAULT ''::text NOT NULL,
    content_zh_tw text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'unread'::text NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: system_message_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.system_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: system_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_message_id_seq OWNED BY public.system_message.id;

--
-- Name: todo; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.todo (
    id integer NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    content_en_us text DEFAULT ''::text NOT NULL,
    content_ja_jp text DEFAULT ''::text NOT NULL,
    content_zh_cn text DEFAULT ''::text NOT NULL,
    content_zh_tw text DEFAULT ''::text NOT NULL,
    completed_time timestamp(3) without time zone,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL,
    type text DEFAULT 'forum'::text NOT NULL
);

--
-- Name: todo_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.todo_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: todo_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.todo_id_seq OWNED BY public.todo.id;

--
-- Name: topic; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic (
    id integer NOT NULL,
    title character varying(233) NOT NULL,
    content text NOT NULL,
    view integer DEFAULT 0 NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    category text NOT NULL,
    tag text[],
    status_update_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    edited timestamp(3) without time zone,
    upvote_time timestamp(3) without time zone,
    user_id integer NOT NULL,
    best_answer_id integer,
    pinned_reply_id integer,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL,
    is_nsfw boolean DEFAULT false NOT NULL
);

--
-- Name: topic_comment; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_comment (
    id integer NOT NULL,
    content character varying(1007) DEFAULT ''::character varying NOT NULL,
    topic_id integer NOT NULL,
    topic_reply_id integer NOT NULL,
    user_id integer NOT NULL,
    target_user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_comment_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_comment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_comment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_comment_id_seq OWNED BY public.topic_comment.id;

--
-- Name: topic_comment_like; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_comment_like (
    id integer NOT NULL,
    topic_comment_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_comment_like_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_comment_like_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_comment_like_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_comment_like_id_seq OWNED BY public.topic_comment_like.id;

--
-- Name: topic_dislike; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_dislike (
    id integer NOT NULL,
    topic_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_dislike_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_dislike_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_dislike_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_dislike_id_seq OWNED BY public.topic_dislike.id;

--
-- Name: topic_favorite; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_favorite (
    id integer NOT NULL,
    topic_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_favorite_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_favorite_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_favorite_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_favorite_id_seq OWNED BY public.topic_favorite.id;

--
-- Name: topic_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_id_seq OWNED BY public.topic.id;

--
-- Name: topic_like; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_like (
    id integer NOT NULL,
    topic_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_like_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_like_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_like_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_like_id_seq OWNED BY public.topic_like.id;

--
-- Name: topic_poll; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_poll (
    id integer NOT NULL,
    title character varying(100) NOT NULL,
    description character varying(500) DEFAULT ''::character varying NOT NULL,
    type text DEFAULT 'single'::text NOT NULL,
    min_choice integer DEFAULT 1 NOT NULL,
    max_choice integer DEFAULT 1 NOT NULL,
    deadline timestamp(3) without time zone,
    status text DEFAULT 'open'::text NOT NULL,
    notification_sent boolean DEFAULT false NOT NULL,
    result_visibility text DEFAULT 'always'::text NOT NULL,
    is_anonymous boolean DEFAULT false NOT NULL,
    can_change_vote boolean DEFAULT true NOT NULL,
    topic_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_poll_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_poll_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_poll_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_poll_id_seq OWNED BY public.topic_poll.id;

--
-- Name: topic_poll_option; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_poll_option (
    id integer NOT NULL,
    text character varying(100) NOT NULL,
    poll_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_poll_option_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_poll_option_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_poll_option_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_poll_option_id_seq OWNED BY public.topic_poll_option.id;

--
-- Name: topic_poll_vote; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_poll_vote (
    id integer NOT NULL,
    poll_id integer NOT NULL,
    option_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_poll_vote_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_poll_vote_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_poll_vote_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_poll_vote_id_seq OWNED BY public.topic_poll_vote.id;

--
-- Name: topic_reply; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_reply (
    id integer NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    floor integer DEFAULT 0 NOT NULL,
    edited timestamp(3) without time zone,
    user_id integer NOT NULL,
    topic_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_reply_dislike; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_reply_dislike (
    id integer NOT NULL,
    user_id integer NOT NULL,
    topic_reply_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_reply_dislike_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_reply_dislike_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_reply_dislike_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_reply_dislike_id_seq OWNED BY public.topic_reply_dislike.id;

--
-- Name: topic_reply_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_reply_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_reply_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_reply_id_seq OWNED BY public.topic_reply.id;

--
-- Name: topic_reply_like; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_reply_like (
    id integer NOT NULL,
    user_id integer NOT NULL,
    topic_reply_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_reply_like_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_reply_like_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_reply_like_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_reply_like_id_seq OWNED BY public.topic_reply_like.id;

--
-- Name: topic_reply_target; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_reply_target (
    id integer NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    reply_id integer NOT NULL,
    target_reply_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_reply_target_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_reply_target_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_reply_target_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_reply_target_id_seq OWNED BY public.topic_reply_target.id;

--
-- Name: topic_section; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_section (
    id integer NOT NULL,
    name text NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_section_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_section_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_section_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_section_id_seq OWNED BY public.topic_section.id;

--
-- Name: topic_section_relation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_section_relation (
    topic_id integer NOT NULL,
    topic_section_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_upvote; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.topic_upvote (
    id integer NOT NULL,
    topic_id integer NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: topic_upvote_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.topic_upvote_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: topic_upvote_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.topic_upvote_id_seq OWNED BY public.topic_upvote.id;

--
-- Name: unmoe; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.unmoe (
    id integer NOT NULL,
    name text NOT NULL,
    result text DEFAULT ''::text NOT NULL,
    desc_en_us text DEFAULT ''::text NOT NULL,
    desc_ja_jp text DEFAULT ''::text NOT NULL,
    desc_zh_cn text DEFAULT ''::text NOT NULL,
    desc_zh_tw text DEFAULT ''::text NOT NULL,
    user_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: unmoe_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.unmoe_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: unmoe_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.unmoe_id_seq OWNED BY public.unmoe.id;

--
-- Name: update_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.update_log (
    id integer NOT NULL,
    type text NOT NULL,
    version text DEFAULT ''::text NOT NULL,
    content_en_us text DEFAULT ''::text NOT NULL,
    content_ja_jp text DEFAULT ''::text NOT NULL,
    content_zh_cn text DEFAULT ''::text NOT NULL,
    content_zh_tw text DEFAULT ''::text NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL,
    user_id integer DEFAULT 2 NOT NULL
);

--
-- Name: update_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.update_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: update_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.update_log_id_seq OWNED BY public.update_log.id;

--
-- Name: user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public."user" (
    id integer NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    ip text DEFAULT ''::text NOT NULL,
    avatar text DEFAULT ''::text NOT NULL,
    role integer DEFAULT 1 NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    moemoepoint integer DEFAULT 7 NOT NULL,
    bio character varying(107) DEFAULT ''::character varying NOT NULL,
    daily_check_in integer DEFAULT 0 NOT NULL,
    daily_image_count integer DEFAULT 0 NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL,
    daily_toolset_upload_count integer DEFAULT 0 NOT NULL
);

--
-- Name: user_follow; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.user_follow (
    id integer NOT NULL,
    follower_id integer NOT NULL,
    followed_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: user_follow_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.user_follow_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: user_follow_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_follow_id_seq OWNED BY public.user_follow.id;

--
-- Name: user_friend; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.user_friend (
    id integer NOT NULL,
    user_id integer NOT NULL,
    friend_id integer NOT NULL,
    created timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp(3) without time zone NOT NULL
);

--
-- Name: user_friend_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.user_friend_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: user_friend_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_friend_id_seq OWNED BY public.user_friend.id;

--
-- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- Name: user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_id_seq OWNED BY public."user".id;

--
-- Name: chat_message id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_message ALTER COLUMN id SET DEFAULT nextval('public.chat_message_id_seq'::regclass);

--
-- Name: chat_message_reaction id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_message_reaction ALTER COLUMN id SET DEFAULT nextval('public.chat_message_reaction_id_seq'::regclass);

--
-- Name: chat_message_read_by id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_message_read_by ALTER COLUMN id SET DEFAULT nextval('public.chat_message_read_by_id_seq'::regclass);

--
-- Name: chat_room id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_room ALTER COLUMN id SET DEFAULT nextval('public.chat_room_id_seq'::regclass);

--
-- Name: chat_room_admin id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_room_admin ALTER COLUMN id SET DEFAULT nextval('public.chat_room_admin_id_seq'::regclass);

--
-- Name: chat_room_participant id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_room_participant ALTER COLUMN id SET DEFAULT nextval('public.chat_room_participant_id_seq'::regclass);

--
-- Name: doc_article id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_article ALTER COLUMN id SET DEFAULT nextval('public.doc_article_id_seq'::regclass);

--
-- Name: doc_category id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_category ALTER COLUMN id SET DEFAULT nextval('public.doc_category_id_seq'::regclass);

--
-- Name: doc_tag id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_tag ALTER COLUMN id SET DEFAULT nextval('public.doc_tag_id_seq'::regclass);

--
-- Name: galgame id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame ALTER COLUMN id SET DEFAULT nextval('public.galgame_id_seq'::regclass);

--
-- Name: galgame_alias id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_alias ALTER COLUMN id SET DEFAULT nextval('public.galgame_alias_id_seq'::regclass);

--
-- Name: galgame_comment id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_comment ALTER COLUMN id SET DEFAULT nextval('public.galgame_comment_id_seq'::regclass);

--
-- Name: galgame_comment_like id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_comment_like ALTER COLUMN id SET DEFAULT nextval('public.galgame_comment_like_id_seq'::regclass);

--
-- Name: galgame_contributor id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_contributor ALTER COLUMN id SET DEFAULT nextval('public.galgame_contributor_id_seq'::regclass);

--
-- Name: galgame_engine id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_engine ALTER COLUMN id SET DEFAULT nextval('public.galgame_engine_id_seq'::regclass);

--
-- Name: galgame_favorite id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_favorite ALTER COLUMN id SET DEFAULT nextval('public.galgame_favorite_id_seq'::regclass);

--
-- Name: galgame_history id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_history ALTER COLUMN id SET DEFAULT nextval('public.galgame_history_id_seq'::regclass);

--
-- Name: galgame_like id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_like ALTER COLUMN id SET DEFAULT nextval('public.galgame_like_id_seq'::regclass);

--
-- Name: galgame_link id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_link ALTER COLUMN id SET DEFAULT nextval('public.galgame_link_id_seq'::regclass);

--
-- Name: galgame_official id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_official ALTER COLUMN id SET DEFAULT nextval('public.galgame_official_id_seq'::regclass);

--
-- Name: galgame_official_alias id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_official_alias ALTER COLUMN id SET DEFAULT nextval('public.galgame_official_alias_id_seq'::regclass);

--
-- Name: galgame_pr id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_pr ALTER COLUMN id SET DEFAULT nextval('public.galgame_pr_id_seq'::regclass);

--
-- Name: galgame_rating id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_rating ALTER COLUMN id SET DEFAULT nextval('public.galgame_rating_id_seq'::regclass);

--
-- Name: galgame_rating_comment id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_rating_comment ALTER COLUMN id SET DEFAULT nextval('public.galgame_rating_comment_id_seq'::regclass);

--
-- Name: galgame_rating_like id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_rating_like ALTER COLUMN id SET DEFAULT nextval('public.galgame_rating_like_id_seq'::regclass);

--
-- Name: galgame_resource id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_resource ALTER COLUMN id SET DEFAULT nextval('public.galgame_resource_id_seq'::regclass);

--
-- Name: galgame_resource_like id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_resource_like ALTER COLUMN id SET DEFAULT nextval('public.galgame_resource_like_id_seq'::regclass);

--
-- Name: galgame_resource_link id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_resource_link ALTER COLUMN id SET DEFAULT nextval('public.galgame_resource_link_id_seq'::regclass);

--
-- Name: galgame_series id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_series ALTER COLUMN id SET DEFAULT nextval('public.galgame_series_id_seq'::regclass);

--
-- Name: galgame_tag id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_tag ALTER COLUMN id SET DEFAULT nextval('public.galgame_tag_id_seq'::regclass);

--
-- Name: galgame_tag_alias id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_tag_alias ALTER COLUMN id SET DEFAULT nextval('public.galgame_tag_alias_id_seq'::regclass);

--
-- Name: galgame_toolset id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_toolset ALTER COLUMN id SET DEFAULT nextval('public.galgame_toolset_id_seq'::regclass);

--
-- Name: galgame_toolset_alias id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_toolset_alias ALTER COLUMN id SET DEFAULT nextval('public.galgame_toolset_alias_id_seq'::regclass);

--
-- Name: galgame_toolset_category id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_toolset_category ALTER COLUMN id SET DEFAULT nextval('public.galgame_toolset_category_id_seq'::regclass);

--
-- Name: galgame_toolset_comment id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_toolset_comment ALTER COLUMN id SET DEFAULT nextval('public.galgame_toolset_comment_id_seq'::regclass);

--
-- Name: galgame_toolset_contributor id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_toolset_contributor ALTER COLUMN id SET DEFAULT nextval('public.galgame_toolset_contributor_id_seq'::regclass);

--
-- Name: galgame_toolset_practicality id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_toolset_practicality ALTER COLUMN id SET DEFAULT nextval('public.galgame_toolset_practicality_id_seq'::regclass);

--
-- Name: galgame_toolset_resource id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_toolset_resource ALTER COLUMN id SET DEFAULT nextval('public.galgame_toolset_resource_id_seq'::regclass);

--
-- Name: galgame_website id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_website ALTER COLUMN id SET DEFAULT nextval('public.galgame_website_id_seq'::regclass);

--
-- Name: galgame_website_category id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_website_category ALTER COLUMN id SET DEFAULT nextval('public.galgame_website_category_id_seq'::regclass);

--
-- Name: galgame_website_comment id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_website_comment ALTER COLUMN id SET DEFAULT nextval('public.galgame_website_comment_id_seq'::regclass);

--
-- Name: galgame_website_tag id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.galgame_website_tag ALTER COLUMN id SET DEFAULT nextval('public.galgame_website_tag_id_seq'::regclass);

--
-- Name: message id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message ALTER COLUMN id SET DEFAULT nextval('public.message_id_seq'::regclass);

--
-- Name: report id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report ALTER COLUMN id SET DEFAULT nextval('public.report_id_seq'::regclass);

--
-- Name: system_message id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_message ALTER COLUMN id SET DEFAULT nextval('public.system_message_id_seq'::regclass);

--
-- Name: todo id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.todo ALTER COLUMN id SET DEFAULT nextval('public.todo_id_seq'::regclass);

--
-- Name: topic id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic ALTER COLUMN id SET DEFAULT nextval('public.topic_id_seq'::regclass);

--
-- Name: topic_comment id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_comment ALTER COLUMN id SET DEFAULT nextval('public.topic_comment_id_seq'::regclass);

--
-- Name: topic_comment_like id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_comment_like ALTER COLUMN id SET DEFAULT nextval('public.topic_comment_like_id_seq'::regclass);

--
-- Name: topic_dislike id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_dislike ALTER COLUMN id SET DEFAULT nextval('public.topic_dislike_id_seq'::regclass);

--
-- Name: topic_favorite id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_favorite ALTER COLUMN id SET DEFAULT nextval('public.topic_favorite_id_seq'::regclass);

--
-- Name: topic_like id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_like ALTER COLUMN id SET DEFAULT nextval('public.topic_like_id_seq'::regclass);

--
-- Name: topic_poll id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_poll ALTER COLUMN id SET DEFAULT nextval('public.topic_poll_id_seq'::regclass);

--
-- Name: topic_poll_option id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_poll_option ALTER COLUMN id SET DEFAULT nextval('public.topic_poll_option_id_seq'::regclass);

--
-- Name: topic_poll_vote id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_poll_vote ALTER COLUMN id SET DEFAULT nextval('public.topic_poll_vote_id_seq'::regclass);

--
-- Name: topic_reply id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_reply ALTER COLUMN id SET DEFAULT nextval('public.topic_reply_id_seq'::regclass);

--
-- Name: topic_reply_dislike id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_reply_dislike ALTER COLUMN id SET DEFAULT nextval('public.topic_reply_dislike_id_seq'::regclass);

--
-- Name: topic_reply_like id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_reply_like ALTER COLUMN id SET DEFAULT nextval('public.topic_reply_like_id_seq'::regclass);

--
-- Name: topic_reply_target id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_reply_target ALTER COLUMN id SET DEFAULT nextval('public.topic_reply_target_id_seq'::regclass);

--
-- Name: topic_section id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_section ALTER COLUMN id SET DEFAULT nextval('public.topic_section_id_seq'::regclass);

--
-- Name: topic_upvote id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.topic_upvote ALTER COLUMN id SET DEFAULT nextval('public.topic_upvote_id_seq'::regclass);

--
-- Name: unmoe id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.unmoe ALTER COLUMN id SET DEFAULT nextval('public.unmoe_id_seq'::regclass);

--
-- Name: update_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.update_log ALTER COLUMN id SET DEFAULT nextval('public.update_log_id_seq'::regclass);

--
-- Name: user id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."user" ALTER COLUMN id SET DEFAULT nextval('public.user_id_seq'::regclass);

--
-- Name: user_follow id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_follow ALTER COLUMN id SET DEFAULT nextval('public.user_follow_id_seq'::regclass);

--
-- Name: user_friend id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_friend ALTER COLUMN id SET DEFAULT nextval('public.user_friend_id_seq'::regclass);

--
-- Name: chat_message chat_message_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message
        ADD CONSTRAINT chat_message_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_message_reaction chat_message_reaction_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message_reaction
        ADD CONSTRAINT chat_message_reaction_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_message_read_by chat_message_read_by_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message_read_by
        ADD CONSTRAINT chat_message_read_by_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_room_admin chat_room_admin_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_room_admin
        ADD CONSTRAINT chat_room_admin_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_room_participant chat_room_participant_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_room_participant
        ADD CONSTRAINT chat_room_participant_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_room chat_room_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_room
        ADD CONSTRAINT chat_room_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: doc_article doc_article_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.doc_article
        ADD CONSTRAINT doc_article_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: doc_article_tag_relation doc_article_tag_relation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.doc_article_tag_relation
        ADD CONSTRAINT doc_article_tag_relation_pkey PRIMARY KEY (doc_article_id, doc_tag_id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: doc_category doc_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.doc_category
        ADD CONSTRAINT doc_category_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: doc_tag doc_tag_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.doc_tag
        ADD CONSTRAINT doc_tag_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_alias galgame_alias_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_alias
        ADD CONSTRAINT galgame_alias_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_comment_like galgame_comment_like_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_comment_like
        ADD CONSTRAINT galgame_comment_like_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_comment galgame_comment_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_comment
        ADD CONSTRAINT galgame_comment_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_contributor galgame_contributor_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_contributor
        ADD CONSTRAINT galgame_contributor_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_engine galgame_engine_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_engine
        ADD CONSTRAINT galgame_engine_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_engine_relation galgame_engine_relation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_engine_relation
        ADD CONSTRAINT galgame_engine_relation_pkey PRIMARY KEY (galgame_id, engine_id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_favorite galgame_favorite_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_favorite
        ADD CONSTRAINT galgame_favorite_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_history galgame_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_history
        ADD CONSTRAINT galgame_history_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_like galgame_like_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_like
        ADD CONSTRAINT galgame_like_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_link galgame_link_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_link
        ADD CONSTRAINT galgame_link_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_official_alias galgame_official_alias_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_official_alias
        ADD CONSTRAINT galgame_official_alias_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_official galgame_official_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_official
        ADD CONSTRAINT galgame_official_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_official_relation galgame_official_relation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_official_relation
        ADD CONSTRAINT galgame_official_relation_pkey PRIMARY KEY (galgame_id, official_id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame galgame_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame
        ADD CONSTRAINT galgame_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_pr galgame_pr_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_pr
        ADD CONSTRAINT galgame_pr_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating_comment galgame_rating_comment_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating_comment
        ADD CONSTRAINT galgame_rating_comment_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating_like galgame_rating_like_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating_like
        ADD CONSTRAINT galgame_rating_like_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating galgame_rating_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating
        ADD CONSTRAINT galgame_rating_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_resource_like galgame_resource_like_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_resource_like
        ADD CONSTRAINT galgame_resource_like_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_resource_link galgame_resource_link_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_resource_link
        ADD CONSTRAINT galgame_resource_link_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_resource galgame_resource_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_resource
        ADD CONSTRAINT galgame_resource_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_series galgame_series_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_series
        ADD CONSTRAINT galgame_series_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_tag_alias galgame_tag_alias_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_tag_alias
        ADD CONSTRAINT galgame_tag_alias_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_tag galgame_tag_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_tag
        ADD CONSTRAINT galgame_tag_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_tag_relation galgame_tag_relation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_tag_relation
        ADD CONSTRAINT galgame_tag_relation_pkey PRIMARY KEY (galgame_id, tag_id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_alias galgame_toolset_alias_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_alias
        ADD CONSTRAINT galgame_toolset_alias_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_category galgame_toolset_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_category
        ADD CONSTRAINT galgame_toolset_category_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_category_relation galgame_toolset_category_relation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_category_relation
        ADD CONSTRAINT galgame_toolset_category_relation_pkey PRIMARY KEY (toolset_id, category_id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_comment galgame_toolset_comment_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_comment
        ADD CONSTRAINT galgame_toolset_comment_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_contributor galgame_toolset_contributor_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_contributor
        ADD CONSTRAINT galgame_toolset_contributor_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset galgame_toolset_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset
        ADD CONSTRAINT galgame_toolset_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_practicality galgame_toolset_practicality_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_practicality
        ADD CONSTRAINT galgame_toolset_practicality_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_resource galgame_toolset_resource_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_resource
        ADD CONSTRAINT galgame_toolset_resource_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_category galgame_website_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_category
        ADD CONSTRAINT galgame_website_category_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_comment galgame_website_comment_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_comment
        ADD CONSTRAINT galgame_website_comment_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_favorite galgame_website_favorite_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_favorite
        ADD CONSTRAINT galgame_website_favorite_pkey PRIMARY KEY (user_id, website_id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_like galgame_website_like_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_like
        ADD CONSTRAINT galgame_website_like_pkey PRIMARY KEY (user_id, website_id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website galgame_website_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website
        ADD CONSTRAINT galgame_website_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_tag galgame_website_tag_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_tag
        ADD CONSTRAINT galgame_website_tag_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_tag_relation galgame_website_tag_relation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_tag_relation
        ADD CONSTRAINT galgame_website_tag_relation_pkey PRIMARY KEY (galgame_website_id, galgame_website_tag_id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: message message_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.message
        ADD CONSTRAINT message_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: report report_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.report
        ADD CONSTRAINT report_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: system_message system_message_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.system_message
        ADD CONSTRAINT system_message_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: todo todo_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.todo
        ADD CONSTRAINT todo_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_comment_like topic_comment_like_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_comment_like
        ADD CONSTRAINT topic_comment_like_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_comment topic_comment_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_comment
        ADD CONSTRAINT topic_comment_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_dislike topic_dislike_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_dislike
        ADD CONSTRAINT topic_dislike_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_favorite topic_favorite_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_favorite
        ADD CONSTRAINT topic_favorite_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_like topic_like_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_like
        ADD CONSTRAINT topic_like_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic topic_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic
        ADD CONSTRAINT topic_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_poll_option topic_poll_option_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_poll_option
        ADD CONSTRAINT topic_poll_option_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_poll topic_poll_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_poll
        ADD CONSTRAINT topic_poll_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_poll_vote topic_poll_vote_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_poll_vote
        ADD CONSTRAINT topic_poll_vote_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply_dislike topic_reply_dislike_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply_dislike
        ADD CONSTRAINT topic_reply_dislike_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply_like topic_reply_like_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply_like
        ADD CONSTRAINT topic_reply_like_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply topic_reply_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply
        ADD CONSTRAINT topic_reply_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply_target topic_reply_target_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply_target
        ADD CONSTRAINT topic_reply_target_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_section topic_section_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_section
        ADD CONSTRAINT topic_section_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_section_relation topic_section_relation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_section_relation
        ADD CONSTRAINT topic_section_relation_pkey PRIMARY KEY (topic_id, topic_section_id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_upvote topic_upvote_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_upvote
        ADD CONSTRAINT topic_upvote_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: unmoe unmoe_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.unmoe
        ADD CONSTRAINT unmoe_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: update_log update_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.update_log
        ADD CONSTRAINT update_log_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: user_follow user_follow_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.user_follow
        ADD CONSTRAINT user_follow_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: user_friend user_friend_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.user_friend
        ADD CONSTRAINT user_friend_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public."user"
        ADD CONSTRAINT user_pkey PRIMARY KEY (id);
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_message_reaction_chat_message_id_user_id_reaction_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS chat_message_reaction_chat_message_id_user_id_reaction_key ON public.chat_message_reaction USING btree (chat_message_id, user_id, reaction);

--
-- Name: chat_message_read_by_chat_message_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS chat_message_read_by_chat_message_id_user_id_key ON public.chat_message_read_by USING btree (chat_message_id, user_id);

--
-- Name: chat_room_admin_chat_room_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS chat_room_admin_chat_room_id_user_id_key ON public.chat_room_admin USING btree (chat_room_id, user_id);

--
-- Name: chat_room_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS chat_room_name_key ON public.chat_room USING btree (name);

--
-- Name: chat_room_participant_chat_room_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS chat_room_participant_chat_room_id_user_id_key ON public.chat_room_participant USING btree (chat_room_id, user_id);

--
-- Name: doc_article_path_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS doc_article_path_key ON public.doc_article USING btree (path);

--
-- Name: doc_article_slug_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS doc_article_slug_key ON public.doc_article USING btree (slug);

--
-- Name: doc_category_slug_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS doc_category_slug_key ON public.doc_category USING btree (slug);

--
-- Name: doc_tag_slug_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS doc_tag_slug_key ON public.doc_tag USING btree (slug);

--
-- Name: galgame_alias_galgame_id_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_alias_galgame_id_name_key ON public.galgame_alias USING btree (galgame_id, name);

--
-- Name: galgame_comment_like_galgame_comment_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_comment_like_galgame_comment_id_user_id_key ON public.galgame_comment_like USING btree (galgame_comment_id, user_id);

--
-- Name: galgame_contributor_galgame_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_contributor_galgame_id_user_id_key ON public.galgame_contributor USING btree (galgame_id, user_id);

--
-- Name: galgame_engine_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_engine_name_key ON public.galgame_engine USING btree (name);

--
-- Name: galgame_favorite_galgame_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_favorite_galgame_id_user_id_key ON public.galgame_favorite USING btree (galgame_id, user_id);

--
-- Name: galgame_like_galgame_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_like_galgame_id_user_id_key ON public.galgame_like USING btree (galgame_id, user_id);

--
-- Name: galgame_official_alias_galgame_official_id_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_official_alias_galgame_official_id_name_key ON public.galgame_official_alias USING btree (galgame_official_id, name);

--
-- Name: galgame_official_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_official_name_key ON public.galgame_official USING btree (name);

--
-- Name: galgame_rating_galgame_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS galgame_rating_galgame_id_idx ON public.galgame_rating USING btree (galgame_id);

--
-- Name: galgame_rating_like_galgame_rating_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_rating_like_galgame_rating_id_user_id_key ON public.galgame_rating_like USING btree (galgame_rating_id, user_id);

--
-- Name: galgame_rating_overall_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS galgame_rating_overall_idx ON public.galgame_rating USING btree (overall);

--
-- Name: galgame_rating_user_id_galgame_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_rating_user_id_galgame_id_key ON public.galgame_rating USING btree (user_id, galgame_id);

--
-- Name: galgame_resource_like_galgame_resource_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_resource_like_galgame_resource_id_user_id_key ON public.galgame_resource_like USING btree (galgame_resource_id, user_id);

--
-- Name: galgame_resource_link_galgame_resource_id_url_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_resource_link_galgame_resource_id_url_key ON public.galgame_resource_link USING btree (galgame_resource_id, url);

--
-- Name: galgame_series_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_series_name_key ON public.galgame_series USING btree (name);

--
-- Name: galgame_tag_alias_galgame_tag_id_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_tag_alias_galgame_tag_id_name_key ON public.galgame_tag_alias USING btree (galgame_tag_id, name);

--
-- Name: galgame_tag_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_tag_name_key ON public.galgame_tag USING btree (name);

--
-- Name: galgame_toolset_alias_toolset_id_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_toolset_alias_toolset_id_name_key ON public.galgame_toolset_alias USING btree (toolset_id, name);

--
-- Name: galgame_toolset_category_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_toolset_category_name_key ON public.galgame_toolset_category USING btree (name);

--
-- Name: galgame_toolset_comment_toolset_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS galgame_toolset_comment_toolset_id_idx ON public.galgame_toolset_comment USING btree (toolset_id);

--
-- Name: galgame_toolset_comment_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS galgame_toolset_comment_user_id_idx ON public.galgame_toolset_comment USING btree (user_id);

--
-- Name: galgame_toolset_contributor_toolset_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_toolset_contributor_toolset_id_user_id_key ON public.galgame_toolset_contributor USING btree (toolset_id, user_id);

--
-- Name: galgame_toolset_resource_toolset_id_content_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_toolset_resource_toolset_id_content_key ON public.galgame_toolset_resource USING btree (toolset_id, content);

--
-- Name: galgame_vndb_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_vndb_id_key ON public.galgame USING btree (vndb_id);

--
-- Name: galgame_website_category_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_website_category_name_key ON public.galgame_website_category USING btree (name);

--
-- Name: galgame_website_comment_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS galgame_website_comment_user_id_idx ON public.galgame_website_comment USING btree (user_id);

--
-- Name: galgame_website_comment_website_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS galgame_website_comment_website_id_idx ON public.galgame_website_comment USING btree (website_id);

--
-- Name: galgame_website_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_website_name_key ON public.galgame_website USING btree (name);

--
-- Name: galgame_website_tag_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_website_tag_name_key ON public.galgame_website_tag USING btree (name);

--
-- Name: galgame_website_url_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS galgame_website_url_key ON public.galgame_website USING btree (url);

--
-- Name: topic_best_answer_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_best_answer_id_key ON public.topic USING btree (best_answer_id);

--
-- Name: topic_comment_like_topic_comment_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_comment_like_topic_comment_id_user_id_key ON public.topic_comment_like USING btree (topic_comment_id, user_id);

--
-- Name: topic_dislike_topic_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_dislike_topic_id_user_id_key ON public.topic_dislike USING btree (topic_id, user_id);

--
-- Name: topic_favorite_topic_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_favorite_topic_id_user_id_key ON public.topic_favorite USING btree (topic_id, user_id);

--
-- Name: topic_like_topic_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_like_topic_id_user_id_key ON public.topic_like USING btree (topic_id, user_id);

--
-- Name: topic_pinned_reply_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_pinned_reply_id_key ON public.topic USING btree (pinned_reply_id);

--
-- Name: topic_poll_vote_poll_id_option_id_user_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_poll_vote_poll_id_option_id_user_id_key ON public.topic_poll_vote USING btree (poll_id, option_id, user_id);

--
-- Name: topic_poll_vote_user_id_poll_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS topic_poll_vote_user_id_poll_id_idx ON public.topic_poll_vote USING btree (user_id, poll_id);

--
-- Name: topic_reply_dislike_user_id_topic_reply_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_reply_dislike_user_id_topic_reply_id_key ON public.topic_reply_dislike USING btree (user_id, topic_reply_id);

--
-- Name: topic_reply_like_user_id_topic_reply_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_reply_like_user_id_topic_reply_id_key ON public.topic_reply_like USING btree (user_id, topic_reply_id);

--
-- Name: topic_reply_target_reply_id_target_reply_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS topic_reply_target_reply_id_target_reply_id_key ON public.topic_reply_target USING btree (reply_id, target_reply_id);

--
-- Name: user_email_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS user_email_key ON public."user" USING btree (email);

--
-- Name: user_follow_follower_id_followed_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS user_follow_follower_id_followed_id_key ON public.user_follow USING btree (follower_id, followed_id);

--
-- Name: user_friend_user_id_friend_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS user_friend_user_id_friend_id_key ON public.user_friend USING btree (user_id, friend_id);

--
-- Name: user_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS user_name_key ON public."user" USING btree (name);

--
-- Name: chat_message chat_message_chat_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message
        ADD CONSTRAINT chat_message_chat_room_id_fkey FOREIGN KEY (chat_room_id) REFERENCES public.chat_room(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_message_reaction chat_message_reaction_chat_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message_reaction
        ADD CONSTRAINT chat_message_reaction_chat_message_id_fkey FOREIGN KEY (chat_message_id) REFERENCES public.chat_message(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_message_reaction chat_message_reaction_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message_reaction
        ADD CONSTRAINT chat_message_reaction_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_message_read_by chat_message_read_by_chat_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message_read_by
        ADD CONSTRAINT chat_message_read_by_chat_message_id_fkey FOREIGN KEY (chat_message_id) REFERENCES public.chat_message(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_message_read_by chat_message_read_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message_read_by
        ADD CONSTRAINT chat_message_read_by_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_message chat_message_receiver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message
        ADD CONSTRAINT chat_message_receiver_id_fkey FOREIGN KEY (receiver_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_message chat_message_sender_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_message
        ADD CONSTRAINT chat_message_sender_id_fkey FOREIGN KEY (sender_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_room_admin chat_room_admin_chat_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_room_admin
        ADD CONSTRAINT chat_room_admin_chat_room_id_fkey FOREIGN KEY (chat_room_id) REFERENCES public.chat_room(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_room_admin chat_room_admin_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_room_admin
        ADD CONSTRAINT chat_room_admin_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_room_participant chat_room_participant_chat_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_room_participant
        ADD CONSTRAINT chat_room_participant_chat_room_id_fkey FOREIGN KEY (chat_room_id) REFERENCES public.chat_room(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: chat_room_participant chat_room_participant_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.chat_room_participant
        ADD CONSTRAINT chat_room_participant_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: doc_article doc_article_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.doc_article
        ADD CONSTRAINT doc_article_author_id_fkey FOREIGN KEY (author_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: doc_article doc_article_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.doc_article
        ADD CONSTRAINT doc_article_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.doc_category(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: doc_article_tag_relation doc_article_tag_relation_doc_article_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.doc_article_tag_relation
        ADD CONSTRAINT doc_article_tag_relation_doc_article_id_fkey FOREIGN KEY (doc_article_id) REFERENCES public.doc_article(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: doc_article_tag_relation doc_article_tag_relation_doc_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.doc_article_tag_relation
        ADD CONSTRAINT doc_article_tag_relation_doc_tag_id_fkey FOREIGN KEY (doc_tag_id) REFERENCES public.doc_tag(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_alias galgame_alias_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_alias
        ADD CONSTRAINT galgame_alias_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_comment galgame_comment_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_comment
        ADD CONSTRAINT galgame_comment_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_comment_like galgame_comment_like_galgame_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_comment_like
        ADD CONSTRAINT galgame_comment_like_galgame_comment_id_fkey FOREIGN KEY (galgame_comment_id) REFERENCES public.galgame_comment(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_comment_like galgame_comment_like_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_comment_like
        ADD CONSTRAINT galgame_comment_like_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_comment galgame_comment_target_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_comment
        ADD CONSTRAINT galgame_comment_target_user_id_fkey FOREIGN KEY (target_user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_comment galgame_comment_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_comment
        ADD CONSTRAINT galgame_comment_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_contributor galgame_contributor_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_contributor
        ADD CONSTRAINT galgame_contributor_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_contributor galgame_contributor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_contributor
        ADD CONSTRAINT galgame_contributor_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_engine_relation galgame_engine_relation_engine_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_engine_relation
        ADD CONSTRAINT galgame_engine_relation_engine_id_fkey FOREIGN KEY (engine_id) REFERENCES public.galgame_engine(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_engine_relation galgame_engine_relation_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_engine_relation
        ADD CONSTRAINT galgame_engine_relation_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_favorite galgame_favorite_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_favorite
        ADD CONSTRAINT galgame_favorite_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_favorite galgame_favorite_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_favorite
        ADD CONSTRAINT galgame_favorite_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_history galgame_history_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_history
        ADD CONSTRAINT galgame_history_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_history galgame_history_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_history
        ADD CONSTRAINT galgame_history_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_like galgame_like_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_like
        ADD CONSTRAINT galgame_like_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_like galgame_like_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_like
        ADD CONSTRAINT galgame_like_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_link galgame_link_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_link
        ADD CONSTRAINT galgame_link_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_link galgame_link_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_link
        ADD CONSTRAINT galgame_link_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_official_alias galgame_official_alias_galgame_official_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_official_alias
        ADD CONSTRAINT galgame_official_alias_galgame_official_id_fkey FOREIGN KEY (galgame_official_id) REFERENCES public.galgame_official(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_official_relation galgame_official_relation_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_official_relation
        ADD CONSTRAINT galgame_official_relation_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_official_relation galgame_official_relation_official_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_official_relation
        ADD CONSTRAINT galgame_official_relation_official_id_fkey FOREIGN KEY (official_id) REFERENCES public.galgame_official(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_pr galgame_pr_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_pr
        ADD CONSTRAINT galgame_pr_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_pr galgame_pr_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_pr
        ADD CONSTRAINT galgame_pr_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating_comment galgame_rating_comment_galgame_rating_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating_comment
        ADD CONSTRAINT galgame_rating_comment_galgame_rating_id_fkey FOREIGN KEY (galgame_rating_id) REFERENCES public.galgame_rating(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating_comment galgame_rating_comment_target_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating_comment
        ADD CONSTRAINT galgame_rating_comment_target_user_id_fkey FOREIGN KEY (target_user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating_comment galgame_rating_comment_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating_comment
        ADD CONSTRAINT galgame_rating_comment_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating galgame_rating_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating
        ADD CONSTRAINT galgame_rating_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE RESTRICT;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating_like galgame_rating_like_galgame_rating_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating_like
        ADD CONSTRAINT galgame_rating_like_galgame_rating_id_fkey FOREIGN KEY (galgame_rating_id) REFERENCES public.galgame_rating(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating_like galgame_rating_like_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating_like
        ADD CONSTRAINT galgame_rating_like_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_rating galgame_rating_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_rating
        ADD CONSTRAINT galgame_rating_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_resource galgame_resource_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_resource
        ADD CONSTRAINT galgame_resource_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_resource_like galgame_resource_like_galgame_resource_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_resource_like
        ADD CONSTRAINT galgame_resource_like_galgame_resource_id_fkey FOREIGN KEY (galgame_resource_id) REFERENCES public.galgame_resource(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_resource_like galgame_resource_like_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_resource_like
        ADD CONSTRAINT galgame_resource_like_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_resource_link galgame_resource_link_galgame_resource_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_resource_link
        ADD CONSTRAINT galgame_resource_link_galgame_resource_id_fkey FOREIGN KEY (galgame_resource_id) REFERENCES public.galgame_resource(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_resource galgame_resource_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_resource
        ADD CONSTRAINT galgame_resource_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame galgame_series_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame
        ADD CONSTRAINT galgame_series_id_fkey FOREIGN KEY (series_id) REFERENCES public.galgame_series(id) ON UPDATE CASCADE ON DELETE SET NULL;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_tag_alias galgame_tag_alias_galgame_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_tag_alias
        ADD CONSTRAINT galgame_tag_alias_galgame_tag_id_fkey FOREIGN KEY (galgame_tag_id) REFERENCES public.galgame_tag(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_tag_relation galgame_tag_relation_galgame_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_tag_relation
        ADD CONSTRAINT galgame_tag_relation_galgame_id_fkey FOREIGN KEY (galgame_id) REFERENCES public.galgame(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_tag_relation galgame_tag_relation_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_tag_relation
        ADD CONSTRAINT galgame_tag_relation_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.galgame_tag(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_alias galgame_toolset_alias_toolset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_alias
        ADD CONSTRAINT galgame_toolset_alias_toolset_id_fkey FOREIGN KEY (toolset_id) REFERENCES public.galgame_toolset(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_category_relation galgame_toolset_category_relation_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_category_relation
        ADD CONSTRAINT galgame_toolset_category_relation_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.galgame_toolset_category(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_category_relation galgame_toolset_category_relation_toolset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_category_relation
        ADD CONSTRAINT galgame_toolset_category_relation_toolset_id_fkey FOREIGN KEY (toolset_id) REFERENCES public.galgame_toolset(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_comment galgame_toolset_comment_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_comment
        ADD CONSTRAINT galgame_toolset_comment_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.galgame_toolset_comment(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_comment galgame_toolset_comment_toolset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_comment
        ADD CONSTRAINT galgame_toolset_comment_toolset_id_fkey FOREIGN KEY (toolset_id) REFERENCES public.galgame_toolset(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_comment galgame_toolset_comment_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_comment
        ADD CONSTRAINT galgame_toolset_comment_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_contributor galgame_toolset_contributor_toolset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_contributor
        ADD CONSTRAINT galgame_toolset_contributor_toolset_id_fkey FOREIGN KEY (toolset_id) REFERENCES public.galgame_toolset(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_contributor galgame_toolset_contributor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_contributor
        ADD CONSTRAINT galgame_toolset_contributor_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_practicality galgame_toolset_practicality_toolset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_practicality
        ADD CONSTRAINT galgame_toolset_practicality_toolset_id_fkey FOREIGN KEY (toolset_id) REFERENCES public.galgame_toolset(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_practicality galgame_toolset_practicality_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_practicality
        ADD CONSTRAINT galgame_toolset_practicality_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_resource galgame_toolset_resource_toolset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_resource
        ADD CONSTRAINT galgame_toolset_resource_toolset_id_fkey FOREIGN KEY (toolset_id) REFERENCES public.galgame_toolset(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset_resource galgame_toolset_resource_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset_resource
        ADD CONSTRAINT galgame_toolset_resource_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_toolset galgame_toolset_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_toolset
        ADD CONSTRAINT galgame_toolset_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame galgame_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame
        ADD CONSTRAINT galgame_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website galgame_website_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website
        ADD CONSTRAINT galgame_website_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.galgame_website_category(id) ON UPDATE CASCADE ON DELETE RESTRICT;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_comment galgame_website_comment_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_comment
        ADD CONSTRAINT galgame_website_comment_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.galgame_website_comment(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_comment galgame_website_comment_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_comment
        ADD CONSTRAINT galgame_website_comment_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_comment galgame_website_comment_website_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_comment
        ADD CONSTRAINT galgame_website_comment_website_id_fkey FOREIGN KEY (website_id) REFERENCES public.galgame_website(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_favorite galgame_website_favorite_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_favorite
        ADD CONSTRAINT galgame_website_favorite_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_favorite galgame_website_favorite_website_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_favorite
        ADD CONSTRAINT galgame_website_favorite_website_id_fkey FOREIGN KEY (website_id) REFERENCES public.galgame_website(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_like galgame_website_like_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_like
        ADD CONSTRAINT galgame_website_like_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_like galgame_website_like_website_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_like
        ADD CONSTRAINT galgame_website_like_website_id_fkey FOREIGN KEY (website_id) REFERENCES public.galgame_website(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_tag_relation galgame_website_tag_relation_galgame_website_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_tag_relation
        ADD CONSTRAINT galgame_website_tag_relation_galgame_website_id_fkey FOREIGN KEY (galgame_website_id) REFERENCES public.galgame_website(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website_tag_relation galgame_website_tag_relation_galgame_website_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website_tag_relation
        ADD CONSTRAINT galgame_website_tag_relation_galgame_website_tag_id_fkey FOREIGN KEY (galgame_website_tag_id) REFERENCES public.galgame_website_tag(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: galgame_website galgame_website_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.galgame_website
        ADD CONSTRAINT galgame_website_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: message message_receiver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.message
        ADD CONSTRAINT message_receiver_id_fkey FOREIGN KEY (receiver_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: message message_sender_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.message
        ADD CONSTRAINT message_sender_id_fkey FOREIGN KEY (sender_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: system_message system_message_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.system_message
        ADD CONSTRAINT system_message_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: todo todo_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.todo
        ADD CONSTRAINT todo_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic topic_best_answer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic
        ADD CONSTRAINT topic_best_answer_id_fkey FOREIGN KEY (best_answer_id) REFERENCES public.topic_reply(id) ON DELETE SET NULL;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_comment_like topic_comment_like_topic_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_comment_like
        ADD CONSTRAINT topic_comment_like_topic_comment_id_fkey FOREIGN KEY (topic_comment_id) REFERENCES public.topic_comment(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_comment_like topic_comment_like_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_comment_like
        ADD CONSTRAINT topic_comment_like_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_comment topic_comment_target_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_comment
        ADD CONSTRAINT topic_comment_target_user_id_fkey FOREIGN KEY (target_user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_comment topic_comment_topic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_comment
        ADD CONSTRAINT topic_comment_topic_id_fkey FOREIGN KEY (topic_id) REFERENCES public.topic(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_comment topic_comment_topic_reply_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_comment
        ADD CONSTRAINT topic_comment_topic_reply_id_fkey FOREIGN KEY (topic_reply_id) REFERENCES public.topic_reply(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_comment topic_comment_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_comment
        ADD CONSTRAINT topic_comment_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_dislike topic_dislike_topic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_dislike
        ADD CONSTRAINT topic_dislike_topic_id_fkey FOREIGN KEY (topic_id) REFERENCES public.topic(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_dislike topic_dislike_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_dislike
        ADD CONSTRAINT topic_dislike_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_favorite topic_favorite_topic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_favorite
        ADD CONSTRAINT topic_favorite_topic_id_fkey FOREIGN KEY (topic_id) REFERENCES public.topic(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_favorite topic_favorite_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_favorite
        ADD CONSTRAINT topic_favorite_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_like topic_like_topic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_like
        ADD CONSTRAINT topic_like_topic_id_fkey FOREIGN KEY (topic_id) REFERENCES public.topic(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_like topic_like_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_like
        ADD CONSTRAINT topic_like_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic topic_pinned_reply_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic
        ADD CONSTRAINT topic_pinned_reply_id_fkey FOREIGN KEY (pinned_reply_id) REFERENCES public.topic_reply(id) ON DELETE SET NULL;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_poll_option topic_poll_option_poll_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_poll_option
        ADD CONSTRAINT topic_poll_option_poll_id_fkey FOREIGN KEY (poll_id) REFERENCES public.topic_poll(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_poll topic_poll_topic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_poll
        ADD CONSTRAINT topic_poll_topic_id_fkey FOREIGN KEY (topic_id) REFERENCES public.topic(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_poll topic_poll_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_poll
        ADD CONSTRAINT topic_poll_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_poll_vote topic_poll_vote_option_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_poll_vote
        ADD CONSTRAINT topic_poll_vote_option_id_fkey FOREIGN KEY (option_id) REFERENCES public.topic_poll_option(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_poll_vote topic_poll_vote_poll_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_poll_vote
        ADD CONSTRAINT topic_poll_vote_poll_id_fkey FOREIGN KEY (poll_id) REFERENCES public.topic_poll(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_poll_vote topic_poll_vote_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_poll_vote
        ADD CONSTRAINT topic_poll_vote_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply_dislike topic_reply_dislike_topic_reply_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply_dislike
        ADD CONSTRAINT topic_reply_dislike_topic_reply_id_fkey FOREIGN KEY (topic_reply_id) REFERENCES public.topic_reply(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply_dislike topic_reply_dislike_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply_dislike
        ADD CONSTRAINT topic_reply_dislike_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply_like topic_reply_like_topic_reply_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply_like
        ADD CONSTRAINT topic_reply_like_topic_reply_id_fkey FOREIGN KEY (topic_reply_id) REFERENCES public.topic_reply(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply_like topic_reply_like_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply_like
        ADD CONSTRAINT topic_reply_like_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply_target topic_reply_target_reply_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply_target
        ADD CONSTRAINT topic_reply_target_reply_id_fkey FOREIGN KEY (reply_id) REFERENCES public.topic_reply(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply_target topic_reply_target_target_reply_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply_target
        ADD CONSTRAINT topic_reply_target_target_reply_id_fkey FOREIGN KEY (target_reply_id) REFERENCES public.topic_reply(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply topic_reply_topic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply
        ADD CONSTRAINT topic_reply_topic_id_fkey FOREIGN KEY (topic_id) REFERENCES public.topic(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_reply topic_reply_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_reply
        ADD CONSTRAINT topic_reply_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_section_relation topic_section_relation_topic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_section_relation
        ADD CONSTRAINT topic_section_relation_topic_id_fkey FOREIGN KEY (topic_id) REFERENCES public.topic(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_section_relation topic_section_relation_topic_section_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_section_relation
        ADD CONSTRAINT topic_section_relation_topic_section_id_fkey FOREIGN KEY (topic_section_id) REFERENCES public.topic_section(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_upvote topic_upvote_topic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_upvote
        ADD CONSTRAINT topic_upvote_topic_id_fkey FOREIGN KEY (topic_id) REFERENCES public.topic(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic_upvote topic_upvote_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic_upvote
        ADD CONSTRAINT topic_upvote_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: topic topic_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.topic
        ADD CONSTRAINT topic_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: unmoe unmoe_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.unmoe
        ADD CONSTRAINT unmoe_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: update_log update_log_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.update_log
        ADD CONSTRAINT update_log_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: user_follow user_follow_followed_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.user_follow
        ADD CONSTRAINT user_follow_followed_id_fkey FOREIGN KEY (followed_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: user_follow user_follow_follower_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.user_follow
        ADD CONSTRAINT user_follow_follower_id_fkey FOREIGN KEY (follower_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: user_friend user_friend_friend_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.user_friend
        ADD CONSTRAINT user_friend_friend_id_fkey FOREIGN KEY (friend_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
-- Name: user_friend user_friend_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $$
BEGIN
    ALTER TABLE ONLY public.user_friend
        ADD CONSTRAINT user_friend_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id) ON UPDATE CASCADE ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'baseline skip: %', SQLERRM;
END $$;

--
--
