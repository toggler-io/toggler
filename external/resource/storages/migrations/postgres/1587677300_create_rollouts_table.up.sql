CREATE TABLE release_rollouts
(
    id      UUID NOT NULL PRIMARY KEY,
    flag_id UUID NOT NULL,
    env_id  UUID NOT NULL,
    plan    JSON NOT NULL
);

CREATE INDEX find_release_rollout_by_release_flag_and_deployment_environment
    ON release_rollouts USING btree (flag_id, env_id);
