# syntax=docker/dockerfile:1

FROM golang:1.26.5-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY src ./src

ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
	go build -trimpath -ldflags="-s -w" -o /game-server ./cmd/server

FROM alpine:3.24

RUN apk add --no-cache libffi libx11

COPY --from=build /game-server /game-server

ENV HOME=/tmp \
	XDG_CACHE_HOME=/tmp \
	SERVER_HOST=0.0.0.0 \
	SERVER_PORT=58008

EXPOSE 58008/tcp
USER 65532:65532

ENTRYPOINT ["/game-server"]
