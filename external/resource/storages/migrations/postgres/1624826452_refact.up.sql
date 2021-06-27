ALTER TABLE "deployment_environments"
    RENAME TO "release_environments";

CREATE VIEW "deployment_environments" AS
SELECT *
FROM "release_environments";

-- Supporting backward compatible column names would require schema based versioning.
-- For the sake of simplicity in this case, this will be skipped due to no real production use.
ALTER TABLE "release_pilots"
    RENAME COLUMN "external_id" TO "public_id";

ALTER INDEX "lookup_pilots_by_feature_flag_id_and_external_id"
    RENAME TO "lookup_pilots_by_flag_id_and_public_id";
