-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: candles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.candles (
    id integer NOT NULL,
    symbol character varying(50) NOT NULL,
    exchange character varying(50) NOT NULL,
    timeframe character varying(10) NOT NULL,
    "timestamp" bigint NOT NULL,
    open double precision NOT NULL,
    high double precision NOT NULL,
    low double precision NOT NULL,
    close double precision NOT NULL,
    volume double precision NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: candles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.candles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: candles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.candles_id_seq OWNED BY public.candles.id;


--
-- Name: exchange; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exchange (
    id integer NOT NULL,
    name character varying(50)
);


--
-- Name: exchange_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exchange_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exchange_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exchange_id_seq OWNED BY public.exchange.id;


--
-- Name: order; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public."order" (
    id bigint NOT NULL,
    symbol_id integer NOT NULL,
    exch_id integer NOT NULL,
    type_id integer NOT NULL,
    side_id integer NOT NULL,
    status_id integer NOT NULL,
    base_qty numeric(20,6),
    quote_qty numeric(20,6),
    ext_id character varying(300) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone,
    tp_price numeric(20,6),
    sl_price numeric(20,6)
);


--
-- Name: order_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.order_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: order_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.order_id_seq OWNED BY public."order".id;


--
-- Name: order_side; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_side (
    id integer NOT NULL,
    code character varying(10)
);


--
-- Name: order_side_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.order_side_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: order_side_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.order_side_id_seq OWNED BY public.order_side.id;


--
-- Name: order_status; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_status (
    id integer NOT NULL,
    code character varying(50)
);


--
-- Name: order_status_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.order_status_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: order_status_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.order_status_id_seq OWNED BY public.order_status.id;


--
-- Name: order_type; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_type (
    id integer NOT NULL,
    code character varying(30)
);


--
-- Name: order_type_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.order_type_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: order_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.order_type_id_seq OWNED BY public.order_type.id;


--
-- Name: product; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.product (
    id integer NOT NULL,
    code character varying(50)
);


--
-- Name: symbol; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.symbol (
    id integer NOT NULL,
    code character varying(100) NOT NULL,
    prod_id integer NOT NULL,
    base character varying(50) NOT NULL,
    quote character varying(50) NOT NULL
);


--
-- Name: order_v; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.order_v AS
 SELECT o.id,
    o.created_at,
    o.updated_at,
    sm.code AS symbol,
    ex.name AS exchange_name,
    p.code AS product_code,
    ot.code AS order_type,
    os.code AS order_status,
    oside.code AS order_side,
    o.base_qty,
    o.quote_qty,
    o.ext_id
   FROM ((((((public."order" o
     LEFT JOIN public.symbol sm ON ((o.symbol_id = sm.id)))
     LEFT JOIN public.exchange ex ON ((o.exch_id = ex.id)))
     LEFT JOIN public.product p ON ((sm.prod_id = p.id)))
     LEFT JOIN public.order_type ot ON ((o.type_id = ot.id)))
     LEFT JOIN public.order_status os ON ((o.status_id = os.id)))
     LEFT JOIN public.order_side oside ON ((o.side_id = oside.id)));


--
-- Name: product_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.product_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: product_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.product_id_seq OWNED BY public.product.id;


--
-- Name: strategy_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.strategy_runs (
    id bigint NOT NULL,
    run_id uuid DEFAULT gen_random_uuid(),
    symbol character varying(20) NOT NULL,
    start_time timestamp without time zone NOT NULL,
    end_time timestamp without time zone,
    initial_capital numeric(20,8) NOT NULL,
    final_capital numeric(20,8),
    strategy_type character varying(50) NOT NULL,
    strategy_params jsonb NOT NULL,
    total_trades integer DEFAULT 0,
    winning_trades integer DEFAULT 0,
    losing_trades integer DEFAULT 0,
    total_profit numeric(20,8) DEFAULT 0,
    max_drawdown numeric(5,2),
    sharpe_ratio numeric(10,4),
    win_rate numeric(5,2),
    environment character varying(20) DEFAULT 'backtest'::character varying,
    notes text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: strategy_runs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.strategy_runs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: strategy_runs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.strategy_runs_id_seq OWNED BY public.strategy_runs.id;


--
-- Name: symbol_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.symbol_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: symbol_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.symbol_id_seq OWNED BY public.symbol.id;


--
-- Name: time_series_metrics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.time_series_metrics (
    id integer NOT NULL,
    "timestamp" timestamp without time zone NOT NULL,
    symbol character varying(20) NOT NULL,
    metric_name character varying(50) NOT NULL,
    metric_value numeric(20,8) NOT NULL,
    labels jsonb,
    created_at timestamp without time zone DEFAULT now()
);


--
-- Name: time_series_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.time_series_metrics_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: time_series_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.time_series_metrics_id_seq OWNED BY public.time_series_metrics.id;


--
-- Name: trade_metrics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trade_metrics (
    id bigint NOT NULL,
    "timestamp" timestamp without time zone DEFAULT now() NOT NULL,
    symbol character varying(20) NOT NULL,
    side character varying(10) NOT NULL,
    price numeric(20,8) NOT NULL,
    quantity numeric(20,8) NOT NULL,
    profit numeric(20,8),
    portfolio_value numeric(20,8),
    strategy_params jsonb,
    indicators jsonb,
    decision_trigger character varying(100),
    signal_strength numeric(5,2),
    win_rate numeric(5,2),
    max_drawdown numeric(5,2),
    sharpe_ratio numeric(10,4),
    created_at timestamp without time zone DEFAULT now()
);


--
-- Name: trade_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.trade_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: trade_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.trade_metrics_id_seq OWNED BY public.trade_metrics.id;


--
-- Name: trader; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trader (
    id integer NOT NULL,
    symbol_id integer NOT NULL,
    init_qty double precision NOT NULL
);


--
-- Name: candles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.candles ALTER COLUMN id SET DEFAULT nextval('public.candles_id_seq'::regclass);


--
-- Name: exchange id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange ALTER COLUMN id SET DEFAULT nextval('public.exchange_id_seq'::regclass);


--
-- Name: order id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."order" ALTER COLUMN id SET DEFAULT nextval('public.order_id_seq'::regclass);


--
-- Name: order_side id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_side ALTER COLUMN id SET DEFAULT nextval('public.order_side_id_seq'::regclass);


--
-- Name: order_status id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_status ALTER COLUMN id SET DEFAULT nextval('public.order_status_id_seq'::regclass);


--
-- Name: order_type id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_type ALTER COLUMN id SET DEFAULT nextval('public.order_type_id_seq'::regclass);


--
-- Name: product id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product ALTER COLUMN id SET DEFAULT nextval('public.product_id_seq'::regclass);


--
-- Name: strategy_runs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.strategy_runs ALTER COLUMN id SET DEFAULT nextval('public.strategy_runs_id_seq'::regclass);


--
-- Name: symbol id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.symbol ALTER COLUMN id SET DEFAULT nextval('public.symbol_id_seq'::regclass);


--
-- Name: time_series_metrics id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_series_metrics ALTER COLUMN id SET DEFAULT nextval('public.time_series_metrics_id_seq'::regclass);


--
-- Name: trade_metrics id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trade_metrics ALTER COLUMN id SET DEFAULT nextval('public.trade_metrics_id_seq'::regclass);


--
-- Name: candles candles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.candles
    ADD CONSTRAINT candles_pkey PRIMARY KEY (id);


--
-- Name: candles candles_symbol_exchange_timeframe_timestamp_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.candles
    ADD CONSTRAINT candles_symbol_exchange_timeframe_timestamp_key UNIQUE (symbol, exchange, timeframe, "timestamp");


--
-- Name: exchange exchange_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange
    ADD CONSTRAINT exchange_pkey PRIMARY KEY (id);


--
-- Name: order order_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_pkey PRIMARY KEY (id);


--
-- Name: order_side order_side_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_side
    ADD CONSTRAINT order_side_pkey PRIMARY KEY (id);


--
-- Name: order_status order_status_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_status
    ADD CONSTRAINT order_status_pkey PRIMARY KEY (id);


--
-- Name: order_type order_type_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_type
    ADD CONSTRAINT order_type_pkey PRIMARY KEY (id);


--
-- Name: product product_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product
    ADD CONSTRAINT product_pkey PRIMARY KEY (id);


--
-- Name: strategy_runs strategy_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.strategy_runs
    ADD CONSTRAINT strategy_runs_pkey PRIMARY KEY (id);


--
-- Name: symbol symbol_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.symbol
    ADD CONSTRAINT symbol_pkey PRIMARY KEY (id);


--
-- Name: time_series_metrics time_series_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_series_metrics
    ADD CONSTRAINT time_series_metrics_pkey PRIMARY KEY (id);


--
-- Name: trade_metrics trade_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trade_metrics
    ADD CONSTRAINT trade_metrics_pkey PRIMARY KEY (id);


--
-- Name: trader trader_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trader
    ADD CONSTRAINT trader_pk PRIMARY KEY (id);


--
-- Name: idx_candles_symbol_exchange_timeframe_timestamp; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_candles_symbol_exchange_timeframe_timestamp ON public.candles USING btree (symbol, exchange, timeframe, "timestamp");


--
-- Name: idx_candles_timestamp; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_candles_timestamp ON public.candles USING btree ("timestamp");


--
-- Name: idx_strategy_runs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_strategy_runs_created_at ON public.strategy_runs USING btree (created_at);


--
-- Name: idx_strategy_runs_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_strategy_runs_run_id ON public.strategy_runs USING btree (run_id);


--
-- Name: idx_strategy_runs_symbol; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_strategy_runs_symbol ON public.strategy_runs USING btree (symbol);


--
-- Name: idx_time_series_metrics__created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_series_metrics__created_at ON public.time_series_metrics USING btree (created_at);


--
-- Name: idx_time_series_metrics__main; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_series_metrics__main ON public.time_series_metrics USING btree ("timestamp", symbol, metric_name);


--
-- Name: idx_trade_metrics__time__side; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trade_metrics__time__side ON public.trade_metrics USING btree ("timestamp", side);


--
-- Name: idx_trade_metrics__time__symbol; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trade_metrics__time__symbol ON public.trade_metrics USING btree ("timestamp", symbol);


--
-- Name: idx_trade_metrics_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trade_metrics_created_at ON public.trade_metrics USING btree (created_at);


--
-- Name: candles update_candles_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_candles_updated_at BEFORE UPDATE ON public.candles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: order order_exch_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_exch_id_fkey FOREIGN KEY (exch_id) REFERENCES public.exchange(id);


--
-- Name: order order_side_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_side_id_fkey FOREIGN KEY (side_id) REFERENCES public.order_side(id);


--
-- Name: order order_status_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_status_id_fkey FOREIGN KEY (status_id) REFERENCES public.order_status(id);


--
-- Name: order order_symbol_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_symbol_id_fkey FOREIGN KEY (symbol_id) REFERENCES public.symbol(id);


--
-- Name: order order_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_type_id_fkey FOREIGN KEY (type_id) REFERENCES public.order_type(id);


--
-- Name: symbol symbol_prod_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.symbol
    ADD CONSTRAINT symbol_prod_id_fkey FOREIGN KEY (prod_id) REFERENCES public.product(id);


--
-- Name: trader trader_symbol_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trader
    ADD CONSTRAINT trader_symbol_id_fk FOREIGN KEY (symbol_id) REFERENCES public.symbol(id);


--
-- PostgreSQL database dump complete
--


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
