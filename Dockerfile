# One image for everything: the API, the app and the landing page.

# The app's TypeScript is compiled here; Node is not needed at runtime.
FROM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN mkdir -p ../server/site/static/js && npm run build

FROM golang:1.26-alpine AS server
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
COPY --from=frontend /src/server/site/static/js ./site/static/js
# The release, from git describe on the host (make image, CI); .git isn't copied in.
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /300sf

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=server /300sf ./
COPY server/email-templates ./email-templates
# The licenses of 300sf and of what it ships with go with every copy.
COPY LICENSE THIRD_PARTY_NOTICES.md ./
COPY LICENSES ./LICENSES
USER app
# The config holds secrets, so it is mounted at runtime rather than built in.
CMD ["./300sf", "-config", "/config/300sf.yml"]
