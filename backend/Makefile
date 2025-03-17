migrations_dir = db/migrations

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING=postgresql://user:pass@localhost:5432/fiendlist

goose/create:
	goose -dir $(migrations_dir) create rename_this_file sql

goose/status:
	goose -dir $(migrations_dir) status

goose/up:
	goose -dir $(migrations_dir) up

goose/down:
	goose -dir $(migrations_dir) down