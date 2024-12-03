ALTER TABLE symbol_infos
    ADD COLUMN category VARCHAR(50) NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS candlestick_indicators
(
    id          VARCHAR(50) PRIMARY KEY NOT NULL,
    name        VARCHAR(250)            NOT NULL,
    exchange    VARCHAR(50)             NOT NULL,
    symbol      VARCHAR(50)             NOT NULL,
    unit        VARCHAR(10)             NOT NULL,
    interval    INT                     NOT NULL DEFAULT 1,
    start_time  TIMESTAMP               NOT NULL,
    open_price  double precision        NOT NULL DEFAULT 0,
    high_price  double precision        NOT NULL DEFAULT 0,
    low_price   double precision        NOT NULL DEFAULT 0,
    close_price double precision        NOT NULL DEFAULT 0,
    created_at  TIMESTAMP               NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX candlestick_indicators_name_idx ON candlestick_indicators (name);
CREATE INDEX candlestick_indicators_exchange_idx ON candlestick_indicators (exchange);
CREATE INDEX candlestick_indicators_symbol_idx ON candlestick_indicators (symbol);
CREATE INDEX candlestick_indicators_unit_idx ON candlestick_indicators (unit);
CREATE INDEX candlestick_indicators_interval_idx ON candlestick_indicators (interval);
CREATE INDEX candlestick_indicators_start_time_idx ON candlestick_indicators (start_time);
