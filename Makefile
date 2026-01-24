.PHONY: test setup

test:
	go test ./...

setup:
	git config core.hooksPath .githooks
