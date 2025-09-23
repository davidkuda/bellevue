SET ROLE app;

BEGIN;

SET search_path TO bellevue;

insert into price_categories ( name )
values
	('reduced'),
	('regular'),
	('surplus');


insert into taxes (
	mwst_satz,
	code,
	name,
	description
) values (
	810,
	'B81',
	'Normalsatz',
	'Alle steuerbaren Leistungen, die nicht dem Sondersatz oder reduzierten Satz unterliegen, sind zum Normalsatz steuerbar.'
), (
	260,
	'B26',
	'Reduzierter Satz',
	'Die Steuer wird zum reduzierten Satz von 2,6 % erhoben auf dem Entgelt (und der Einfuhr) von Lieferungen wie Lebensmittel oder Bücher.'
), (
	380,
	'B38',
	'Sondersatz für Beherbergung',
	'Der Sondersatz auf Beherbergungsleistungen von 3,8 % findet Anwendung auf dem Gewähren von Unterkunft einschliesslich der allfälligen Abgabe eines Frühstücks, selbst wenn dieses separat in Rechnung gestellt wird.'
);


-- TODO: not sure if the tax_id is correct here. It is on product, too, maybe that's enough.
-- On the other hand, Lebensmittelertrag will always be 2.6% MWST.
insert into financial_accounts (
	code,
	name,
	friendly_name,
	tax_id
) values (
	3000,
	'Lebensmittelertrag',
	'Essen',
	(select id from bellevue.taxes where code = 'B81')
);

insert into products (
	name,
	financial_account_id
) values (
	'Breakfast',
	(select id from financial_accounts where code = 3000)
), (
	'Lunch',
	(select id from financial_accounts where code = 3000)
), (
	'Dinner',
	(select id from financial_accounts where code = 3000)
);

insert into prices (
	product_id,
	price_category_id,
	price,
	valid_from
) values (
	(select id from products where name = 'Breakfast'),
	(select id from price_categories where name = 'regular'),
	800,
	now()
);

COMMIT;

-- last migration step is updating public.schema_migrations.
-- app has no permissions. so use dev.
SET ROLE dev;
