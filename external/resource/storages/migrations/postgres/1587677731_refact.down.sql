ALTER TABLE "release_pilots"
    DROP CONSTRAINT "pilot_uniq_combination",
    ADD CONSTRAINT "pilot_uniq_combination" UNIQUE ("flag_id", "external_id");

ALTER TABLE "release_pilots"
    RENAME COLUMN "flag_id" TO "feature_flag_id";

ALTER TABLE "release_pilots"
    RENAME COLUMN "is_participating"
        TO "enrolled";

ALTER TABLE "release_pilots"
    RENAME TO "pilots";

ALTER TABLE "release_flags"
    ADD COLUMN "rollout_rand_seed"                   BIGINT NOT NULL DEFAULT 424242,
    ADD COLUMN "rollout_strategy_percentage"         BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN "rollout_strategy_decision_logic_api" TEXT;

CREATE VIEW "feature_flags" AS
SELECT *
FROM "release_flags";
