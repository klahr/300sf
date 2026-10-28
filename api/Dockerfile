FROM golang:latest as builder

LABEL maintainer="Joachim Klahr <joachim@klahr.se>"

RUN mkdir /app
ADD . /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o 300sf

ARG CONFIG_FILE
COPY $CONFIG_FILE ./300sf.conf
COPY ./private_key.pem ./public_key.pem ./
COPY ./email-templates ./email-templates/

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/300sf ./
COPY --from=builder /app/300sf.conf ./
COPY --from=builder /app/private_key.pem ./
COPY --from=builder /app/public_key.pem ./
COPY --from=builder /app/email-templates ./email-templates/

CMD ["./300sf", "-config", "./300sf.conf"]
