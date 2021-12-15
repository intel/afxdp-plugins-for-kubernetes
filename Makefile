.PHONY: all e2e

all: build test static

format:
	@echo "******     Go Format     ******"
	@echo
	-go fmt github.com/intel/cndp_device_plugin/...
	@echo
	@echo
	@echo "******   Clang Format    ******"
	@echo
	-clang-format -i -style=file internal/bpf/*.c internal/bpf/*.h
	@echo
	@echo

buildc:
	@echo "******     Build BPF     ******"
	@echo
	gcc ./internal/bpf/bpfWrapper.c -lbpf -c -o ./internal/bpf/bpfWrapper.o
	ar rs ./internal/bpf/libwrapper.a ./internal/bpf/bpfWrapper.o  &> /dev/null
	@echo
	@echo

build: format buildc
	@echo "******     Build DP      ******"
	@echo
	go build -o ./bin/cndp-dp ./cmd/cndp-dp
	@echo
	@echo
	@echo "******     Build CNI     ******"
	@echo
	go build -o ./bin/cndp ./cmd/cni/main
	@echo
	@echo

image: build
	@echo "******   Docker Image    ******"
	@echo
	docker build -t cndp-device-plugin -f images/amd64.dockerfile .
	@echo
	@echo

undeploy:
	@echo "******  Stop Daemonset   ******"
	@echo
	kubectl delete -f ./deployments/daemonset.yml --ignore-not-found=true
	@echo
	@echo

deploy: image undeploy
	@echo "****** Deploy Daemonset  ******"
	@echo
	kubectl create -f ./deployments/daemonset.yml
	@echo
	@echo

test: buildc
	@echo "******    Unit Tests     ******"
	@echo
	go test $(shell go list ./... | grep -v "/e2e" | grep -v "/internal/resourcesapi")
	@echo
	@echo

e2e: build
	@echo "******     E2e Test      ******"
	@echo
	cd test/e2e/ && ./e2e-test.sh
	@echo
	@echo

fuzzcni:
	@echo "******     Fuzz CNI      ******"
	@echo
	cd test/fuzz/cni && ./fuzz.sh
	@echo
	@echo

static:
	@echo "******      Go Lint      ******"
	@echo
	golint -set_exit_status ./...
	@echo
	@echo
	@echo "******   GolangCI-Lint   ******"
	@echo
	golangci-lint run
	@echo
	@echo
	@echo "******      Go Vet       ******"
	@echo
	for pkg in $$(go list github.com/intel/cndp_device_plugin/...); do echo $$pkg && go vet $$pkg; done
	@echo
	@echo
	@echo "******     Hadolint      ******"
	@echo
	for file in $$(find . -type f -iname "*dockerfile*"); do echo $$file && docker run --rm -i hadolint/hadolint < $$file; done
	@echo
	@echo
	@echo "******    Shellcheck     ******"
	@echo
	for file in $$(find . -iname "*.sh"); do echo $$file && shellcheck $$file; done
	

cloc: format
	@echo "******    Update CLOC    ******"
	@echo
	@cloc $(shell git ls-files)
	sed -i "/<\!---clocstart--->/,/<\!---clocend--->/c\<\!---clocstart--->\n\`\`\`\n$$(cloc $$(git ls-files) | sed -n '/-----/,$$p' | sed -z 's/\n/\\n/g')\n\`\`\`\n\<\!---clocend--->" README.md
	@echo
	@echo

clean:
	@echo "******      Cleanup      ******"
	@echo
	rm -f ./bin/cndp
	rm -f ./bin/cndp-dp
	rm -f ./internal/bpf/bpfWrapper.o
	rm -f ./internal/bpf/libwrapper.a
	@echo
	@echo
