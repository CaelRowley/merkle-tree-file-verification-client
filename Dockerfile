FROM golang:1.22.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN  go build -o main app/main.go   

CMD ["./main"]