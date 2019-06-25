CREATE TABLE "tokens"
(
    id        BIGSERIAL   NOT NULL,
    token     TEXT        NOT NULL,
    owner_uid TEXT        NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    duration  BIGINT      NOT NULL,

    CONSTRAINT tokens_token_is_uniq UNIQUE (token)
);

CREATE INDEX lookup_tokens_by_token_text ON tokens USING btree (token);
