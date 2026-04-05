include ./make/bundle.mk
include ./make/migrate.mk
include ./make/postgres.mk

.PHONY: fmt/ui
fmt/ui:
	./node_modules/prettier/bin/prettier.cjs --write ./ui
