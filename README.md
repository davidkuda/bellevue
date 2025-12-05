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
2. Run the docker compose with `docker compose up -d`. 
Note that we have mapped the data directory into the `db` folder. The data will be persisted and you only need to do this setup once.
To delete the database data, just delete the folder `db/data_postgres` that was created when you ran docker compose. 

4. Jump into a bash terminal on the container with:

```sh
docker exec -it \
-e PGPASSWORD="pa55word" \
-e PGDATABASE="bellevue" \
-e PGUSER="bellevue" \
postgres bash
```

5. Run the initialisation scripts that you copied earlier
```sh
psql -f migrations/000001_create_roles.up.sql
psql -f migrations/000002_data_model_products.up.sql
psql -f migrations/000003_populate_product_tables.up.sql
```


If you have `psql` as a command on your system and docker is running in a container, you can export env vars and access psql in the container directly.

```sh
export PGDATABASE=bellevue
export PGUSER=bellevueadmin
export PGPASSWORD=pa55sword
export PGPORT=5432
export PGHOST=localhost

psql -X -q -c '\conninfo'
```

## npm
Install npm dependencies with `npm install` in the root of the directory and then `npm install esbuild`

## make
In order to make the JS and CSS bundels, you need to update the path to esbuild in the `make/bundle.mk` file. 

If you install esbuild locally to the project, you should find it in a folder called `node_modules`. For me, it was `node_modules/esbuild/bin/esbuild`

Then run 
```sh
make bundle/css
make bundle/js
```

## Envriomental variables
The website is configured via envriomental variables. Create a file called `envs` and past the following variables into it:
```sh
export DB_SCHEME="postgres"
export DB_USER="bellevue"
export DB_PASSWORD="pa55word"
export DB_ADDRESS="localhost"
export DB_NAME="bellevue"

export PGHOST=localhost
export PGPORT="5432"
export PGUSER="bellevue"
export PGDATABASE="bellevue"

export PG_DSN="contact your administrator"
export JWT_SECRET_KEY="contact your administrator"

export COOKIE_DOMAIN=localhost

export OIDC_CLIENT_ID="contact your administrator"
export OIDC_CLIENT_SECRET="contact your administrator"
export OIDC_ISSUER="https://sso.dwbn.org"
export OIDC_REDIRECT_URL="http://localhost:8875/login/dwbn/callback"
export SESSION_SECRET="randomly-genearted-secret--see-Makefile"
```
## Run the website!
Install go packages with `go mod tidy` and then run the website with `go run ./cmd/web/`. If all went correctly, you should be able to see it at [http://localhost:8875](http://localhost:8875).
