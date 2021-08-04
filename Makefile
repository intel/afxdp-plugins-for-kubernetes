CLOC := $(shell command -v cloc 2> /dev/null)

.PHONY: all

start:
	@echo "********************************************"
	@echo "*   CNDP Makefile - Running  Makefile      *"
	@echo "********************************************"

format:
	@echo "***   Go Format   ***"
	@echo "Runing Go Format on the following files:"
	go fmt github.com/intel/cndp_device_plugin/...
	@echo

	@echo "*** Clang Format  ***"
	@echo "Runing Clang Format on the following C files:"	
	clang-format -i -style=file pkg/bpf/*.c pkg/bpf/*.h
	@echo

lint:
	@echo "***   Go Lint     ***"
	go get -u golang.org/x/lint/golint &> /dev/null
	golint=$(shell go list -f {{.Target}} golang.org/x/lint/golint)
	golint ./...
	@echo

build:
	@echo "***   Build DP    ***"
	gcc ./pkg/bpf/bpfWrapper.c -lbpf -c -o ./pkg/bpf/bpfWrapper.o
	ar rs ./pkg/bpf/libwrapper.a ./pkg/bpf/bpfWrapper.o  &> /dev/null
	go build -o ./bin/cndp-dp ./cmd/cndp-dp
	@echo

	@echo "***   Build CNI   ***"
	go build -o ./bin/cndp ./cmd/cndp-cni
	@echo

test: 
	@echo "***  Unit Tests   ***"
	go test $(shell go list ./... | grep -v "/examples/e2e-test/" | grep -v "/pkg/resourcesapi")
	@echo

#installation:
#       install to deploy once we have a damenset that can deploy DP and CNI

cloc: 
	./update-cloc.sh
	@echo

clean:
	@echo "***    Cleanup    ***"
	rm -f ./bin/cndp
	rm -f ./bin/cndp-dp
	@echo

end:
	@echo "Build complete!"	

all: start format lint build test cloc clean end
