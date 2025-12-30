BEGIN;

insert into bellevue.users (
	id,
	first_name,
	last_name,
	email,
	method,
	hashed_password
)
select	id,
	first_name,
	last_name,
	email,
	'password',
	hashed_password
from	auth.users
where	id != 1;

COMMIT;


BEGIN;

-- TODO: snacks
-- TODO: comments

-----------------------------------------------------------------------
-- option 1: manual psql with variables

-- \set product 'breakfast'
-- \set price 1100

select
	o.user_id,
	(
		select id
		from products
		where code = :'product'
		and price_category_id = (
			select id
			from price_categories
			where name = 'reduced'
		)
	) as product_id,
	(select id from taxes where code = 'B81') as tax_id,
	(select id from price_categories where name = 'reduced') as pricecat_id,
	null as invoice_id,
	o.activity_date as date,
	:'price' as unit_price,
	o.breakfast_count as quantity,
	o.created_at
from	bellevue.bellevue_origins o
where	o.breakfast_count > 0;


-----------------------------------------------------------------------
-- option 2: cross join

-- v1:
INSERT INTO bellevue.consumptions (
    user_id,
    product_id,
    tax_id,
    pricecat_id,
    invoice_id,
    date,
    unit_price,
    quantity,
    created_at
)
SELECT
    o.user_id,
    p.id AS product_id,
    t.id AS tax_id,
    pc.id AS pricecat_id,
    NULL AS invoice_id,
    o.activity_date AS date,
    v.unit_price,
    v.quantity,
    o.created_at
FROM bellevue.bellevue_origins o
CROSS JOIN LATERAL (
    VALUES
        ('breakfast', 1100, o.breakfast_count, 'reduced'),
        ('lunch',     1100, o.lunch_count,     'reduced'),
        ('dinner',    1100, o.dinner_count,    'reduced'),
        ('coffee',     100, o.coffee_count,    'reduced'),
        ('sauna',      750, o.sauna_count,     'regular'),
        ('lecture',   1200, o.lecture_count,   'regular')
) AS v(product_code, unit_price, quantity, pricecat_name)
JOIN price_categories pc
  ON pc.name = v.pricecat_name
JOIN products p
  ON p.code = v.product_code
 AND p.price_category_id = pc.id
JOIN taxes t
  ON t.code = 'B81'
WHERE v.quantity > 0;


ROLLBACK;

BEGIN;

-- v2: with group by such that:
/*
 user_id | product_id | date        | unit_price | quantity
---------+------------+------------ +----------- +---------
       1 |          2 | 2025-11-25  |       1100 |       1
       1 |          2 | 2025-11-25  |       1100 |       1
*/
-- becomes:
/*
 user_id | product_id | date        | unit_price | quantity
---------+------------+------------ +----------- +---------
       1 |          2 | 2025-11-25  |       1100 |       2
*/

WITH raw AS (
    SELECT
        o.user_id,
        p.id AS product_id,
        t.id AS tax_id,
        pc.id AS pricecat_id,
        NULL::int AS invoice_id,
        o.activity_date AS date,
        v.unit_price,
        v.quantity,
        o.created_at
    FROM bellevue.bellevue_origins o
    CROSS JOIN LATERAL (
        VALUES
            ('breakfast', 1100, o.breakfast_count, 'reduced'),
            ('lunch',     1100, o.lunch_count,     'reduced'),
            ('dinner',    1100, o.dinner_count,    'regular'),
            ('coffee',     100, o.coffee_count,    'regular'),
            ('sauna',      750, o.sauna_count,     'regular'),
            ('lecture',   1200, o.lecture_count,   'regular')
    ) AS v(product_code, unit_price, quantity, pricecat_name)
    JOIN price_categories pc
      ON pc.name = v.pricecat_name
    JOIN products p
      ON p.code = v.product_code
     AND p.price_category_id = pc.id
    JOIN taxes t
      ON t.code = 'B81'
    WHERE v.quantity > 0
)
INSERT INTO bellevue.consumptions (
    user_id,
    product_id,
    tax_id,
    pricecat_id,
    invoice_id,
    date,
    unit_price,
    quantity,
    created_at
)
SELECT
    user_id,
    product_id,
    tax_id,
    pricecat_id,
    invoice_id,
    date,
    unit_price,
    SUM(quantity)        AS quantity,
    MAX(created_at)      AS created_at
FROM raw
GROUP BY
    user_id,
    product_id,
    tax_id,
    pricecat_id,
    invoice_id,
    date,
    unit_price;

COMMIT;
