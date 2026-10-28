FROM golang:1.24-alpine AS builder

LABEL maintainer="Joachim Klahr <joachim@klahr.se>"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o 300sf

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /root/

ARG CONFIG_FILE
COPY --from=builder /app/300sf ./
COPY $CONFIG_FILE ./300sf.conf
COPY ./email-templates ./email-templates/

CMD ["./300sf", "-config", "./300sf.conf"]
