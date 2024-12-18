ARG SERVICE_NAME

FROM golang:1.23-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.5

CMD ["air", "-c", ".air.toml"]
