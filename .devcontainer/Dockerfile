# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26
FROM mcr.microsoft.com/devcontainers/go:${GO_VERSION}-bookworm

ENV TZ=Asia/Tokyo

# エディタの保存時フォーマット（go.formatTool=goimports / import 整理）が利用する
# goimports を事前導入し、初回起動時のツール導入プロンプトを回避する。
RUN go install golang.org/x/tools/cmd/goimports@latest
