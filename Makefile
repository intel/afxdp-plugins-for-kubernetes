# Copyright(c) 2022 Intel Corporation.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

excluded_from_utests = "/test/e2e|/test/fuzz"

.PHONY: all e2e

all: format build test static

format:
	@echo "******     Go Format     ******"
	@echo
	-go fmt github.com/intel/afxdp-plugins-for-kubernetes/...
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

builddp: buildc
	@echo "******     Build DP      ******"
	@echo
	go build -o ./bin/afxdp-dp ./cmd/deviceplugin
	@echo
	@echo

buildcni: buildc
	@echo "******     Build CNI     ******"
	@echo
	go build -o ./bin/afxdp ./cmd/cni
	@echo
	@echo

build: builddp buildcni

docker:
	@echo "******   Docker Image    ******"
	@echo
	docker build -t afxdp-device-plugin -f images/amd64.dockerfile .
	@echo
	@echo 

podman:
	@echo "******   Podman Image    ******"
	@echo
	podman build -t afxdp-device-plugin -f images/amd64.dockerfile .
	@echo
	@echo

image:
	if $(MAKE) podman; then \
	 echo "Podman build succeeded"; \
	else \
	 echo "Podman build failed, trying docker.."; \
	 $(MAKE) docker; \
	fi

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
	go test $(shell go list ./... | grep -vE $(excluded_from_utests) | grep -v "/internal/resourcesapi")
	@echo
	@echo

e2e: build
	@echo "******     Basic E2E     ******"
	@echo
	cd test/e2e/ && ./e2e-test.sh
	@echo
	@echo

e2efull: build
	@echo "******     Full E2E      ******"
	@echo
	cd test/e2e/ && ./e2e-test.sh --full
	@echo
	@echo

e2edaemon: image
	@echo "******   E2E Daemonset   ******"
	@echo
	cd test/e2e/ && ./e2e-test.sh --daemonset
	@echo
	@echo

e2efulldaemon: image
	@echo "****** Full E2E DaemSet  ******"
	@echo
	cd test/e2e/ && ./e2e-test.sh --full --daemonset
	@echo
	@echo

static-ci: 
	@echo "******   Verify dependencies   ******"
	@echo
	go mod verify
	@echo
	@echo
	@echo "******   Run staticcheck   ******"
	@echo
	staticcheck ./...
	@echo
	@echo
	@echo "******      Go Vet       ******"
	@echo
	for pkg in $$(go list github.com/intel/afxdp-plugins-for-kubernetes/...); do echo $$pkg && go vet $$pkg; done
	@echo
	@echo

static: static-ci
	@echo "******   GolangCI-Lint   ******"
	@echo
	golangci-lint run
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
	@echo
	@echo
	@echo "******       Trivy       ******"
	@echo
	trivy image afxdp-device-plugin --no-progress --format json
	trivy fs . --no-progress --format json
	@echo
	@echo

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
	rm -f ./bin/afxdp
	rm -f ./bin/afxdp-dp
	rm -f ./internal/bpf/bpfWrapper.o
	rm -f ./internal/bpf/libwrapper.a
	@echo
	@echo
