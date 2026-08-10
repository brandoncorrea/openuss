.PHONY: build
build:
	go build -o bin/openuss ./cmd/openuss

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: image
image:
	docker image build -t openuss:local .

.PHONY: run
run: image
	docker run --rm -d \
		--name openuss \
		--hostname openuss.uss5.localutm \
		--network interop_ecosystem_network \
		openuss:local

.PHONY: stop
stop:
	-docker rm -f openuss
