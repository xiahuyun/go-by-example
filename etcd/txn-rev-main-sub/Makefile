SHELL := /bin/bash

.PHONY: help tidy run demo clean

help:
	@echo "Available targets:"
	@echo "  make tidy  - run go mod tidy"
	@echo "  make run   - run demo against ETCD_ENDPOINTS and ETCD_DATA_DIR"
	@echo "  make demo  - start/reuse local etcd and run demo"
	@echo "  make clean - remove local artifacts"

tidy:
	go mod tidy

run:
	go run .

demo:
	./run-with-local-etcd.sh

clean:
	rm -rf .tmp
