CREATE TABLE release_allows
(
    id      BIGSERIAL NOT NULL PRIMARY KEY,
    flag_id BIGINT    NOT NULL,
    ip_addr INET
);