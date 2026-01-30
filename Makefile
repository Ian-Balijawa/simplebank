DB_URL=postgresql://postgres:secret@localhost:5432/simple_bank?sslmode=disable
NETWORK=bank-network
POSTGRES_CONTAINER=postgres

network:
	docker network inspect $(NETWORK) >/dev/null 2>&1 || docker network create $(NETWORK)

postgres:
	docker rm -f $(POSTGRES_CONTAINER) 2>/dev/null || true
	docker run --name $(POSTGRES_CONTAINER) \
		--network $(NETWORK) \
		-p 5432:5432 \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=secret \
		-d postgres:14-alpine

postgres-remove:
	docker rm -f $(POSTGRES_CONTAINER) 2>/dev/null || true

mysql:
	docker rm -f mysql8 2>/dev/null || true
	docker run --name mysql8 \
		--network $(NETWORK) \
		-p 3306:3306 \
		-e MYSQL_ROOT_PASSWORD=secret \
		-d mysql:8

redis:
	docker rm -f redis 2>/dev/null || true
	docker run --name redis \
		--network $(NETWORK) \
		-p 6379:6379 \
		-d redis:7-alpine

createdb:
	docker exec -it $(POSTGRES_CONTAINER) createdb \
		--username=postgres \
		--owner=postgres \
		simple_bank || true

dropdb:
	docker exec -it $(POSTGRES_CONTAINER) dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

# =========================
# Documentation & Schema
# =========================

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

# =========================
# Code Generation
# =========================

sqlc:
	sqlc generate

mock:
	mockgen -package mockdb -destination db/mock/store.go \
		github.com/techschool/simplebank/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go \
		github.com/techschool/simplebank/worker TaskDistributor

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=doc/swagger \
		--openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
		proto/*.proto
	statik -src=./doc/swagger -dest=./doc

# =========================
# Dev & Test
# =========================

test:
	go test -v -cover -short ./...

server:
	go run main.go

evans:
	evans --host localhost --port 9090 -r repl

# =========================
# Phony
# =========================

.PHONY: \
	network postgres postgres-remove mysql redis \
	createdb dropdb \
	migrateup migratedown migrateup1 migratedown1 \
	new_migration \
	db_docs db_schema \
	sqlc mock proto \
	test server evans
