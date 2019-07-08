CREATE TABLE "pilots"
(
    id              BIGSERIAL NOT NULL,
    feature_flag_id BIGINT    NOT NULL,
    external_id     TEXT      NOT NULL,
    enrolled        BOOLEAN   NOT NULL,

    CONSTRAINT pilot_uniq_combination UNIQUE (feature_flag_id, external_id)
);

CREATE INDEX lookup_pilots_by_feature_flag_id_and_external_id ON pilots USING btree (feature_flag_id, external_id);
