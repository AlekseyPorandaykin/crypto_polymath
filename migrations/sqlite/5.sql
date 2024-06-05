CREATE TABLE IF NOT EXISTS analytics
(
    id              VARCHAR(50) PRIMARY KEY NOT NULL,
    name            VARCHAR(50)             NOT NULL,
    exchange        VARCHAR(50)             NOT NULL,
    symbol          VARCHAR(50)             NOT NULL,
    unit            VARCHAR(10)             NOT NULL,
    interval        INT                     NOT NULL DEFAULT 1,
    datetime        TIMESTAMP               NOT NULL,
    depth           INT                     NOT NULL DEFAULT 1,
    by_indicator    VARCHAR(50)             NOT NULL,
    indicator_depth INT                     NOT NULL DEFAULT 1,
    value           double precision        NOT NULL DEFAULT 0,
    created_at      TIMESTAMP               NOT NULL DEFAULT CURRENT_TIMESTAMP
);