# Go 参数
GOVERSION=1.25
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GO_HTTP_PORT=8080

# 二进制文件名
APP_NAME = app

VERSION := $(shell git describe --tags --always --long --dirty 2>/dev/null || echo "v0.0.0")
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell TZ=Asia/Shanghai date +"%Y-%m-%d %H:%M:%S")

# 检查是否在 CI 环境下，如果在 CI 中使用 GitLab 的预设变量
ifdef CI_COMMIT_TAG
	VERSION := $(CI_COMMIT_TAG)
else ifdef CI_COMMIT_SHORT_SHA
	VERSION := $(CI_COMMIT_SHORT_SHA)
endif

# 主要目标
all: test build

build: clean
	@echo "Building with version: $(VERSION)"
	$(GOBUILD) -o $(APP_NAME) -v -ldflags "-X 'github.com/dingdayu/go-project-template/model/entity.BuildVersion=$(VERSION)' -X 'github.com/dingdayu/go-project-template/model/entity.BuildTime=$(BUILD_TIME)'"

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

clean:
	$(GOCLEAN)
	rm -f $(APP_NAME)

run:
	OTEL_SERVICE_NAME=$(APP_NAME) OTEL_EXPORTER_OTLP_ENDPOINT="http://otelcol-opentelemetry-collector.observability.svc:4318" go run . http

deps:
	$(GOGET) -v -t -d ./...

# 交叉编译
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(APP_NAME)_linux_amd64 -v -ldflags "-X 'github.com/dingdayu/go-project-template/model/entity.BuildVersion=$(VERSION)' -X 'github.com/dingdayu/go-project-template/model/entity.BuildTime=$(BUILD_TIME)'"

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(APP_NAME).exe -v -ldflags "-X 'github.com/dingdayu/go-project-template/model/entity.BuildVersion=$(VERSION)' -X 'github.com/dingdayu/go-project-template/model/entity.BuildTime=$(BUILD_TIME)'"


help:
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

download-geoip: ## download geoip database
	curl -O -J -L -L -u $(MM_ACCOUNT_ID):$(MM_LICENSE_KEY) https://download.maxmind.com/geoip/databases/GeoLite2-City/download?suffix=tar.gz && \
	mkdir -p GeoLite2-City && \
	tar -zxvf GeoLite2-City_*.tar.gz --strip-components=1 -C GeoLite2-City

tidy: ## go mod tidy
	$(GO) mod tidy

fmt: ## go fmt
	$(GO) fmt $(PKG)

lint: ## golangci-lint run
	golangci-lint run

docker-build: ## docker build image
	docker build --build-arg GOVERSION=$(GOVERSION) --build-arg APP_NAME=$(APP_NAME) --build-arg GO_HTTP_PORT=$(GO_HTTP_PORT) --build-arg MM_ACCOUNT_ID=$(MM_ACCOUNT_ID) --build-arg MM_LICENSE_KEY=$(MM_LICENSE_KEY) -t $(APP_NAME):dev .


# 版本化迁移方案，使用 Atlas + GORM
# 参考：https://atlasgo.io/getting-started/gorm.html
# migrate-db:
# 	atlas migrate diff $(name) --env gorm

# inspect-db:
# 	atlas schema inspect \
# 	  --url "$(POSTGRES_DSN)" \
# 	  -w

# inspect-gorm:
# 	atlas schema inspect --env gorm --url "env://src"

# migrate-up:
# 	atlas migrate apply --env gorm -u "$(PGURL)"

# migrate-plan:
# 	atlas migrate apply --env gorm -u "$(PGURL)" --dry-run

.PHONY: all build test clean run deps build-linux build-windows deploy receiver