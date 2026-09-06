.PHONY: build
build:
	go build -o bin/openuss ./cmd/openuss

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: dev
dev:
	set -a; [ -f .env ] && . .env; set +a; go run ./cmd/openuss

.PHONY: image
image:
	docker compose build

.PHONY: run
run:
	docker compose up -d --build --wait

.PHONY: stop
stop:
	-docker compose down

.PHONY: apis
apis:
	scripts/generate-apis.sh

export INTERUSS_MONITORING_DIR ?= ../monitoring

.PHONY: refresh-mocks
refresh-mocks:
	scripts/refresh-mocks.sh

.PHONY: automated-tests
automated-tests:
	set -a; [ -f .env ] && . .env; set +a; scripts/automated-tests.sh
