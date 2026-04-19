set role developer;

begin;

delete from product_form_order;

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

delete from products
where code = 'food';

delete from products
where code = 'course/weekend';

delete from products
where code = 'course/8thkarmapa';

commit;
