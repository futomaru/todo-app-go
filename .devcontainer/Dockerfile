# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26
FROM mcr.microsoft.com/devcontainers/go:${GO_VERSION}-bookworm

ENV TZ=Asia/Tokyo

# Makefile の `make fmt` が利用する goimports を追加で導入する。
RUN go install golang.org/x/tools/cmd/goimports@latest
