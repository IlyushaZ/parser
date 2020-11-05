.PHONY: run
run:
	go build -o parser main.go && ./parser

.PHONY: build
build:
	go build -o parser main.go

.PHONY: escape-analysis
escape-analysis:
	go build -gcflags "-m -m"

.PHONY: compose
compose:
	docker-compose up -d

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: push-image
push-image:
	docker build -t ilyushagod/parser . && docker push ilyushagod/parser

.DEFAULT_GOAL := compose