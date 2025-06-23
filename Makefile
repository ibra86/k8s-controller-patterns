APP = k8s-controller-patterns
GOOS ?= linux
GOARCH ?= amd64
VERSION ?= $(sell git describe --tags --always --dirty)
BUILD_FLAGS = -v -o $(APP) -ldflags "-X=github.com/ibra86/$(APP)/cmd.appVersion=$(VERSION)"
ARGS ?= 


.PHONY: all build test run docker-build clean

all: build

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) main.go

test:
	go test ./cmd

run:
	go run main.go $(ARGS)

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(APP):latest .

clean:
	rm -f $(APP)