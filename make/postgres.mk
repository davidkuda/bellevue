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

psql/davidkuda:
	psql \
	--host localhost \
	--username davidkuda \
	--port 5432 \
	--dbname kuda_ai

user ?= dev
psql/dev:
	psql \
	--host localhost \
	--username ${user} \
	--port 5432 \
	--dbname kuda_ai
