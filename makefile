IMPORT_PATH=github.com/bluemir/todo
BIN_NAME=$(notdir $(IMPORT_PATH))

default: $(BIN_NAME)

GIT_COMMIT_ID = $(shell git rev-parse --short HEAD)
VERSION=$(GIT_COMMIT_ID)-$(shell date +"%Y%m%d.%H%M%S")

# if gopath not set, make inside current dir
ifeq ($(GOPATH),)
	GOPATH=$(PWD)/.GOPATH
endif

GO_SOURCES = $(shell find . -name ".GOPATH" -prune -o -type f -name '*.go' -print)

# Automatic runner
DIRS = $(shell find . -name dist -prune -o -name ".git" -prune -o -type d -print)

.sources:
	@echo $(DIRS) makefile \
		$(GO_SOURCES) | tr " " "\n"
run: $(BIN_NAME)
	./$(BIN_NAME)
auto-run:
	while true; do \
		make .sources | entr -rd make run ;  \
		echo "hit ^C again to quit" && sleep 1  \
	; done
reset:
	ps -e | grep make | grep -v grep | awk '{print $$1}' | xargs kill

## Binary build
$(BIN_NAME): $(GO_SOURCES) $(GOPATH)/src/$(IMPORT_PATH)
	go get -v -d $(IMPORT_PATH)            # can replace with glide
	go build -v \
		-ldflags "-X main.VERSION=$(VERSION)" \
		-o $(BIN_NAME) .
	@echo Build DONE

clean:
	rm -rf dist/ vendor/ $(BIN_NAME)
	go clean

$(GOPATH)/src/$(IMPORT_PATH):
	@echo "make symbolic link on $(GOPATH)/src/$(IMPORT_PATH)..."
	@mkdir -p $(dir $(GOPATH)/src/$(IMPORT_PATH))
	ln -s $(PWD) $(GOPATH)/src/$(IMPORT_PATH)

.PHONY: .sources run auto-run reset tools clean
