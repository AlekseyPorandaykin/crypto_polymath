CREATE SCHEMA IF NOT EXISTS crypto_polymath;

CREATE TABLE IF NOT EXISTS crypto_polymath.prices
(
    id         VARCHAR(50) PRIMARY KEY NOT NULL,
    symbol     VARCHAR(50)             NOT NULL,
    exchange   VARCHAR(50)             NOT NULL,
    value      double precision        NOT NULL DEFAULT 0,
    created_at TIMESTAMP               NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX prices_exchange_idx ON crypto_polymath.prices (exchange);
CREATE INDEX prices_symbol_idx ON crypto_polymath.prices (symbol);

CREATE TABLE IF NOT EXISTS crypto_polymath.candlestick
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

CREATE INDEX candlestick_exchange_idx ON crypto_polymath.candlestick (exchange);
CREATE INDEX candlestick_symbol_idx ON crypto_polymath.candlestick (symbol);
CREATE INDEX candlestick_unit_idx ON crypto_polymath.candlestick (unit);
CREATE INDEX candlestick_interval_idx ON crypto_polymath.candlestick (interval);
CREATE INDEX candlestick_start_time_idx ON crypto_polymath.candlestick (start_time);

CREATE TABLE IF NOT EXISTS crypto_polymath.indicators
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
CREATE INDEX IF NOT EXISTS indicators_exchange_idx ON crypto_polymath.indicators (exchange);
CREATE INDEX IF NOT EXISTS indicators_symbol_idx ON crypto_polymath.indicators (symbol);
CREATE INDEX IF NOT EXISTS indicators_unit_idx ON crypto_polymath.indicators (unit);
CREATE INDEX IF NOT EXISTS indicators_interval_idx ON crypto_polymath.indicators (interval);
CREATE INDEX IF NOT EXISTS indicators_datetime_idx ON crypto_polymath.indicators (datetime);
CREATE INDEX IF NOT EXISTS indicators_name_idx ON crypto_polymath.indicators (name);
CREATE INDEX IF NOT EXISTS indicators_depth_idx ON crypto_polymath.indicators (depth);

CREATE TABLE IF NOT EXISTS crypto_polymath.symbol_infos
(
    id                VARCHAR(50) PRIMARY KEY NOT NULL,
    exchange          VARCHAR(50)             NOT NULL,
    symbol            VARCHAR(50)             NOT NULL,
    base_asset        VARCHAR(20)             NOT NULL,
    quote_asset       VARCHAR(20)             NOT NULL,
    category          VARCHAR(50)             NOT NULL DEFAULT '',
    funding_rate      double precision        NOT NULL DEFAULT 0,
    next_funding_time TIMESTAMP               NULL,
    created_at        TIMESTAMP               NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE crypto_polymath.symbol_infos
    ADD COLUMN IF NOT EXISTS funding_rate double precision NOT NULL DEFAULT 0;

ALTER TABLE crypto_polymath.symbol_infos
    ADD COLUMN IF NOT EXISTS next_funding_time TIMESTAMP NULL;

CREATE TABLE IF NOT EXISTS crypto_polymath.analytics
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


CREATE TABLE IF NOT EXISTS crypto_polymath.candlestick_indicators
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

CREATE INDEX candlestick_indicators_name_idx ON crypto_polymath.candlestick_indicators (name);
CREATE INDEX candlestick_indicators_exchange_idx ON crypto_polymath.candlestick_indicators (exchange);
CREATE INDEX candlestick_indicators_symbol_idx ON crypto_polymath.candlestick_indicators (symbol);
CREATE INDEX candlestick_indicators_unit_idx ON crypto_polymath.candlestick_indicators (unit);
CREATE INDEX candlestick_indicators_interval_idx ON crypto_polymath.candlestick_indicators (interval);
CREATE INDEX candlestick_indicators_start_time_idx ON crypto_polymath.candlestick_indicators (start_time);


CREATE TABLE IF NOT EXISTS crypto_polymath.queues
(
    id         VARCHAR(50) PRIMARY KEY NOT NULL,
    name       VARCHAR(250)            NOT NULL,
    body       json                    NOT NULL,
    created_at TIMESTAMP               NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX queues_name ON crypto_polymath.queues (name);
