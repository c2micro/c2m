# директория с бинарями
BIN_DIR=$(PWD)/bin
# entry-point c2m сервера
C2M_DIR=$(PWD)/cmd/c2msrv
# компиляторы
CC=gcc
CXX=g++
# для сборки
VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "0.0.0")
BUILD=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags="-s -w -X github.com/c2micro/c2msrv/internal/version.gitCommit=${BUILD} -X github.com/c2micro/c2msrv/internal/version.gitVersion=${VERSION}"
TAGS=sqlite_foreign_keys

.PHONY: c2msrv
c2msrv:
	@mkdir -p ${BIN_DIR}
	@echo "Building server..."
	CGO_ENABLED=0 CC=${CC} CXX=${CXX} go build ${LDFLAGS} -tags="${TAGS}" -o ${BIN_DIR}/c2msrv ${C2M_DIR}
	@strip bin/c2msrv

.PHONY: c2msrv-release
c2msrv-release:
	@mkdir -p ${BIN_DIR}
	@echo "Building server..."
	@CGO_ENABLED=0 CC=${CC} CXX=${CXX} go build ${LDFLAGS} -tags="${TAGS}" -o ${BIN_DIR}/c2msrv ${C2M_DIR}
	@echo "Strip..."
	@strip ${BIN_DIR}/c2msrv
	@echo "Compress..."
	@rm -f ${BIN_DIR}/c2msrv.release
	@upx -9 -o ${BIN_DIR}/c2msrv.release ${BIN_DIR}/c2msrv

.PHONY: c2msrv-race
c2msrv-race:
	@mkdir -p ${BIN_DIR}
	@echo "Building race server..."
	CC=${CC} CXX=${CXX} go build -race ${LDFLAGS} -o ${BIN_DIR}/c2msrv.race ${C2M_DIR}

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
