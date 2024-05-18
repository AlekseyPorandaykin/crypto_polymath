CREATE TABLE IF NOT EXISTS indicators
(
    id         VARCHAR(50) PRIMARY KEY NOT NULL,
    symbol     VARCHAR(50)             NOT NULL,
    exchange   VARCHAR(50)             NOT NULL,
    unit       VARCHAR(10)             NOT NULL,
    interval   INT                     NOT NULL DEFAULT 1,
    name       VARCHAR(50)             NOT NULL,
    datetime   TIMESTAMP               NOT NULL,
    depth      INT                     NOT NULL DEFAULT 1,
    value      double precision        NOT NULL DEFAULT 0,
    created_at TIMESTAMP               NOT NULL DEFAULT CURRENT_TIMESTAMP
);