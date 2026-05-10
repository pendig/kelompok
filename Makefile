GOCACHE ?= $(CURDIR)/.gocache

.PHONY: api clean db-migrate help seed-demo test tidy

help:
	@GOCACHE=$(GOCACHE) go run ./cmd/kelompok help

tidy:
	GOCACHE=$(GOCACHE) go mod tidy

test:
	GOCACHE=$(GOCACHE) go test ./...

api:
	GOCACHE=$(GOCACHE) go run ./cmd/kelompok-api

db-migrate:
	GOCACHE=$(GOCACHE) go run ./cmd/kelompok db migrate

seed-demo:
	GOCACHE=$(GOCACHE) go run ./cmd/kelompok seed demo

clean:
	rm -rf .gocache bin
