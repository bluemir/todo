IMPORT_PATH=github.com/bluemir/todo
BIN_NAME=$(notdir $(IMPORT_PATH))

default: $(BIN_NAME)

GIT_COMMIT_ID = $(shell git rev-parse --short HEAD)
VERSION=$(GIT_COMMIT_ID)-$(shell date +"%Y%m%d.%H%M%S")

GO_SOURCES = $(shell find . -type f -name '*.go' -print)

# Automatic runner
DIRS = $(shell find . -name dist -prune -o -name ".git" -prune -o -type d -print)

.sources:
	@echo $(DIRS) makefile \
		$(GO_SOURCES) | tr " " "\n"
run: $(BIN_NAME)
	./$(BIN_NAME) -vvvv -i example.yaml run --dry-run -- cat 'hey something'

auto-run:
	while true; do \
		make .sources | entr -rd make run ;  \
		echo "hit ^C again to quit" && sleep 1  \
	; done
reset:
	ps -e | grep make | grep -v grep | awk '{print $$1}' | xargs kill

## Binary build
$(BIN_NAME): $(GO_SOURCES)
	go build -v \
		-ldflags "-X main.VERSION=$(VERSION)" \
		-o $(BIN_NAME) .
	@echo Build DONE

## Multi platform
deploy: build/linux/amd64/$(BIN_NAME)
deploy: build/linux/arm/$(BIN_NAME)
deploy: build/windows/amd64/$(BIN_NAME)
#deploy: build/windows/arm/$(BIN_NAME)
# make hook.mk file for your hook (example. following lines)
#deploy:
	# TODO scp or upload binary
	# TODO call hook to deploy(ex. docker command)

build/%/$(BIN_NAME): export GOOS=$(subst /,,$(dir $*))
build/%/$(BIN_NAME): export GOARCH=$(notdir $*)
build/%/$(BIN_NAME):
	@echo --------------------------BUILD $$GOOS $$GOARCH-----------------------------
	make clean
	make $(BIN_NAME)
	mkdir -p $(@D)
	mv $(BIN_NAME) $@

clean:
	rm -rf dist/ vendor/ $(BIN_NAME)
	go clean

.PHONY: .sources run auto-run reset tools clean
