SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

.PHONY: help test vet race check build verify

help:
	@printf '%s\n' \
		'make test    - run the full Go test suite' \
		'make vet     - run Go vet' \
		'make race    - race-test the challenge package' \
		'make check   - run formatting, vet, tests, race tests, and shell checks' \
		'make build   - build the production-style binary in .bin/go-away' \
		'make verify  - run check and build'

test:
	./ops/go.sh test ./...

vet:
	./ops/go.sh vet ./...

race:
	./ops/go.sh test -race ./lib/challenge

check:
	./ops/check.sh

build:
	./ops/build.sh

verify: check build
