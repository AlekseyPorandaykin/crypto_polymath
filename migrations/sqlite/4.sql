CREATE TABLE IF NOT EXISTS symbol_infos
(
    id          VARCHAR(50) PRIMARY KEY NOT NULL,
    exchange    VARCHAR(50)             NOT NULL,
    symbol      VARCHAR(50)             NOT NULL,
    base_asset  VARCHAR(20)             NOT NULL,
    quote_asset VARCHAR(20)             NOT NULL,
    created_at  TIMESTAMP               NOT NULL DEFAULT CURRENT_TIMESTAMP
);