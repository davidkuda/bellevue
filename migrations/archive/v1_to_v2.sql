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


-----------------------------------------------------------------------
-- bellevue_origin counts to consumptions:

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


-----------------------------------------------------------------------
-- bellevue_origin snacks_chf to consumptions:

/*
bellevue=# select * from bellevue_origins where snacks_chf > 0;

 id  | user_id | activity_date | breakfast_count | lunch_count | dinner_count | coffee_count | sauna_count | lecture_count | snacks_chf |    comment    | total_price |          created_at
-----+---------+---------------+-----------------+-------------+--------------+--------------+-------------+---------------+------------+---------------+-------------+-------------------------------
 100 |       1 | 2025-10-25    |               0 |           1 |            0 |            0 |           1 |             1 |        800 |               |        3850 | 2025-10-24 22:24:09.130195+00
 114 |      12 | 2025-10-16    |               1 |           1 |            1 |            3 |           1 |             1 |        900 |               |        6150 | 2025-11-12 15:11:54.645521+00
 125 |      13 | 2025-11-20    |               1 |           0 |            0 |            0 |           1 |             0 |        400 |               |        1950 | 2025-11-20 18:34:36.797032+00

bellevue=# select * from consumptions limit 5;

 id  | user_id | product_id | tax_id | pricecat_id | invoice_id |    date    | unit_price | quantity | total_price |          created_at
-----+---------+------------+--------+-------------+------------+------------+------------+----------+-------------+-------------------------------
 168 |       1 |         11 |      1 |           3 |            | 2025-12-28 |       1100 |        1 |        1100 | 2025-12-28 23:16:51.567654+00
 169 |       1 |         12 |      1 |           3 |            | 2025-12-28 |       1500 |        3 |        4500 | 2025-12-28 23:16:51.567654+00
 170 |       1 |          7 |      1 |           2 |            | 2025-12-28 |       1300 |        2 |        2600 | 2025-12-28 23:16:51.567654+00
 171 |       1 |          4 |      1 |           1 |            | 2025-12-28 |        100 |        2 |         200 | 2025-12-28 23:16:51.567654+00
 172 |       1 |          9 |      1 |           2 |            | 2025-12-28 |        750 |        1 |         750 | 2025-12-28 23:16:51.567654+00
(5 rows)

bellevue=# select * from products;

 id | financial_account_id | price_category_id | tax_id |     name      |   code    | pricing_mode | price |          valid_from           |          created_at           |          updated_at           | deleted_at
----+----------------------+-------------------+--------+---------------+-----------+--------------+-------+-------------------------------+-------------------------------+-------------------------------+------------
...
 15 |                    1 |                   |      1 | Snacks/Drinks | kiosk     | custom       |       | 2025-12-10 22:25:39.111992+00 | 2025-12-10 22:25:39.111992+00 | 2025-12-10 22:25:39.111992+00 |
...
*/

begin;

insert into bellevue.consumptions (
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
select
    user_id,
    (select id from products where code = 'kiosk') as product_id,
    (select id from taxes where code = 'B81') as tax_id,
    null as pricecat_id,
    null as invoice_id,
    activity_date,
    snacks_chf,
    1 as quantity,
    created_at
from bellevue_origins
where snacks_chf > 0;

commit;


-----------------------------------------------------------------------
-- bellevue_origin comments to comments:

begin;

INSERT INTO bellevue.comments (
    user_id,
    date,
    comment,
    created_at,
    updated_at
)
SELECT
    user_id,
    activity_date,
    string_agg(comment, E'; ' ORDER BY id) AS comment,
    MIN(created_at) AS created_at,
    MAX(created_at) AS updated_at
FROM bellevue.bellevue_origins
WHERE comment IS NOT NULL
  AND comment <> ''
GROUP BY user_id, activity_date;

commit;
