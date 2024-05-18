CREATE TABLE IF NOT EXISTS prices
(
    id         VARCHAR(50) PRIMARY KEY NOT NULL,
    symbol     VARCHAR(50)             NOT NULL,
    exchange   VARCHAR(50)             NOT NULL,
    value      double precision        NOT NULL DEFAULT 0,
    created_at TIMESTAMP               NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX prices_exchange_idx ON prices (exchange);
CREATE INDEX prices_symbol_idx ON prices (symbol);