BEGIN;

insert into forms (name) values ('NewActivity');

insert into form_fields (
	form_id,
	field_type_id,
	name,
	label,
	position,
	required
) values (
	1,
	(select id from form_field_types where name = 'date'),
	'date',
	'Date',
	0,
	true,
), (
	1,
	(select id from form_field_types where name = 'counter'),
	'date',
	'Date',
	0,
	true,
;

insert into form_field_types (name)
values
	('text'),
	('textarea'),
	('number'),
	('date'),
	('counter'),
	('select'),
	('checkbox')
;

COMMIT;
