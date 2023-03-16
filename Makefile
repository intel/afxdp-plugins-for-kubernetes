# Copyright(c) 2022 Intel Corporation.
# Copyright(c) Red Hat Inc.
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

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

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

buildxdp:
	@echo "******     Build xdp_pass     ******"
	make -C ./internal/bpf/xdp-pass/
	@echo "******     Build xdp_afxdp_redirect     ******"
	make -C ./internal/bpf/xdp-afxdp-redirect/
	@echo

buildc:
	@echo "******     Build BPF     ******"
	@echo
	gcc ./internal/bpf/bpfWrapper.c -lbpf -c -o ./internal/bpf/bpfWrapper.o
	ar rs ./internal/bpf/libwrapper.a ./internal/bpf/bpfWrapper.o  &> /dev/null
	@echo
	@echo

builddp: buildc buildxdp
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

##@ General Build - assumes K8s environment is already setup
docker: ## Build docker image
	@echo "******  Docker Image    ******"
	@echo
	docker build -t afxdp-device-plugin -f images/amd64.dockerfile .
	@echo
	@echo

podman: ## Build podman image
	@echo "******  Podman Image    ******"
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

undeploy: ## Undeploy the Deamonset
	@echo "******  Stop Daemonset   ******"
	@echo
	kubectl delete -f ./deployments/daemonset.yml --ignore-not-found=true
	@echo
	@echo

deploy: image undeploy ## Deploy the Deamonset and CNI
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

# static-ci: consists of static analysis tools required for the public CI
# repository workflow /.github/workflows/public-ci.yml
# Note: the public repository CI comprises of further static analysis tools via the
# superlinter job: golangci-lint, hadolint, clang-format and shellcheck

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

# static: consists of static analysis tools required for internal CI repository workflows and locally
# run tests. static includes static-ci test module.
static: static-ci
	@echo "******   GolangCI-Lint   ******"
	@echo
	golangci-lint run
	@echo
	@echo
	@echo "******     Hadolint      ******"
	@echo
	for file in $$(find . -type f -iname "*dockerfile*" -not -path "./.git/*"); do echo $$file && docker run --rm -i hadolint/hadolint < $$file; done
	@echo
	@echo
	@echo "******    Shellcheck     ******"
	@echo
	for file in $$(find . -iname "*.sh" -not -path "./.git/*"); do echo $$file && shellcheck $$file; done
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

##@ General setup

.PHONY: setup-flannel
setup-flannel: ## Setup flannel
	kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml

.PHONY: setup-multus
setup-multus: ## Setup multus
	kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset.yml

##@ Kind Deployment - sets up a kind cluster and deploys the plugin and CNI

.PHONY: del-kind
del-kind: ## Remove a kind cluster called af-xdp-deployment
	kind delete cluster --name af-xdp-deployment

.PHONY: setup-kind
setup-kind: del-kind ## Setup a kind cluster called af-xdp-deployment
	mkdir -p /tmp/afxdp_dp/
	mkdir -p /tmp/afxdp_dp2/
	kind create cluster --config hack/kind-config.yaml --name af-xdp-deployment

.PHONY: label-kind-nodes
label-kind-nodes: ## label the kind worker nodes with cndp="true"
	kubectl label node af-xdp-deployment-worker cndp="true"
	kubectl label node af-xdp-deployment-worker2 cndp="true"

.PHONY: kind-deploy
kind-deploy: image undeploy ## Deploy the Deamonset and CNI in Kind
	@echo "****** Deploy Daemonset  ******"
	@echo
	kind load --name af-xdp-deployment docker-image afxdp-device-plugin
	kubectl create -f ./deployments/daemonset-kind.yaml
	@echo
	@echo

.PHONY: run-on-kind
run-on-kind: del-kind setup-kind label-kind-nodes setup-multus kind-deploy ## Setup a kind cluster and deploy the device plugin
	@echo "******       Kind Setup complete       ******"
