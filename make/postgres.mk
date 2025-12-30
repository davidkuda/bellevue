db/date:
	date +%Y-%m-%d
	date +%F
	echo today is $$(date +%F) ok

db/backup/bellevue/sudo:
	sudo -u postgres \
	pg_dump bellevue \
	--data-only  \
	--schema bellevue \
	--column-inserts > pg-backup.bellevue.$$(date +%F).sql

db/backup/bellevue:
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

# -X => --no-psqlrc
db/restore:
	psql -X bellevue \
	--single-transaction \
	--file ${file}

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
