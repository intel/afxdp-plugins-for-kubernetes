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

build:
	echo "***** Build  CNDP files *****"
	echo "Building Device Plugin"
	go build -o ./bin/cndp-dp ./cmd/cndp-dp
	echo "Building CNI"
	go build -o ./bin/cndp ./cmd/cndp-cni
	
test: 
	echo "***** Testing CNDP files *****"	
	go test ./... -v *_test.go
run:
	cd ./bin && ./cndp
	cd ./bin && ./cndp-dp

clean:
	echo "***** cleaning CNDP files *****"
	rm -f ./bin/cndp
	rm -f ./bin/cndp-dp


end:
	echo "Build complete!"	

