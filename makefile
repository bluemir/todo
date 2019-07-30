IMPORT_PATH=$(shell cat go.mod | head -n 1 | awk '{print $$2}')
BIN_NAME=$(notdir $(IMPORT_PATH))

default: build/$(BIN_NAME)

GIT_COMMIT_ID = $(shell git rev-parse --short HEAD)
VERSION=$(GIT_COMMIT_ID)-$(shell date +"%Y%m%d.%H%M%S")

GO_SOURCES = $(shell find . -type f -name '*.go' -print)

run: build/$(BIN_NAME)
	build/$(BIN_NAME) -vvvv -i example.yaml run --dry-run -- cat 'hey something'

## Binary build
build/$(BIN_NAME): build/release///$(BIN_NAME)
	cp $< $@

## Multi platform
release: build/release/linux/amd64/$(BIN_NAME)
release: build/release/linux/arm/$(BIN_NAME)
release: build/release/windows/amd64/$(BIN_NAME)

build/release/%/$(BIN_NAME): export GOOS=$(subst /,,$(dir $*))
build/release/%/$(BIN_NAME): export GOARCH=$(notdir $*)
build/release/%/$(BIN_NAME): $(GO_SOURCES)
	@echo --------------------------BUILD $$GOOS $$GOARCH-----------------------------
	go build -v \
		-ldflags "-X main.VERSION=$(VERSION)" \
		-o $@ .
	@echo Build DONE

clean:
	rm -rf build/
	go clean

.PHONY: default run release clean
