
FROM golang:latest

WORKDIR /app

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./.air.toml ./.air.toml
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum


RUN go build -o main cmd/api/main.go

EXPOSE 8080

CMD ["./main"]