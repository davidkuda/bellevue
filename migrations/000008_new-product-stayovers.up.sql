set role developer;

begin;

insert into financial_accounts (
	tax_id,
	code,
	name,
	view_name
) values (
	(select id from taxes where mwst_satz = 380),
	3104,
	'Übernachtung Mehrbettzimmer',
	'Übernachtung'
);

insert into products (
	name,
	code,
	price,
	financial_account_id,
	price_category_id,
	tax_id
) values (
	'Übernachtung (Karmayoga)',
	'stay/karmayoga',
	1500,
	(select id from financial_accounts where code = 3104),
	(select id from price_categories where name = 'regular'),
	(select id from taxes where code = 'B38')
);

delete from product_form_order;

insert into product_form_order (code, sort_order)
VALUES
  ('stay/karmayoga', 5),
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
