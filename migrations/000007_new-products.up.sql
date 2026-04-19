set role developer;

begin;

insert into products (
	name,
	code,
	price,
	financial_account_id,
	price_category_id,
	tax_id
) values (
	'Kurs (Wochenende)',
	'course/weekend',
	4000,
	(select id from financial_accounts where code = 3400),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
), (
	'Kurs (8. Karmapa)',
	'course/8thkarmapa',
	7000,
	(select id from financial_accounts where code = 3400),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B81')
);

insert into products (
	name,
	code,
	pricing_mode,
	financial_account_id,
	tax_id
) values (
	'Essen (Freibetrag)',
	'food',
	'custom',
	(select id from financial_accounts where code = 3000),
	(select id from taxes where code = 'B81')
);

delete from product_form_order;

insert into product_form_order (code, sort_order)
VALUES
  ('breakfast', 10),
  ('lunch', 20),
  ('dinner', 30),
  ('coffee', 40),
  ('sauna', 50),
  ('lecture', 60),
  ('course/weekend', 63),
  ('course/8thkarmapa', 67),
  ('food', 70),
  ('kiosk', 80),
  ('donations', 90)
;

commit;
