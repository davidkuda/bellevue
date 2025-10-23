bundle/js:
	esbuild \
	--bundle \
	--minify \
	./ui/static/js/bundle.js \
	--outfile=./ui/static/dist/app.js

bundle/js/watch:
	esbuild \
	--bundle \
	--minify \
	./ui/static/js/bundle.js \
	--outfile=./ui/static/dist/app.js \
	--watch


bundle/css:
	esbuild \
	--bundle \
	--minify \
	./ui/static/css/bundle.css \
	--outfile=./ui/static/dist/styles.css

bundle/css/watch:
	esbuild \
	--bundle \
	--minify \
	./ui/static/css/bundle.css \
	--outfile=./ui/static/dist/styles.css \
	--watch

fmt/ui: fmt/ui/static fmt/ui/templates

fmt/ui/static:
	biome format ./ui/static/css/ --write
	biome format ./ui/static/js/ --write

fmt/ui/templates:
	./node_modules/prettier/bin/prettier.cjs \
	./ui/html/ \
	--write

PG_DSN_ADMIN = postgres://davidkuda:@${DB_ADDRESS}/${DB_NAME}?sslmode=disable
PG_DSN_APP = postgres://${DB_USER}:${DB_PASSWORD}@${DB_ADDRESS}/${DB_NAME}?sslmode=disable

db/backup/songs:
	pg_dump \
	--data-only \
	--column-inserts \
	--no-privileges \
	--no-owner \
	--table=songs \
	> ./data/postgres/2025-07-20--backup--songs


db/backup/full:
	pg_dump \
	--data-only \
	--column-inserts \
	--no-privileges \
	--no-owner \
	> ./data/postgres/2025-07-20--backup--full

db/init:
	createdb kuda_ai

db/drop:
	dropdb kuda_ai

db/restore:
	psql -X kuda_ai --single-transaction < ./data/postgres/dumpfile--data-only

db/migrate/newsql:
	@migrate create \
	-seq \
	-ext=.sql \
	-dir=./migrations \
	${name}

db/migrate/up-all:
	@migrate \
	-path=./migrations \
	-database=${PG_DSN_ADMIN} \
	up

db/migrate/version:
	migrate \
	-path=./migrations/ \
	-database=${PG_DSN_ADMIN} \
	version

db/migrate/force:
	@migrate \
	-path=./migrations/ \
	-database=${PG_DSN_ADMIN} \
	force ${version}

# migrate down one step
db/migrate/down-1:
	@migrate \
	-path=./migrations/ \
	-database=${PG_DSN_ADMIN} \
	down 1

# migrate up one step
db/migrate/up-1:
	@migrate \
	-verbose \
	-path=./migrations/ \
	-database=${PG_DSN_ADMIN} \
	up 1

user ?= dev
psql/dev:
	psql \
	--host localhost \
	--username ${user} \
	--port 5432 \
	--dbname kuda_ai
