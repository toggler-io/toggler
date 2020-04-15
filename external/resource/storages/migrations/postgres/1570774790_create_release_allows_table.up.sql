CREATE TABLE release_flag_ip_addr_allows
(
    id      UUID NOT NULL PRIMARY KEY,
    flag_id UUID NOT NULL,
    ip_addr INET
);

CREATE INDEX find_release_flag_ip_addr_allows_by_flag_id
    ON release_flag_ip_addr_allows USING btree (flag_id);
