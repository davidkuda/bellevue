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
	code,
	price,
	financial_account_id,
	price_category_id,
	tax_id
) values (
	'Breakfast',
	'breakfast',
	800,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'reduced'),
	(select id from taxes where code = 'B81')
), (
	'Lunch',
	'lunch',
	1100,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'reduced'),
	(select id from taxes where code = 'B81')
), (
	'Dinner',
	'dinner',
	1100,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'reduced'),
	(select id from taxes where code = 'B81')
), (
	'Coffee',
	'coffee',
	100,
	(select id from financial_accounts where code = 3002),
	(select id from price_categories where name = 'reduced'),
	(select id from taxes where code = 'B81')
)
;

-- regular products:
insert into products (
	name,
	code,
	price,
	financial_account_id,
	price_category_id,
	tax_id
) values (
	'Breakfast',
	'breakfast',
	900,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Lunch',
	'lunch',
	1300,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Dinner',
	'dinner',
	1300,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Coffee',
	'coffee',
	200,
	(select id from financial_accounts where code = 3002),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Sauna',
	'sauna',
	750,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Lectures',
	'lecture',
	1200,
	(select id from financial_accounts where code = 3400),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
)
;

-- surplus products:
insert into products (
	name,
	code,
	price,
	financial_account_id,
	price_category_id,
	tax_id
) values (
	'Breakfast',
	'breakfast',
	1100,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'surplus'),
	(select id from taxes where code = 'B81')
), (
	'Lunch',
	'lunch',
	1500,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'surplus'),
	(select id from taxes where code = 'B81')
), (
	'Dinner',
	'dinner',
	1500,
	(select id from financial_accounts where code = 3000),
	(select id from price_categories where name = 'surplus'),
	(select id from taxes where code = 'B81')
), (
	'Coffee',
	'coffee',
	300,
	(select id from financial_accounts where code = 3002),
	(select id from price_categories where name = 'surplus'),
	(select id from taxes where code = 'B81')
)
;

-- custom priced products:
insert into products (
	name,
	code,
	pricing_mode,
	financial_account_id,
	tax_id
) values (
	'Snacks/Drinks',
	'kiosk',
	'custom',
	(select id from financial_accounts where code = 3000),
	(select id from taxes where code = 'B81')
), (
	'Donations',
	'donations',
	'custom',
	(select id from financial_accounts where code = 3000),
	(select id from taxes where code = 'B81')
)
;


insert into product_form_order (code, sort_order)
VALUES
  ('breakfast', 10),
  ('lunch', 20),
  ('dinner', 30),
  ('coffee', 40),
  ('sauna', 50),
  ('lecture', 60),
  ('kiosk', 70),
  ('donations', 80)
;

COMMIT;
