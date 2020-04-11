BEGIN;

ALTER INDEX "look_feature_flags_by_name"
    RENAME TO "look_release_flag_by_name";

ALTER TABLE "feature_flags"
    RENAME TO "release_flags";

CREATE VIEW "feature_flags" AS SELECT * FROM "release_flags";

COMMIT;