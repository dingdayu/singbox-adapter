# syntax=docker/dockerfile:1
ARG GOVERSION=1.25
ARG APP_NAME=app
ARG HTTP_PORT=8080


FROM golang:${GOVERSION}-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /out/${APP_NAME} .


FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /out/${APP_NAME} /app/${APP_NAME}
ENV GIN_MODE=release PORT=${HTTP_PORT} OTEL_SERVICE_NAME=${APP_NAME}
EXPOSE ${HTTP_PORT}
USER nonroot:nonroot

# Define a health check to monitor the service status
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl --fail http://localhost:${HTTP_PORT}/health || exit 1

# Set the entrypoint command to start the processgo binary with 'http' argument
ENTRYPOINT ["/app/${APP_NAME}", "http"]
