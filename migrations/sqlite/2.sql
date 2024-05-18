CREATE TABLE IF NOT EXISTS candlestick
(
    id          VARCHAR(50) PRIMARY KEY NOT NULL,
    symbol      VARCHAR(50)             NOT NULL,
    exchange    VARCHAR(50)             NOT NULL,
    unit        VARCHAR(10)             NOT NULL,
    interval    INT                     NOT NULL DEFAULT 1,
    start_time  TIMESTAMP               NOT NULL,
    open_price  double precision        NOT NULL DEFAULT 0,
    high_price  double precision        NOT NULL DEFAULT 0,
    low_price   double precision        NOT NULL DEFAULT 0,
    close_price double precision        NOT NULL DEFAULT 0,
    volume      double precision        NOT NULL DEFAULT 0,
    created_at  TIMESTAMP               NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX candlestick_exchange_idx ON candlestick (exchange);
CREATE INDEX candlestick_symbol_idx ON candlestick (symbol);
CREATE INDEX candlestick_unit_idx ON candlestick (unit);
CREATE INDEX candlestick_interval_idx ON candlestick (interval);
CREATE INDEX candlestick_start_time_idx ON candlestick (start_time);
