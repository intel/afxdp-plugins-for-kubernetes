.PHONY: all

all: start compilation build test end

start:
	echo "***** CNDP Makefile - running Makefile *****"

compilation:
	echo "***** Compiling CNDP binaries *****"
	echo "***** Go Format *****"
	gofmt -w -s .
	echo "***** Go Lint *****"
	#install golint if needed, suppress output to keep output clean
	go get -u golang.org/x/lint/golint &> /dev/null
	#where was golint installed?
	golint=$(go list -f {{.Target}} golang.org/x/lint/golint)
	#run golint
	echo "The following files have linting issues:"
	golint ./...
	#need to add a watcher anytime a string changes 

format:
	echo "***** Format CNDP files *****"
	echo "Runing C Formatter..."
	clang-format -i ./pkg/bpf/wrapper.c

build:
	echo "***** Build  CNDP files *****"
	echo "Building Device Plugin"
	gcc ./pkg/bpf/wrapper.c -lbpf -c -o ./pkg/bpf/wrapper.o
	ar rs ./pkg/bpf/libwrapper.a ./pkg/bpf/wrapper.o  &> /dev/null
	go build -o ./bin/cndp-dp ./cmd/cndp-dp
	echo "Building CNI"
	go build -o ./bin/cndp ./cmd/cndp-cni
	
test: 
	echo "***** Testing CNDP files *****"	
	go test ./... -v *_test.go
	go test github.com/intel/cndp_device_plugin/...
run:
	cd ./bin && ./cndp
	cd ./bin && ./cndp-dp

clean:
	echo "***** cleaning CNDP files *****"
	rm -f ./bin/cndp
	rm -f ./bin/cndp-dp


end:
	echo "Build complete!"	

