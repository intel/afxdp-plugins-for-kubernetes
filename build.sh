#!/bin/bash

echo "***** Go Format *****"
echo "Runing Go Format on the following files:"
go fmt github.com/intel/cndp_device_plugin/...
echo

echo "***** Clang Format *****"
echo "Runing C Formatter"
clang-format -i -style=file pkg/bpf/*.c pkg/bpf/*.h
echo

echo "***** Go Lint *****"
#install golint if needed, suppress output to keep output clean
go get -u golang.org/x/lint/golint &> /dev/null
#where was golint installed?
golint=$(go list -f {{.Target}} golang.org/x/lint/golint)
#run golint
echo "The following files have linting issues:"
eval "$golint ./..."
echo

echo "***** Build *****"
echo "Building Device Plugin"
gcc ./pkg/bpf/bpfWrapper.c -lbpf -c -o ./pkg/bpf/bpfWrapper.o
ar rs ./pkg/bpf/libwrapper.a ./pkg/bpf/bpfWrapper.o  &> /dev/null
go build -o ./bin/cndp-dp ./cmd/cndp-dp
echo "Building CNI"
go build -o ./bin/cndp ./cmd/cndp-cni
echo

echo "***** Unit Tests *****"
echo "Running unit tests:" 
go test $(go list ./... | grep -v "/examples/e2e-test/" | grep -v "/pkg/resourcesapi")
echo

echo "***** Update CLOC *****"
if hash cloc 2>/dev/null; then
	cloc $(git ls-files)
	sed -i "/<\!---clocstart--->/,/<\!---clocend--->/c\<\!---clocstart--->\n\`\`\`\n$(cloc $(git ls-files) | sed -n '/-----/,$p' | sed -z 's/\n/\\n/g')\n\`\`\`\n\<\!---clocend--->" README.md
else
	echo "CLOC not installed, skipping"
fi

echo "Build complete!"
