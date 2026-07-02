.DEFAULT_GOAL := help
.PHONY: help run build test test-race fmt vet tidy

help:        ## このヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

run:         ## サーバを起動 (http://localhost:8080)
	go run ./cmd/todoapp

build:       ## 全パッケージをビルド
	go build ./...

test:        ## 全テストを実行
	go test ./...

test-race:   ## データ競合検出つきでテスト
	go test ./... -race

fmt:         ## gofmt で整形（goimports があれば import も整理）
	gofmt -w .

vet:         ## go vet で静的解析
	go vet ./...

tidy:        ## go.mod / go.sum を最小化
	go mod tidy
