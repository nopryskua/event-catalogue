ARG SERVICE_NAME

FROM golang:1.23-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.1

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
