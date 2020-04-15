CREATE TABLE "feature_flags"
(
    id                                  UUID   NOT NULL PRIMARY KEY,
    name                                TEXT   NOT NULL,
    rollout_rand_seed                   BIGINT NOT NULL,

    -- fix this to SMALLINT
    -- this should be fixed by having the percentage as a int8 in the model
    rollout_strategy_percentage         BIGINT NOT NULL,
    rollout_strategy_decision_logic_api TEXT,

    CONSTRAINT feature_flags_name_is_uniq UNIQUE (name)
);

CREATE INDEX look_feature_flags_by_name ON feature_flags USING btree (name);
