FROM golang:1.25.5-alpine3.23 AS base

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

FROM base AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build

FROM alpine:3.23 AS runner

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

RUN mkdir -p /app/images && chown -R appuser:appgroup /app

COPY --from=build --chown=appuser:appgroup /app/image-api /app/

USER appuser

ENTRYPOINT [ "./image-api" ]
