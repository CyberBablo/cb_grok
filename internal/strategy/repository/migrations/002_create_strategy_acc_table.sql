-- Create strategy_acc table
CREATE TABLE IF NOT EXISTS strategy_acc(
    id bigint primary key generated always as identity,
    symbol VARCHAR(512) not null,
    timeframe VARCHAR(255) not null,
    strategy_id bigint not null,
    dt_upd timestamp without time zone not null default now(),
    dt_create timestamp without time zone not null default now(),
    FOREIGN KEY (strategy_id) REFERENCES strategy(id)
);

CREATE UNIQUE INDEX uidx__strategy_acc__core on strategy_acc(symbol, timeframe, strategy_id);
CREATE INDEX idx__strategy_acc on strategy_acc(symbol, timeframe);


CREATE TABLE IF NOT EXISTS strategy_timeframe(
    id bigint primary key generated always as identity,
    name VARCHAR(512) not null,
		value json not null
);
