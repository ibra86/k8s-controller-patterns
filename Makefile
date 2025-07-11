APP = k8s-controller-patterns
GOOS ?= linux
GOARCH ?= amd64
VERSION ?= $(sell git describe --tags --always --dirty)
BUILD_FLAGS = -v -o $(APP) -ldflags "-X=github.com/ibra86/$(APP)/cmd.appVersion=$(VERSION)"
ARGS ?=
CRD_PATH = $(shell pwd)/config/crd


.PHONY: all build test run docker-build clean

all: build

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) main.go

test-ci:
	echo "Using CRD_PATH=${CRD_PATH_GW}"
	CRD_PATH=${CRD_PATH_GW} go test -v -p 1 ./...

test:
	CRD_PATH=${CRD_PATH} go test ./...
vet:
	go vet ./...
lint:
	golangci-lint run
check: test vet lint

run:
	go run main.go $(ARGS)

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(APP):latest .

clean:
	rm -f $(APP)