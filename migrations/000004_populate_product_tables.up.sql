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
), (
	0,
	'0',
	'Nicht Mehrwertssteuerpflichtig',
	'z.B. Spenden'
);


-- TODO: not sure if the tax_id is correct here. It is on product, too, maybe that's enough.
-- On the other hand, Lebensmittelertrag will always be 2.6% MWST.
insert into financial_accounts (
	code,
	name,
	description,
	tax_id
) values (
	3000,
	'Restauration/Verpflegung',
	'Essen',
	(select id from bellevue.taxes where code = 'B81')
), (
	3002,
	'Restauration/Kaffee',
	'Kaffee',
	(select id from bellevue.taxes where code = 'B81')
), (
	3201,
	'Restauration/Kiosk',
	'Bier, Chips, Schokolade',
	(select id from bellevue.taxes where code = 'B81')
), (
	3400,
	'Veranstaltungen',
	'Vorträge',
	(select id from bellevue.taxes where code = 'B81')
), (
	3109,
	'Einnahmen diverse Nebenleistungen',
	'Sauna',
	(select id from bellevue.taxes where code = 'B81')
);

-- reduced products:
insert into products (
	name,
	price,
	financial_account_id,
	price_category_id,
	tax_id
) values (
	'Breakfast',
	800,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'reduced'),
	(select id from taxes where code = 'B81')
), (
	'Lunch',
	1100,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'reduced'),
	(select id from taxes where code = 'B81')
), (
	'Dinner',
	1100,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'reduced'),
	(select id from taxes where code = 'B81')
), (
	'Coffee',
	100,
	(select id from financial_accounts where code = 3002),
	(select id from price_categories where name = 'reduced'),
	(select id from taxes where code = 'B81')
)
;

-- regular products:
insert into products (
	name,
	price,
	financial_account_id,
	price_category_id,
	tax_id
) values (
	'Breakfast',
	900,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Lunch',
	1300,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Dinner',
	1300,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Coffee',
	200,
	(select id from financial_accounts where code = 3002),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Sauna',
	750,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Lectures',
	1200,
	(select id from financial_accounts where code = 3400),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
)
;

-- surplus products:
insert into products (
	name,
	price,
	financial_account_id,
	price_category_id,
	tax_id
) values (
	'Breakfast',
	1100,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'surplus'),
	(select id from taxes where code = 'B81')
), (
	'Lunch',
	1500,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'surplus'),
	(select id from taxes where code = 'B81')
), (
	'Dinner',
	1500,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'surplus'),
	(select id from taxes where code = 'B81')
), (
	'Coffee',
	300,
	(select id from financial_accounts where code = 3002),
	(select id from price_categories where name = 'surplus'),
	(select id from taxes where code = 'B81')
)
;

-- custom priced products:
insert into products (
	name,
	pricing_mode,
	financial_account_id,
	tax_id
) values (
	'Snacks/Drinks',
	'custom',
	(select id from financial_accounts where code = 3000),
	(select id from taxes where code = 'B81')
), (
	'Donations',
	'custom',
	(select id from financial_accounts where code = 3000),
	(select id from taxes where code = 'B81')
)
;


COMMIT;
