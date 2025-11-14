BEGIN;

----------------------------------------------------------------------------------
-- models for forms: -------------------------------------------------------------

-- new activity form
-- Date    => ActivityDate
-- Counter => Breakfast
-- Counter => Lunch
-- Counter => Dinner
-- Counter => Coffee
-- Counter => Lecture
-- Number  => Kiosk
-- Text    => Comment
-- each needs Label and name-prop

-- Prices may vary (e.g. reduced 11, regular 13, surplus 15)
-- maybe fetch price info from the products/price_category/prices table?
-- or define counter with select options?

-- e.g. NewActivityForm
-- maybe columns:
-- - is_active bool default true not null,
-- - version   int  default 1    not null,
create table forms (
	id                    SERIAL8 primary key,
	name                  STRING not null,
	created_at            TIMESTAMPTZ default now() not null
);

-- maybe:
	-- help_text     STRING,
	-- min_int       INT, e.g. 0
	-- max_int       INT,
	-- pattern       STRING,
	-- default_value ???
create table form_fields (
	id            SERIAL primary key,
	form_id       INT8 not null references forms (id) on delete cascade,
	name          STRING not null,
	label         STRING not null,
	field_type_id INT not null references field_types (id),
	position      INT default 0 not null,
	required      BOOL default false not null,
	unique (form_id, name)
);


-- "text", "textarea", "number", "date", "counter", "select", "checkbox"
create table form_field_types (
	id   smallserial primary key,
	name text unique not null
);

-- Discrete options for select/radio/checkbox,
-- e.g. [ ] male [ ] female [ ] other
create table form_field_options (
	id         SERIAL8 primary key,
	field_id   INT8 not null references form_fields (id) on delete cascade,
	value      STRING not null,
	label      STRING not null,
	position   INT8 default 0 not null,
	unique (field_id, value)
);

-- Submissions
create table form_submissions (
	id           SERIAL8 primary key,
	form_id      INT8 not null references forms (id) on delete restrict,
	submitted_at TIMESTAMPTZ default now() not null,
	submitted_by UUID,
	unique (form_id, id)
);

-- Submitted values (EAV)
create table form_submission_values (
	submission_id INT8 not null references form_submissions (id) on delete cascade,
	field_id      INT8 not null references form_fields (id) on delete restrict,
	value_text    STRING,
	primary key (submission_id, field_id)
);

----------------------------------------------------------------------------------
-- Update Permissions: -----------------------------------------------------------

GRANT SELECT, INSERT, UPDATE, DELETE
ON ALL TABLES IN SCHEMA bellevue
TO app;

GRANT USAGE, SELECT, UPDATE
ON ALL SEQUENCES IN SCHEMA bellevue
TO app;

COMMIT;

-- last migration step is updating public.schema_migrations.
-- app has no permissions. so use dev.
SET ROLE dev;
