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

.DEFAULT_GOAL := compose