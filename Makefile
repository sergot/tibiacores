migrations_dir = backend/db/migrations

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING=postgresql://user:pass@localhost:5432/tibiacores

## Sync soul core creature data from TibiaWiki:
##   1. Fetch canonical creature list → data/creatures.txt
##   2. Download any missing soul core GIFs → frontend/public/assets/soulcores/
##   3. Generate a goose migration for new creatures → backend/db/migrations/
## After reviewing the migration: make goose/up && cp data/creatures.txt data/creatures-synced.txt
sync-creatures:
	cd frontend && npx tsx scripts/fetch-soulcore-creatures.ts
	cd frontend && npx tsx scripts/download-soulcore-images.ts
	cd frontend && npx tsx scripts/sync-db-creatures.ts

goose/create:
	goose -dir $(migrations_dir) create rename_this_file sql

goose/status:
	goose -dir $(migrations_dir) status

goose/up:
	goose -dir $(migrations_dir) up

goose/down:
	goose -dir $(migrations_dir) down

goose/reset:
	goose -dir $(migrations_dir) reset