-- Create strategy table
CREATE TABLE IF NOT EXISTS strategy(
    id bigint primary key generated always as identity,
    symbol varchar(512) not null,
    timeframe varchar(255) not null,
    trials int not null default 0,
    workers int not null default 0,
    val_set_days int not null default 0,
    train_set_day int not null default 0,
    win_rate double precision not null,
    data json not null,
    from_dt bigint not null default (EXTRACT(EPOCH FROM now())::BIGINT),
    to_dt bigint not null default (EXTRACT(EPOCH FROM now())::BIGINT),
    dt timestamp without time zone not null default now()
);
