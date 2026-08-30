-- 083: the topic lottery mini-app (抽奖小程序).
--
-- A sibling of topic_poll, not a variant of it: a poll aggregates revisable
-- opinions into counts, a lottery allocates scarce goods once and has winners,
-- secrets and fulfilment. Four tables instead of poll's three because the prize
-- (a tier with N slots) and the secret it hands out (one activation code per
-- slot) are separate things with different access rules.
--
-- topic_lottery_code is deliberately its own table rather than a column on the
-- prize: it is the only table here that must never be joined into a list or
-- detail response, and a separate table makes that reviewable.
--
-- seed/seed_hash implement commit-reveal. seed_hash is published when the
-- lottery is created and seed only when it is drawn, so anyone can recompute
-- the ranking afterwards. Floor lotteries need no randomness and leave both
-- empty.
--
-- No notification_sent flag anywhere: topic_poll has had a dead one since it
-- was created (declared, never read, never written), and the draw already has a
-- status transition that says whether the notification went out.

BEGIN;

CREATE TABLE IF NOT EXISTS topic_lottery (
  id                   serial PRIMARY KEY,
  topic_id             integer NOT NULL,
  user_id              integer NOT NULL,
  title                varchar(100) NOT NULL,
  description          varchar(1000) NOT NULL DEFAULT '',

  entry_mode           text NOT NULL DEFAULT 'signup',
  floor_rule           text NOT NULL DEFAULT '',
  draw_mode            text NOT NULL DEFAULT 'deadline',
  draw_threshold       integer NOT NULL DEFAULT 0,
  deadline             timestamptz,

  min_account_age_days integer NOT NULL DEFAULT 0,
  min_moemoepoint      integer NOT NULL DEFAULT 0,
  show_entrants        boolean NOT NULL DEFAULT true,

  status               text NOT NULL DEFAULT 'open',
  seed_hash            text NOT NULL DEFAULT '',
  seed                 text NOT NULL DEFAULT '',
  entry_count          integer NOT NULL DEFAULT 0,
  drawn_at             timestamptz,

  created              timestamptz NOT NULL DEFAULT now(),
  updated              timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_topic_lottery_topic ON topic_lottery (topic_id);
-- The auto-draw sweep runs every minute and this is the whole of its WHERE.
CREATE INDEX IF NOT EXISTS idx_topic_lottery_due ON topic_lottery (status, deadline);

COMMENT ON COLUMN topic_lottery.seed IS
  'Empty until the draw. Publishing it early would let the author compute the winners before entries close.';

CREATE TABLE IF NOT EXISTS topic_lottery_prize (
  id           serial PRIMARY KEY,
  lottery_id   integer NOT NULL REFERENCES topic_lottery(id) ON DELETE CASCADE,
  name         varchar(100) NOT NULL,
  description  varchar(500) NOT NULL DEFAULT '',
  image_hash   text NOT NULL DEFAULT '',
  delivery     text NOT NULL DEFAULT 'manual',
  point_amount integer NOT NULL DEFAULT 0,
  slots        integer NOT NULL DEFAULT 1,
  sort_order   integer NOT NULL DEFAULT 0,
  created      timestamptz NOT NULL DEFAULT now(),
  updated      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_topic_lottery_prize_lottery ON topic_lottery_prize (lottery_id, sort_order);

CREATE TABLE IF NOT EXISTS topic_lottery_code (
  id         serial PRIMARY KEY,
  lottery_id integer NOT NULL REFERENCES topic_lottery(id) ON DELETE CASCADE,
  prize_id   integer NOT NULL REFERENCES topic_lottery_prize(id) ON DELETE CASCADE,
  secret     text NOT NULL,
  claimed_by integer NOT NULL DEFAULT 0,
  claimed_at timestamptz,
  created    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_topic_lottery_code_prize ON topic_lottery_code (prize_id, claimed_by);

COMMENT ON TABLE topic_lottery_code IS
  'Escrowed activation codes, AES-GCM sealed. Never join this into a list or detail response: Nuxt serialises every fetched payload into the SSR __NUXT__ blob, so one leak into a DTO publishes every code in the page source. The only read path is the authenticated claim endpoint.';

CREATE TABLE IF NOT EXISTS topic_lottery_entry (
  id          serial PRIMARY KEY,
  lottery_id  integer NOT NULL REFERENCES topic_lottery(id) ON DELETE CASCADE,
  user_id     integer NOT NULL,
  reply_floor integer NOT NULL DEFAULT 0,
  prize_id    integer NOT NULL DEFAULT 0,
  code_id     integer NOT NULL DEFAULT 0,
  rank_key    text NOT NULL DEFAULT '',
  fulfillment text NOT NULL DEFAULT '',
  won_at      timestamptz,
  created     timestamptz NOT NULL DEFAULT now(),
  updated     timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_topic_lottery_entry UNIQUE (lottery_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_topic_lottery_entry_user ON topic_lottery_entry (user_id);
CREATE INDEX IF NOT EXISTS idx_topic_lottery_entry_winner ON topic_lottery_entry (lottery_id, prize_id);

COMMIT;
