-- prerequisite: createdb bellevue

BEGIN;

create schema bellevue;

--------------------------------------------------------------------
-- Roles: Groups: Developer ----------------------------------------
-- A developer can CREATE ON SCHEMA, an app can only USAGE. --------

create role developer with nologin;

grant create on schema bellevue to developer;

grant select, insert, update, delete 
on all tables in schema bellevue
TO developer;


--------------------------------------------------------------------
-- Roles: Groups: App ----------------------------------------------

create role application with nologin;

grant usage on schema bellevue to application;

grant select, insert, update, delete 
on all tables in schema bellevue
to application;


--------------------------------------------------------------------
-- Roles: Users: (with login) --------------------------------------

CREATE ROLE dev WITH login PASSWORD 'pa55word' INHERIT;
CREATE ROLE app WITH login PASSWORD 'pa55word' INHERIT;
GRANT developer TO dev;
GRANT application TO app;

ALTER SCHEMA bellevue OWNER TO developer;

COMMIT;
