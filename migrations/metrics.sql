--
-- PostgreSQL database dump
--

-- Dumped from database version 14.18 (Debian 14.18-1.pgdg120+1)
-- Dumped by pg_dump version 17.5

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
-- Name: candles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.candles ALTER COLUMN id SET DEFAULT nextval('public.candles_id_seq'::regclass);


--
-- Name: strategy_runs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.strategy_runs ALTER COLUMN id SET DEFAULT nextval('public.strategy_runs_id_seq'::regclass);


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
-- Name: strategy_runs strategy_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.strategy_runs
    ADD CONSTRAINT strategy_runs_pkey PRIMARY KEY (id);


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
-- PostgreSQL database dump complete
--

