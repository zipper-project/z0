TEST = $(shell go list ./...)
all:
	@go install ./cmd/z0
	go build -o z0 ./cmd/z0/main.go

run:
	@./z0
stop:
clear:
test:
	@echo $(TEST)
	go test $(TEST)

.PHONY: all run stop clear test
