FROM golang:latest

WORKDIR /app

COPY ./.env ./.env 

COPY ./bank_gateway_microservice/cmd ./cmd
COPY ./bank_gateway_microservice/internal ./internal
COPY ./bank_gateway_microservice/.air.toml ./.air.toml
COPY ./bank_gateway_microservice/go.mod ./go.mod
COPY ./bank_gateway_microservice/go.sum ./go.sum


RUN go build -o main cmd/api/main.go

EXPOSE 8080

CMD ["./main"]
