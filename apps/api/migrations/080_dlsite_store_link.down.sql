-- Drops the short-link cache. Nothing is lost: infra keeps the aliases and
-- re-answers the same ones, so 080 up plus normal traffic rebuilds the table.

BEGIN;

DROP TABLE IF EXISTS dlsite_store_link;

COMMIT;
