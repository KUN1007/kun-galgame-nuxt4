-- Link toolset s3 resources to the centralized artifact service
-- (nextmoe-infra). New s3 uploads store the artifact uuid here; the download
-- URL is resolved at read time via the artifact service. Legacy s3 rows keep
-- artifact_uuid='' and their stored Content URL (forward-only; dual-read).
ALTER TABLE galgame_toolset_resource
  ADD COLUMN IF NOT EXISTS artifact_uuid varchar(36) NOT NULL DEFAULT '';

-- One uuid maps to at most one resource row.
CREATE UNIQUE INDEX IF NOT EXISTS idx_toolset_resource_artifact_uuid
  ON galgame_toolset_resource (artifact_uuid)
  WHERE artifact_uuid <> '';
