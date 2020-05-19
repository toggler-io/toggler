ALTER TABLE "pilots"
    RENAME TO "release_pilots";

ALTER TABLE "release_pilots"
    RENAME COLUMN "feature_flag_id" TO "flag_id";

ALTER TABLE "release_pilots"
    DROP CONSTRAINT "pilot_uniq_combination",
    ADD CONSTRAINT "pilot_uniq_combination" UNIQUE ("flag_id", "env_id", "external_id");

ALTER TABLE "release_pilots"
    RENAME COLUMN "enrolled"
        TO "is_participating";

DROP VIEW "feature_flags";

ALTER TABLE "release_flags"
    DROP COLUMN "rollout_rand_seed",
    DROP COLUMN "rollout_strategy_percentage",
    DROP COLUMN "rollout_strategy_decision_logic_api";
