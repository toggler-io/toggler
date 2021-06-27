DROP VIEW "deployment_environments";

ALTER TABLE "release_environments"
    RENAME TO "deployment_environments";

ALTER TABLE "release_pilots"
    RENAME COLUMN "public_id" TO "external_id";

ALTER INDEX "lookup_pilots_by_flag_id_and_public_id"
    RENAME TO "lookup_pilots_by_feature_flag_id_and_external_id";
