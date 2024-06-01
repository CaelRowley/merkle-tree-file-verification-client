FROM golang:1.22.1

WORKDIR /app

COPY . .

RUN go mod download && go mod verify

RUN  go build -o main cmd/main.go

CMD ["./main"]
