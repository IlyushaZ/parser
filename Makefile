.PHONY: run
run:
	go build -o bin/parser cmd/parser/main.go && ./bin/parser

.PHONY: build
build:
	go build -o bin/parser cmd/parser/main.go

.PHONY: escape-analysis
escape-analysis:
	go build -gcflags "-m -m" cmd/parser/main.go

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