-- prerequisite: createdb bellevue

-- developers inherit from group developer.
-- go web app runs with bellevue which inherits from application.
-- migrations should be run with developer.

BEGIN;

-- schema bellevue will be owned by developer;
create schema bellevue;

ALTER DATABASE bellevue
SET search_path = bellevue, public;


--------------------------------------------------------------------
-- Roles: Groups: Developer ----------------------------------------
-- A developer can CREATE ON SCHEMA, an app can only USAGE. --------

create role developer with nologin;

alter schema bellevue owner to developer;

grant create on schema bellevue to developer;

grant select, insert, update, delete 
on all tables in schema bellevue
TO developer;


--------------------------------------------------------------------
-- Roles: Groups: application --------------------------------------

create role application with nologin;

grant usage on schema bellevue to application;

grant select, insert, update, delete 
on all tables in schema bellevue
to application;


--------------------------------------------------------------------
-- Default privileges: developer => application: -------------------

-- every time developer creates a new table, application will
-- receive a grant as specified in:
ALTER DEFAULT PRIVILEGES
FOR ROLE developer
IN SCHEMA bellevue
GRANT SELECT, INSERT, UPDATE, DELETE
ON TABLES
TO application;

-- also consider sequences:
-- USAGE: allows nextval(), currval(), lastval()
-- SELECT: allows currval() and reading the sequence via SELECT directly.
-- UPDATE: allows nextval() and setval() – modifying the sequence’s current value.
ALTER DEFAULT PRIVILEGES
FOR ROLE developer
IN SCHEMA bellevue
GRANT USAGE, SELECT, UPDATE
ON SEQUENCES
TO application;


--------------------------------------------------------------------
-- Roles: Users: (with login) --------------------------------------

CREATE ROLE dev WITH login PASSWORD 'pa55word' INHERIT;
GRANT developer TO dev;

CREATE ROLE bellevue WITH login PASSWORD 'pa55word' INHERIT;
GRANT application TO bellevue;

COMMIT;
