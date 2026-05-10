GOCACHE ?= $(CURDIR)/.gocache

.PHONY: api clean db-migrate help test tidy

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

clean:
	rm -rf .gocache bin
