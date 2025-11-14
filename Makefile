include ./make/bundle.mk
include ./make/migrations.mk
include ./make/postgres.mk

.PHONY: fmt/ui
fmt/ui:
	./node_modules/prettier/bin/prettier.cjs --write ./ui
