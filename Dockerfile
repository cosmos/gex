# syntax = docker/dockerfile:1.2
# WARNING! Use `DOCKER_BUILDKIT=1` with `docker build` to enable --mount feature.

## prep the base image.
#
FROM golang:1.21.5 as base

RUN apt update && \
    apt-get install -y \
        build-essential \
        ca-certificates \
        curl

# enable faster module downloading.
ENV GOPROXY https://proxy.golang.org

## builder stage.
#
FROM base as builder

WORKDIR /

# cache dependencies.
COPY ./go.mod . 
COPY ./go.sum . 
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build go install -v ./...

## prep the final image.
#
FROM base

RUN useradd -ms /bin/bash tendermint
USER tendermint

COPY --from=builder /go/bin/gex /usr/bin

WORKDIR /apps

ENTRYPOINT ["gex -h"]
