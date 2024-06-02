.PHONY: all clean build test run
BINARY_NAME=main

all: clean build test

clean:
	rm -f ${BINARY_NAME}
	rm -rf files/dummy

build:
	go build -o ${BINARY_NAME} cmd/main.go

test:
	go test ./...

run: build
	./${BINARY_NAME}
