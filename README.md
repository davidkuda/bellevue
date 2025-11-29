# Bellevue Activities

## Local Development Setup

## Postgres DB

### Local setup

1. Install `sudo apt install postgresql`
2. Add yourself as user and also create one called bellevue

If your postgres server runs on your machine:

```sh
sudo -u postgres \
  createuser \
    --superuser \
    $USER

# this is the same as running:
# psql -U postgres -c "CREATE ROLE \"$USER\" WITH SUPERUSER LOGIN;"

# once you have your own user in the DB, you no longer need
# to use `sudo -u postgres`.

createuser --superuser bellevue

createdb bellevue
psql -d bellevue
```

If your postgres server runs as a docker container: You specify the user name and the database name with environment variables.

- `createuser bellevue == POSTGRES_USER: bellevue`
- `createdb bellevue   == POSTGRESDB: bellevue`


3. then in postgres, alter the bellevue password
`ALTER USER bellevue WITH PASSWORD 'pa55word';`
CTRL + d to quite the psql program.

4. Run the migration scripts.

To run a script, use the command:

```sh
psql -U username -d myDataBase -a -f myInsertFile

# Options:
# -d / --dbname    bellevue
# -a / --echo-all
# -f / --file      migrations/file.sql
# -h / --host      localhost
# -p / --port      5433 (5432 is default; optional, useful if port is not 5432)

# i.e.:
psql \
  -d bellevue \
  -a \
  -f migrations/000001_create_roles.up.sql
```


### Docker Setup
In the folder `db` there is a docker compose script

1. Change into the db folder
2. Run the docker compose with `docker compose up -d`
3. Copy the migration files to the container with

```sh
docker cp migrations/ postgres:/migrations/
```

4. Jump into a bash terminal on the container with:

```sh
`docker exec -it \
-e PGPASSWORD="pa55word" \
-e PGDATABASE="bellevue" \
-e PGUSER="bellevue" \
postgres bash
```

5. Run the initialisation scripts that you copied earlier
```sh
psql -f migrations/000001_create_roles.up.sql
```


If you have `psql` as a command on your system and docker is running in a container, you can export env vars and access psql in the container directly.

```sh
export PGDATABASE=bellevue
export PGUSER=bellevue
export PGPASSWORD=pa55sword
export PGPORT=5432

psql -X -q -c '\conninfo'
```

### Docker compose

## npm
Install npm dependencies with `npm install` in the root of the directory and then `npm install esbuild`

```
CREATE ROLE
GRANT developer TO dev;
GRANT ROLE
CREATE ROLE kuda_ai WITH login PASSWORD 'pa55word' INHERIT;
CREATE ROLE
GRANT app TO kuda_ai;
GRANT ROLE
ALTER DATABASE kuda_ai OWNER TO dev;
psql:migrations/000001_create_roles.up.sql:40: ERROR:  database "kuda_ai" does not exist
ALTER SCHEMA bellevue OWNER TO dev;
ALTER SCHEMA
```

psql -d bellevue -a -f migrations/000003_data_model_products.up.sql 
BEGIN;
BEGIN
----------------------------------------------------------------------------------
-- user authentication and sessions (cookies) relations:
create table bellevue.users (
                id              SERIAL primary key,
            first_name      TEXT not null,
            last_name       TEXT not null,
            email           TEXT not null unique,
            method          TEXT not null,
                -- email / password logins:
            hashed_password CHAR(60),
                -- openid connect logins:
            sub             TEXT unique,
                created_at      timestamptz NOT NULL DEFAULT now(),
        CHECK (
                case
                        when method = 'password'
                        then hashed_password is not null
                        when method = 'openidconnect'
                        then sub is not null
                        else false -- reject unknown methods
                end
        )
);
CREATE TABLE
-- https://pkg.go.dev/github.com/alexedwards/scs/postgresstore#section-readme
-- https://pkg.go.dev/github.com/alexedwards/scs/pgxstore#section-readme
create table bellevue.sessions (
        token  text primary key,
        data   bytea not null,
        expiry timestamptz not null
);
CREATE TABLE
create index sessions_expiry_idx
on sessions (expiry);
psql:migrations/000003_data_model_products.up.sql:42: ERROR:  relation "sessions" does not exist
alter table bellevue.users
owner to dev;
psql:migrations/000003_data_model_products.up.sql:45: ERROR:  current transaction is aborted, commands ignored until end of transaction block
alter table bellevue.sessions
owner to dev;
psql:migrations/000003_data_model_products.up.sql:48: ERROR:  current transaction is aborted, commands ignored until end of transaction block
grant select, insert, update, delete
on table users, sessions
to app;
psql:migrations/000003_data_model_products.up.sql:52: ERROR:  current transaction is aborted, commands ignored until end of transaction block
grant usage, select, update
on sequence users_id_seq
to app;
psql:migrations/000003_data_model_products.up.sql:56: ERROR:  current transaction is aborted, commands ignored until end of transaction block
----------------------------------------------------------------------------------
-- create a generic model for products and prices --------------------------------
--
-- NOTE:
-- ANSISQL is current_timestamp, PostgreSQL is now() (but you can use both in PG).
-- It refers to the start of the transaction, not point of execution.
--
-- NOTE:
-- use `id INT generated by default as identity primary key,` instead
-- of `id serial`, see https://stackoverflow.com/a/55300741/14501123
-- names: reduced, regular, surplus
create table bellevue.price_categories (
        id          INT
                    generated by default as identity
                    primary key,
        name        TEXT unique not null,
        created_at TIMESTAMPTZ default now() not null,
        updated_at TIMESTAMPTZ default now() not null,
        deleted_at TIMESTAMPTZ
);
psql:migrations/000003_data_model_products.up.sql:80: ERROR:  current transaction is aborted, commands ignored until end of transaction block
-- MWST / VAT
-- friendly-name: what users want on their invoice.
create table bellevue.taxes (
        id            INT
                      generated by default as identity
                      primary key,
        mwst_satz     SMALLINT not null
                      CHECK (mwst_satz between 0 and 10000), -- 8.1% => 810
        code          TEXT unique, -- z.B. B81
        name          TEXT,
        friendly_name TEXT,
        description   TEXT,
        created_at TIMESTAMPTZ default now() not null,
        updated_at TIMESTAMPTZ default now() not null,
        deleted_at TIMESTAMPTZ
);
psql:migrations/000003_data_model_products.up.sql:98: ERROR:  current transaction is aborted, commands ignored until end of transaction block
-- name: Lebensmittelertrag.
-- friendly_name: Essen.
create table bellevue.financial_accounts (
        id          INT
                    generated by default as identity
                    primary key,
        tax_id      INT not null
                    references bellevue.taxes(id),
        code        INT unique, -- 3000, 3xxx
        name        TEXT not null, -- e.g. "Einnahmen diverse Nebenleistungen"
        description TEXT,
        created_at TIMESTAMPTZ default now() not null,
        updated_at TIMESTAMPTZ default now() not null,
        deleted_at TIMESTAMPTZ
);
psql:migrations/000003_data_model_products.up.sql:116: ERROR:  current transaction is aborted, commands ignored until end of transaction block
-- valid_from: do you need valid_to? I don't think so, just use
-- whatever is max(valid_to) that is smaller than today.
-- or skip entirely and introduce on requested feature?
create table bellevue.products (
        id                   INT
                             generated by default as identity
                             primary key,
        financial_account_id INT
                             not null
                             references bellevue.financial_accounts(id),
        price_category_id    INT
                             references bellevue.price_categories(id),
                             check (
                               (pricing_mode = 'fixed'  and price_category_id is not null) or
                               (pricing_mode = 'custom' and price_category_id is null)
                             ),
        tax_id               INT
                             not null
                             references bellevue.taxes(id),
        name                 TEXT not null, -- e.g. Breakfast, may be translated
        code                 TEXT not null, -- e.g. breakfasts, used internally, for instance for forms
        pricing_mode         TEXT
                             not null
                             default 'fixed'
                         check (pricing_mode in ('fixed', 'custom')),
        price                INT -- 11.00 CHF => 1100
                             check (
                               (pricing_mode = 'fixed'  and price is not null) or
                               (pricing_mode = 'custom' and price is null)
                            ),
        valid_from           TIMESTAMPTZ not null default now(),
        unique (name, price_category_id, valid_from),
        created_at TIMESTAMPTZ default now() not null,
        updated_at TIMESTAMPTZ default now() not null,
        deleted_at TIMESTAMPTZ
);
psql:migrations/000003_data_model_products.up.sql:157: ERROR:  current transaction is aborted, commands ignored until end of transaction block
create table bellevue.product_form_order (
        code       text primary key,
        sort_order int not null
);
psql:migrations/000003_data_model_products.up.sql:162: ERROR:  current transaction is aborted, commands ignored until end of transaction block
create table bellevue.comments (
        user_id    INT  not null references bellevue.users(id),
        date       DATE not null,
        comment    TEXT,
        created_at TIMESTAMPTZ default now() not null,
        updated_at TIMESTAMPTZ default now() not null,
        primary key (user_id, date)
);
psql:migrations/000003_data_model_products.up.sql:172: ERROR:  current transaction is aborted, commands ignored until end of transaction block
-- TODO: how should I invoice consumptions?
create table bellevue.invoices_v2 (
        id          serial primary key,
        user_id     int not null references bellevue.users(id),
        period_from date not null,
        period_to   date not null,
        status      text not null default 'draft'
                    check (status in ('draft', 'sent', 'paid', 'cancelled')),
        created_at  timestamptz not null default now(),
        updated_at  timestamptz not null default now(),
        unique (user_id, period_from, period_to)
);
psql:migrations/000003_data_model_products.up.sql:188: ERROR:  current transaction is aborted, commands ignored until end of transaction block
-- TODO: should I keep the following column column?
--      mwst_price INT not null, -- unmutable fact
-- NOTE: if invoice_id is null, editable
create table bellevue.consumptions (
        id          BIGINT generated by default as identity primary key,
        user_id     INT not null
                    references bellevue.users(id),
        product_id  INT not null
                    references products(id),
        tax_id      INT not null
                    references taxes(id),
        pricecat_id INT
                    references price_categories(id),
        invoice_id  int
                    references invoices_v2(id),
        date        DATE,
        unit_price  INT not null, -- unmutable fact after invoice_id is not null
        quantity    int not null
                    check (quantity > 0),
        total_price int generated always as (quantity * unit_price) stored,
        created_at TIMESTAMPTZ default now() not null
);
psql:migrations/000003_data_model_products.up.sql:214: ERROR:  current transaction is aborted, commands ignored until end of transaction block
create index on bellevue.consumptions (user_id);
psql:migrations/000003_data_model_products.up.sql:216: ERROR:  current transaction is aborted, commands ignored until end of transaction block
create index on bellevue.consumptions (invoice_id);
psql:migrations/000003_data_model_products.up.sql:217: ERROR:  current transaction is aborted, commands ignored until end of transaction block
create index on bellevue.consumptions (date);
psql:migrations/000003_data_model_products.up.sql:218: ERROR:  current transaction is aborted, commands ignored until end of transaction block
----------------------------------------------------------------------------------
-- Update Permissions: -----------------------------------------------------------
alter table price_categories
owner to dev;
psql:migrations/000003_data_model_products.up.sql:225: ERROR:  current transaction is aborted, commands ignored until end of transaction block
alter table taxes
owner to dev;
psql:migrations/000003_data_model_products.up.sql:228: ERROR:  current transaction is aborted, commands ignored until end of transaction block
alter table financial_accounts
owner to dev;
psql:migrations/000003_data_model_products.up.sql:231: ERROR:  current transaction is aborted, commands ignored until end of transaction block
alter table products
owner to dev;
psql:migrations/000003_data_model_products.up.sql:234: ERROR:  current transaction is aborted, commands ignored until end of transaction block
alter table product_form_order
owner to dev;
psql:migrations/000003_data_model_products.up.sql:237: ERROR:  current transaction is aborted, commands ignored until end of transaction block
alter table comments
owner to dev;
psql:migrations/000003_data_model_products.up.sql:240: ERROR:  current transaction is aborted, commands ignored until end of transaction block
alter table consumptions
owner to dev;
psql:migrations/000003_data_model_products.up.sql:243: ERROR:  current transaction is aborted, commands ignored until end of transaction block
GRANT SELECT, INSERT, UPDATE, DELETE
ON ALL TABLES IN SCHEMA bellevue
TO app;
psql:migrations/000003_data_model_products.up.sql:247: ERROR:  current transaction is aborted, commands ignored until end of transaction block
GRANT USAGE, SELECT, UPDATE
ON ALL SEQUENCES IN SCHEMA bellevue
TO app;
psql:migrations/000003_data_model_products.up.sql:251: ERROR:  current transaction is aborted, commands ignored until end of transaction block
COMMIT;
ROLLBACK
