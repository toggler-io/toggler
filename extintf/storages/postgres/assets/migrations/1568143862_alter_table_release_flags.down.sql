BEGIN;

DROP VIEW "feature_flags";

ALTER INDEX "look_release_flag_by_name"
    RENAME TO "look_feature_flags_by_name";

ALTER TABLE "release_flags"
    RENAME TO "feature_flags";

COMMIT;