BEGIN;

SET ROLE dev;


create schema if not exists bellevue;

SET ROLE dev;

----------------------------------------------------------------------------------
-- Bellevue Origins: (original design) -------------------------------------------

-- NOTE: Prices are stored in Rappen, not as fraction of CHF.
-- i.e., 108.00 CHF are represented as integer 10800.
-- (in Go, post-process to render: CHF => float64(total_price) / 100.0)
-- Alternatively, you could use NUMERIC(2, 12), but needs more
-- go code to work with and slows down postgres.
-- https://www.postgresql.org/docs/current/datatype-numeric.html#DATATYPE-NUMERIC-DECIMAL

create table bellevue.bellevue_origins (
	id              SERIAL primary key,
	user_id         INT references auth.users(id),
	activity_date   DATE not null,
	breakfast_count INT default 0 not null,
	lunch_count     INT default 0 not null,
	dinner_count    INT default 0 not null,
	coffee_count    INT default 0 not null,
	sauna_count     INT default 0 not null,
	lecture_count   INT default 0 not null,
	snacks_chf      INT default 0 not null,
	comment         TEXT,
	total_price     INT not null,
	created_at      TIMESTAMPTZ default now() not null
);

insert into bellevue.bellevue_origins (
	id,
	user_id,
	activity_date,
	breakfast_count,
	lunch_count,
	dinner_count,
	coffee_count,
	sauna_count,
	lecture_count,
	snacks_chf,
	comment,
	total_price,
	created_at
)
select
	id,
	user_id,
	activity_date,
	breakfast_count,
	lunch_count,
	dinner_count,
	coffee_count,
	sauna_count,
	lecture_count,
	snacks_chf,
	comment,
	total_price,
	created_at
from website.bellevue_activities;

select setval(
    pg_get_serial_sequence('bellevue.bellevue_origins', 'id'),
    coalesce(max(id), 1)
) from bellevue.bellevue_origins;



----------------------------------------------------------------------------------
-- Activities By Month: ----------------------------------------------------------

CREATE TYPE invoice_state AS ENUM (
  'open',
  'paid'
);

DROP TABLE if exists bellevue.invoices;
CREATE TABLE bellevue.invoices (
	id                 SERIAL primary key,
	user_id            INT not null references auth.users(id),
	period             DATE not null,
	total_price_rappen INT not null default 0,
	total_eating       INT not null default 0,
	total_coffee       INT not null default 0,
	total_lecture      INT not null default 0,
	total_sauna        INT not null default 0,
	total_kiosk        INT not null default 0,
	state              INVOICE_STATE not null default 'open',
	period_yyyymm      INT generated always as (
		(EXTRACT(YEAR FROM period)::int * 100)
		+ EXTRACT(MONTH FROM period)::int
	) STORED,
	CHECK(period = date_trunc('month', period)::date),
	CHECK(total_price_rappen = total_eating + total_coffee + total_lecture + total_sauna + total_kiosk),
	UNIQUE(user_id, period)
);


-- Migration:
WITH
	PRICES AS (
		SELECT
			800::int AS breakfast_rappen, -- CHF  8.00
			1100::int AS lunch_rappen, -- CHF 11.00
			1100::int AS dinner_rappen, -- CHF 11.00
			100::int AS coffee_rappen, -- CHF  1.00
			1200::int AS lecture_rappen, -- CHF 12.00
			750::int AS sauna_rappen -- CHF  7.50
	),
	SRC AS (
		SELECT
			o.user_id,
			date_trunc('month', o.activity_date)::date AS period,
			(o.breakfast_count * p.breakfast_rappen + o.lunch_count * p.lunch_rappen + o.dinner_count * p.dinner_rappen) AS eating_rappen,
			(o.coffee_count * p.coffee_rappen) AS coffee_rappen,
			(o.lecture_count * p.lecture_rappen) AS lecture_rappen,
			(o.sauna_count * p.sauna_rappen) AS sauna_rappen,
			o.snacks_chf
		FROM
			bellevue.bellevue_origins o
			CROSS JOIN PRICES p
	),
	AGG AS (
		SELECT
			user_id,
			period,
			SUM(eating_rappen) AS total_eating,
			SUM(coffee_rappen) AS total_coffee,
			SUM(lecture_rappen) AS total_lecture,
			SUM(sauna_rappen) AS total_sauna,
			SUM(snacks_chf) AS total_kiosk
		FROM
			SRC
		GROUP BY
			user_id,
			period
	)
INSERT INTO
	bellevue.invoices (user_id, period, total_price_rappen, total_eating, total_coffee, total_lecture, total_sauna, total_kiosk, state)
SELECT
	user_id,
	period,
	(total_eating + total_coffee + total_lecture + total_sauna + total_kiosk) AS total_price_rappen,
	total_eating,
	total_coffee,
	total_lecture,
	total_sauna,
	total_kiosk,
	'open'::invoice_state
FROM
	AGG;


----------------------------------------------------------------------------------
-- Update Permissions: -----------------------------------------------------------

GRANT USAGE ON SCHEMA bellevue TO kuda_ai;

GRANT SELECT, INSERT, UPDATE, DELETE
ON ALL TABLES IN SCHEMA bellevue
TO app;

GRANT USAGE, SELECT, UPDATE
ON ALL SEQUENCES IN SCHEMA bellevue
TO app;

ALTER TABLE bellevue.bellevue_origins OWNER TO dev;
ALTER TABLE bellevue.invoices OWNER TO dev;

COMMIT;
