# директория с бинарями
BIN_DIR=$(PWD)/bin
# entry-point c2m сервера
C2M_DIR=$(PWD)/cmd/c2m
# компиляторы
CC=gcc
CXX=g++
# для сборки
VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "0.0.0")
BUILD=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags="-s -w -X github.com/c2micro/c2m/internal/version.gitCommit=${BUILD} -X github.com/c2micro/c2m/internal/version.gitVersion=${VERSION}"
TAGS=sqlite_foreign_keys

.PHONY: run-local
run-local: c2m
	@bin/c2m --config config/config.yml -d run

.PHONY: c2m
c2m:
	@mkdir -p ${BIN_DIR}
	@echo "Building server..."
	CGO_ENABLED=0 CC=${CC} CXX=${CXX} go build ${LDFLAGS} -tags="${TAGS}" -o ${BIN_DIR}/c2m ${C2M_DIR}
	@strip bin/c2m

.PHONY: c2m-x64
c2m-x64:
	@mkdir -p ${BIN_DIR}
	@echo "Building server..."
	CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" CC=${CC} CXX=${CXX} go build ${LDFLAGS} -tags="${TAGS}" -o ${BIN_DIR}/c2m.x64 ${C2M_DIR}
	@strip bin/c2m.x64

.PHONY: c2m-release
c2m-release:
	@mkdir -p ${BIN_DIR}
	@echo "Building server..."
	@CGO_ENABLED=0 CC=${CC} CXX=${CXX} go build ${LDFLAGS} -tags="${TAGS}" -o ${BIN_DIR}/c2m ${C2M_DIR}
	@echo "Strip..."
	@strip ${BIN_DIR}/c2m
	@echo "Compress..."
	@rm -f ${BIN_DIR}/c2m.release
	@upx -9 -o ${BIN_DIR}/c2m.release ${BIN_DIR}/c2m

.PHONY: c2m-race
c2m-race:
	@mkdir -p ${BIN_DIR}
	@echo "Building race server..."
	CC=${CC} CXX=${CXX} go build -race ${LDFLAGS} -o ${BIN_DIR}/c2m.race ${C2M_DIR}

.PHONY: dep-shared
dep-shared:
	@echo "Update shared components..."
	@export GOPRIVATE="github.com/c2micro" && go get -u github.com/c2micro/c2mshr && go mod tidy && go mod vendor

.PHONY: ent-gen
ent-gen:
	@echo "Generating ent models..."
	@go generate ./internal/ent

.PHONY: atlas-sqlite
atlas-sqlite:
	@atlas schema inspect -u "ent://internal/ent/schema" --format '{{ sql . "  " }}' --dev-url "sqlite://file?mode=memory&_fk=1"

.PHONY: atlas-erd
atlas-erd:
	@atlas schema inspect -u "ent://internal/ent/schema" --dev-url "sqlite://file?mode=memory&_fk=1" -w

.PHONY: clean
clean:
	@rm -rf ${BINARY_DIR}
