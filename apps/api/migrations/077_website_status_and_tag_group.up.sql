-- 077: website lifecycle status, and a real table for the tag groups.
--
-- status: 'normal' | 'unreachable' | 'closed'. Mirrors friend_link.status —
-- a plain text column with a default, no enum type, so adding a state later is
-- an INSERT-free change.
--
-- multi_select says whether the group's tags are mutually exclusive. The
-- frontend inferred it from group size — "more than one tag means pick one" —
-- which was only ever true by accident: 'misc' held six unrelated single-tag
-- groups, and the moment they became one real group of six that rule would
-- have made 补丁/生肉/wiki mutually exclusive.
--
-- galgame_website_tag_group: the tag taxonomy's second level. It only ever
-- existed as `tag.name.match(/^[a-z_]+/)` on the frontend plus a hardcoded
-- label map, so a group could not be renamed, added, or described without a
-- frontend deploy — and a tag whose name had no matching entry rendered under
-- an empty <legend>. group_id is nullable on purpose: ON DELETE SET NULL drops
-- the tags into the "未分组" bucket instead of taking them with the group.
--
-- Existing rows: the 18 groups below are the exact contents of the retired
-- KUN_TAG_CATEGORY_TITLE map, and every existing tag is assigned by the same
-- prefix rule the frontend used, so the rendered grouping is unchanged. The
-- six tags with no prefix group (patch, localization, raw, wiki, info, cloud)
-- land in 'misc', which is where the frontend's checkboxGroup already put them.
BEGIN;

ALTER TABLE galgame_website
  ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'normal';

-- The category order on /website was whatever order the frontend map literal
-- happened to be written in. Reading the categories from the table instead
-- would have reordered the page by id (telegram first), so the order becomes
-- data the admin owns, seeded to exactly what the page shows today.
ALTER TABLE galgame_website_category
  ADD COLUMN IF NOT EXISTS sort_order integer NOT NULL DEFAULT 0;

UPDATE galgame_website_category SET sort_order = 10 WHERE name = 'resource'  AND sort_order = 0;
UPDATE galgame_website_category SET sort_order = 20 WHERE name = 'community' AND sort_order = 0;
UPDATE galgame_website_category SET sort_order = 30 WHERE name = 'telegram'  AND sort_order = 0;
UPDATE galgame_website_category SET sort_order = 40 WHERE name = 'other'     AND sort_order = 0;

CREATE TABLE IF NOT EXISTS galgame_website_tag_group (
  id          SERIAL PRIMARY KEY,
  name        text NOT NULL UNIQUE,
  label       text NOT NULL DEFAULT '',
  description  text NOT NULL DEFAULT '',
  sort_order   integer NOT NULL DEFAULT 0,
  multi_select boolean NOT NULL DEFAULT false,
  created     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO galgame_website_tag_group (name, label, sort_order, multi_select) VALUES
  ('ad', '广告程度', 10, false),
  ('open', '开源情况', 20, false),
  ('resource', '资源数量', 30, false),
  ('update', '资源更新频率', 40, false),
  ('storage', '下载方式', 50, false),
  ('support', '补档速度', 60, false),
  ('trash_storage', '网赚盘比例', 70, false),
  ('age', '网站年龄', 80, false),
  ('threshold', '访问门槛', 90, false),
  ('intro', '内容介绍', 100, false),
  ('curation', '内容筛选与整理', 110, false),
  ('mobile', '移动端适配', 120, false),
  ('search', '搜索支持', 130, false),
  ('vibe', '网站氛围', 140, false),
  ('active', '社区活跃度', 150, false),
  ('performance', '网站性能', 160, false),
  ('security', '网站安全性', 170, false),
  ('misc', '网站类型', 180, true)
ON CONFLICT (name) DO NOTHING;

ALTER TABLE galgame_website_tag
  ADD COLUMN IF NOT EXISTS group_id integer
  REFERENCES galgame_website_tag_group(id) ON UPDATE CASCADE ON DELETE SET NULL;

UPDATE galgame_website_tag t
   SET group_id = g.id
  FROM galgame_website_tag_group g
 WHERE t.group_id IS NULL
   AND g.name = substring(t.name from '^[a-z_]+');

UPDATE galgame_website_tag
   SET group_id = (SELECT id FROM galgame_website_tag_group WHERE name = 'misc')
 WHERE group_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_galgame_website_tag_group_id
  ON galgame_website_tag (group_id);
CREATE INDEX IF NOT EXISTS idx_galgame_website_status
  ON galgame_website (status);

COMMIT;
