-- Create strategy_acc table
CREATE TABLE IF NOT EXISTS strategy_acc(
    id bigint primary key generated always as identity,
    symbol VARCHAR(512) not null,
    timeframe VARCHAR(255) not null,
    strategy_id bigint not null,
    dt_upd timestamp without time zone not null default now(),
    dt_create timestamp without time zone not null default now()
);
