SHELL := /bin/bash

.PHONY: help tidy build run mvcc-demo clean

help:
	@echo "Available targets:"
	@echo "  make tidy       - run go mod tidy"
	@echo "  make build      - build mvcc demo binary"
	@echo "  make run        - run demo against ETCD_ENDPOINTS (default from main.go)"
	@echo "  make mvcc-demo  - start/reuse local etcd and run demo"
	@echo "  make clean      - remove local build/cache artifacts for this demo"

tidy:
	go mod tidy

build:
	go build -o mvcc .

run:
	go run .

mvcc-demo:
	./run-with-local-etcd.sh

clean:
	rm -f mvcc
	rm -rf .tmp
