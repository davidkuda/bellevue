these notes were taken following the excellent talk `for your eyes only` by Ryan Booz. [Watch the talk on YouTube.](https://youtu.be/mtPM3iZFE04?si=NN7VN_XUaLf7OFkv)


- avoid using TRUST most of the times. Only exception maybe: local dev postgres cluster.
- use `scram-sha-256` for password authentication, not `md5`


Roles:
- own database objects (tables, functions, etc)
- cluster-level privileges (attributes)
- granted privileges to database objects
- can possibly grant privileges to other roles

Role Attributes:
- predefined settings that can be enabled / disabled for a given role
- cluster-level (non-database) privileges
- map to columns in `pg_catalog.pg_roles`

PostgreSQL 15 Attributes:
- LOGIN
- PASSWORD
- SUPERUSER
- INHERIT
- CREATEROLE
- CREATEDB
less common:
- BYPASSRLS
- CONNECTION LIMIT
- REPLICATION LOGIN

good to know: Unless otherwise set, new roles can inherit privileges from other roles and have unlimited connections.

SUPERUSER:
- very powerful
- Bypass all security checks except LOGIN.
- wipe anything out you want
- try avoid using it
- treat with care like `root` on Linux
- cloud providers usually don't provide superuser access.

Superuser-like:
- Normally, you want to operate with a superuser-like role.
- `CREATEROLE` and `CREATEDB`
- allow user management and database ownership

PUBLIC Role:
- all roles are granted implicit membership to PUBLIC
- the public role cannot be deleted
- granted CONNECT USAGE TEMPORARY and EXECUTE by default
- CREATE granted <= PG14, not from PG15

Security Best Practice for PUBLIC:
- revoke all privileges on the public schema from the PUBLIC role
- revoke all database privileges from the PUBLIC role (maybe)

```sql
revoke all on schema public from public;
revoke all on database db_name from public;
```

PRIVILEGES:
- set of access rights to databases and database objects
- can be GRANT or REVOKE.
- explicit GRANT or REVOKE only impacts existing objects.

List:
- SELECT
- CREATE
- INSERT
- UPDATE
- DELETE
- TRUNCATE
- REFERENCES
- TRIGGER
- CONNECT
- TEMPORARY
- EXECUTE
- USAGE
- SET
- ALTER SYSTEM

```sql
grant create on database appdb to admin1;

grant usage, create in schema app to dev1;

grant select, insert, update
  on all tables in schema app to jr_dev;
```

CREATE has different meanings depending on the context.
- CREATE ON DATABASE?
- CREATE ON SCHEMA?


privileges docs: https://www.postgresql.org/docs/current/ddl-priv.html


Inheritance:
- used for groups

```sql
create role sr_dev with login password='abc' INHERIT;
create role reportuser with login password='abc' INHERIT;

create role admin with nologin noinherit;
create role ropriv with nologin noinherit;

grant insert, update, delete
on all tables
in schema app
to admin;

grant select on all tables in schema app to ropriv;

grant admin, ropriv to sr_dev;
grant ropriv to reportuser;
```


Object Ownership:
- object creator = owner
- principle of least privilege: unless specifically granted ahead of time, objects are owned and "accessible" by the creator/superuser only.


The idea is to use a group role to create, then every role that inherits has access.

Another way is to use:

Default Privileges:

```sql
alter default privileges
grant select on tables to public;
```


Providing Object Access:
1. Option: Owner: Explicitly GRANT access after object creation
2. Option: Owner: ALTER DEFAULT PRIVILEGES
3. Option: SET ROLE to app role (group) before creation with correct default privileges
4. Option: use predefined roles (PG14+) `pg_read_all_data` or `pg_write_all_data`


```sql
create role developer with nologin;

set role none;

grant select, insert, update, delete
  on all tables in schema public to developer;

grant create on schema public to developer;

create role dev1 with login password 'abc' inherit;
create role dev2 with login password 'abc' inherit;

grant developer to dev1;
grant developer to dev2;

-- login as dev1, create a new table:
SET ROLE dev1;
create table new_table(col1 text);

-- \d or
select schemaname, tablename, tableowner
from pg_catalog.pg_tables
where tablename = 'new_table';

set role dev2;
-- does not work!
alter table new_table add column col2 int;

-- instead, when creating new_table, use the group
set role developer;
create table new_table(col1 text);

-- if you want to automatically grant other groups
-- permissions, use default privileges

create role application with nologin;

-- this will be considered for objects that I own,
-- but only in the future, not past:
alter default privileges
grant select, insert, update, delete
  on all tables
to application;
```
