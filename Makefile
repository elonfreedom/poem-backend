.PHONY: help run build test lint lint-fix fmt vet clean

## help: 显示帮助信息
help:
	@echo "可用命令:"
	@echo "  make run        - 运行服务"
	@echo "  make build      - 编译二进制"
	@echo "  make test       - 运行测试"
	@echo "  make lint       - 运行 golangci-lint"
	@echo "  make lint-fix   - 运行 golangci-lint 并自动修复"
	@echo "  make fmt        - 格式化代码 (gofumpt + goimports)"
	@echo "  make vet        - 运行 go vet"
	@echo "  make clean      - 清理编译产物"

## run: 运行服务
run:
	go run cmd/server/main.go

## build: 编译二进制
build:
	go build -o bin/server cmd/server/main.go

## test: 运行测试
test:
	go test ./...

## lint: 运行 golangci-lint
lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint 未安装，正在安装..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...

## lint-fix: 运行 golangci-lint 并自动修复
lint-fix:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint 未安装，正在安装..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
	fi
	golangci-lint run --fix ./...

## fmt: 格式化代码 (gofumpt + goimports)
fmt:
	@if ! command -v gofumpt >/dev/null 2>&1; \
		then go install mvdan.cc/gofumpt@latest; \
	fi
	@if ! command -v goimports >/dev/null 2>&1; \
		then go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	gofumpt -w .
	goimports -w -local poem-backend .

## vet: 运行 go vet
vet:
	go vet ./...

## clean: 清理编译产物
clean:
	rm -rf bin/
	go clean
