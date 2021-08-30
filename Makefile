.PHONY: all e2e

all: format lint build test cloc end

format:
	@echo "******     Go Format     ******"
	@echo
	go fmt github.com/intel/cndp_device_plugin/...
	@echo
	@echo

	@echo "******   Clang Format    ******"
	@echo
	clang-format -i -style=file pkg/bpf/*.c pkg/bpf/*.h
	@echo
	@echo

lint:
	@echo "******     Go Lint       ******"
	@echo
	golint -set_exit_status ./...
	@echo
	@echo

build:
	@echo "******     Build DP      ******"
	@echo
	gcc ./pkg/bpf/bpfWrapper.c -lbpf -c -o ./pkg/bpf/bpfWrapper.o
	ar rs ./pkg/bpf/libwrapper.a ./pkg/bpf/bpfWrapper.o  &> /dev/null
	go build -o ./bin/cndp-dp ./cmd/cndp-dp
	@echo
	@echo

	@echo "******     Build CNI     ******"
	@echo
	go build -o ./bin/cndp ./cmd/cndp-cni
	@echo
	@echo

test:
	@echo "******    Unit Tests     ******"
	@echo
	go test $(shell go list ./... | grep -v "/examples/e2e-test/" | grep -v "/pkg/resourcesapi")
	@echo
	@echo

e2e: build
	@echo "******    E2e Test     ******"
	@echo
	cd e2e && ./e2e-test.sh
	@echo
	@echo

#install:
#       install to deploy once we have a damenset that can deploy DP and CNI

cloc:
	@echo "******    Update CLOC    ******"
	@echo
	@cloc $(shell git ls-files)
	$$(sed -i "/<\!---clocstart--->/,/<\!---clocend--->/c\<\!---clocstart--->\n\`\`\`\n$$(cloc $$(git ls-files) | sed -n '/-----/,$$p' | sed -z 's/\n/\\n/g')\n\`\`\`\n\<\!---clocend--->" README.md)
	@echo
	@echo

clean:
	@echo "******      Cleanup      ******"
	@echo
	rm -f ./bin/cndp
	rm -f ./bin/cndp-dp
	@echo
	@echo

end:
	@echo "******  Build complete!  ******"
	@echo
	@echo