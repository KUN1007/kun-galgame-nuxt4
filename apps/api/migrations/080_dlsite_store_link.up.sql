-- 080: local cache of the DLsite purchase short links infra mints for this site.
--
-- GET /v2/store/purchase-links/{product_id} answers Cache-Control: private and
-- is rate limited per developer key, so it cannot be called on a render path —
-- a resource list would spend a whole minute's budget on one page view. It does
-- not have to be: infra pins one alias per (client, product) forever
-- (store_purchase_links has a unique index and inserts ON CONFLICT DO NOTHING,
-- precisely so attribution cannot move), so one mint per product is enough for
-- the life of the site and the answer belongs in a table here.
--
-- product_id is the DLsite workno (RJ doujin / VJ commercial) and the natural
-- key; there is exactly one short link per product for this site. varchar(10)
-- is the store face's own maxLength.
--
-- No backfill and no seeding: the resolver mints lazily as galgames are viewed,
-- which keeps infra's per-client link quota spent on products readers actually
-- open. Until a product is minted its purchase button uses the bare affiliate
-- template (KUN_DLSITE_LINK_TEMPLATE) — the sale still works, only the click
-- attribution is lost, and only for the first reader.
--
-- Safe to drop: 080 down plus normal traffic rebuilds it.

BEGIN;

CREATE TABLE IF NOT EXISTS dlsite_store_link (
  product_id varchar(10) PRIMARY KEY,
  short_url  text        NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE dlsite_store_link IS
  'Cache of infra /v2/store purchase short links (080). One row per DLsite workno; infra owns the alias, this is a copy so the render path never calls out.';

COMMIT;
