APP_NAME=app
CMD_DIR=cmd/$(APP_NAME)
BIN_DIR=bin

.PHONY: build run clean test fmt vet

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)

run:
	go run ./$(CMD_DIR)

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
