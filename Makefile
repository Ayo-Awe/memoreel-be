include .env

migrate-up:
	migrate -path database/migrations -database "${DB_DSN}" -verbose up

migrate-down:
	migrate -path database/migrations -database "${DB_DSN}" -verbose down

migrate-force:
	migrate -path database/migrations -database "${DB_DSN}" -verbose force $(version)

new-migration:
	migrate create -ext sql -dir database/migrations -seq $(name)

.PHONY: migrate-up migrate-down new-migration migrate-force
