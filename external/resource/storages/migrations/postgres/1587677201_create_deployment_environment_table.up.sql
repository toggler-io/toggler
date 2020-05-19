CREATE TABLE deployment_environments
(
    id   UUID NOT NULL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX find_deployment_environment_by_name
    ON deployment_environments USING btree (name);
