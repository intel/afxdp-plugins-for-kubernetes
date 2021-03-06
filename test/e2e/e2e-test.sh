#!/usr/bin/env bash

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

set -e

pids=( )
run_dp="./../../bin/afxdp-dp"
full_run=false
daemonset=false
ci_run=false

cleanup() {
	echo
	echo "*****************************************************"
	echo "*                     Cleanup                       *"
	echo "*****************************************************"
	echo "Delete Pod"
	kubectl delete pod --grace-period 0 --ignore-not-found=true cndp-e2e-test &> /dev/null
	echo "Delete Test App"
	rm -f ./udsTest &> /dev/null
	echo "Delete CNI"
	rm -f /opt/cni/bin/afxdp-e2e &> /dev/null
	echo "Delete Network Attachment Definition"
	kubectl delete network-attachment-definition --ignore-not-found=true cndp-e2e-test &> /dev/null
	echo "Delete Docker Image"
	docker 2>/dev/null rmi cndp-e2e-test || true
	echo "Stop Device Plugin on host (if running)"
	if [ ${#pids[@]} -eq 0 ]; then
		echo "No Device Plugin PID found on host"
	else
		echo "Found Device Plugin PID. Stopping..."
		(( ${#pids[@]} )) && kill "${pids[@]}"
	fi
	echo "Stop Daemonset Device Plugin (if running)"
	kubectl delete --ignore-not-found=true -f ./daemonset.yml
}

build() {
	echo
	echo "*****************************************************"
	echo "*               Build and Install                   *"
	echo "*****************************************************"
	echo
	echo "***** CNI Install *****"
	cp ./../../bin/afxdp /opt/cni/bin/afxdp-e2e
	echo "***** Network Attachment Definition *****"
	kubectl create -f ./nad.yaml
	echo "***** Test App *****"
	go build -tags netgo -o udsTest ./udsTest.go
	echo "***** Docker Image *****"
	docker build -t cndp-e2e-test -f Dockerfile .
}

run() {
	echo
	echo "*****************************************************"
	echo "*              Run Device Plugin                    *"
	echo "*****************************************************"
	if [ "$daemonset" = true ]; then
		if [ "$ci_run" = true ]; then
			echo "*****   Pushing image to registry    *****"
			echo
			docker tag afxdp-device-plugin "$DOCKER_REG"/test/afxdp-device-plugin-e2e:latest
			docker push "$DOCKER_REG"/test/afxdp-device-plugin-e2e:latest
			echo "***** Deploying Device Plugin as daemonset *****"
			echo
			echo "Note that device plugin logs will not be printed to screen on a daemonset run"
			echo "Logs can be viewed separately in /var/log/afxdp-k8s-plugins/cndp-dp-e2e.log"
			echo
			envsubst < ../../.github/daemonset.yml | kubectl apply -f -
			echo "Sleeping for 20 seconds to allow image pull on worker nodes"
			sleep 20
		else
			echo "***** Deploying Device Plugin as daemonset *****"
			echo
			echo "Note that device plugin logs will not be printed to screen on a daemonset run"
			echo "Logs can be viewed separately in /var/log/afxdp-k8s-plugins/cndp-dp-e2e.log"
			echo
			kubectl create -f ./daemonset.yml
		fi
	else
		echo "***** Starting Device Plugin as host binary *****"
		echo
		$run_dp & pids+=( "$!" ) #run the DP and save the PID
	fi
	sleep 10

	echo
	echo "*****************************************************"
	echo "*          Run Pod: 1 container, 1 device           *"
	echo "*****************************************************"
	kubectl create -f pod-1c1d.yaml
	sleep 10
	echo
	echo "***** Netdevs attached to pod (ip a) *****"
	echo
	kubectl exec -i cndp-e2e-test -- ip a
	sleep 2
	echo
	echo "***** Netdevs attached to pod (ip l) *****"
	echo
	kubectl exec -i cndp-e2e-test -- ip l
	sleep 2
	echo
	echo "***** Pod Env Vars *****"
	echo
	kubectl exec -i cndp-e2e-test -- env
	echo
	echo "***** UDS Test *****"
	echo
	kubectl exec -i cndp-e2e-test --container cndp -- udsTest
	echo "***** Delete Pod *****"
	kubectl delete pod --grace-period 0 --ignore-not-found=true cndp-e2e-test &> /dev/null

	if [ "$full_run" = true ]; then
		sleep 5
		echo
		echo "*****************************************************"
		echo "*          Run Pod: 1 container, 2 device           *"
		echo "*****************************************************"
		kubectl create -f pod-1c2d.yaml
		sleep 10
		echo
		echo "***** Netdevs attached to pod (ip a) *****"
		echo
		kubectl exec -i cndp-e2e-test -- ip a
		sleep 2
		echo
		echo "***** Netdevs attached to pod (ip l) *****"
		echo
		kubectl exec -i cndp-e2e-test -- ip l
		sleep 2
		echo
		echo "***** Pod Env Vars *****"
		echo
		kubectl exec -i cndp-e2e-test -- env
		echo
		echo "***** UDS Test *****"
		echo
		kubectl exec -i cndp-e2e-test -- udsTest
		echo
		echo "***** Delete Pod *****"
		kubectl delete pod --grace-period 0 --ignore-not-found=true cndp-e2e-test &> /dev/null
		sleep 5
		echo
		echo "*****************************************************"
		echo "*       Run Pod: 2 containers, 1 device each        *"
		echo "*****************************************************"
		kubectl create -f pod-2c2d.yaml
		sleep 10
		echo
		echo "***** Netdevs attached to pod (ip a) *****"
		echo
		kubectl exec -i cndp-e2e-test -- ip a
		sleep 2
		echo
		echo "***** Netdevs attached to pod (ip l) *****"
		echo
		kubectl exec -i cndp-e2e-test -- ip l
		sleep 2
		echo
		echo "***** Env vars container 1 *****"
		echo
		kubectl exec -i cndp-e2e-test --container cndp -- env
		echo
		echo "***** Env vars container 2 *****"
		echo
		kubectl exec -i cndp-e2e-test --container cndp2 -- env
		echo
		echo "***** UDS Test: Container 1 *****"
		echo
		kubectl exec -i cndp-e2e-test --container cndp -- udsTest
		echo
		echo "***** UDS Test: Container 2 *****"
		echo
		kubectl exec -i cndp-e2e-test --container cndp2 -- udsTest
		echo
		echo "***** Delete Pod *****"
		kubectl delete pod --grace-period 0 --ignore-not-found=true cndp-e2e-test &> /dev/null
		sleep 5
		echo
		echo "*****************************************************"
		echo "*          Run Pod: Timeout (never connect)         *"
		echo "*****************************************************"
		echo "***** Expect Timeout Execution *****"
		kubectl create -f pod-1c1d.yaml
		sleep 10
		echo
		echo "***** UDS Test *****"
		echo
		kubectl exec -i cndp-e2e-test --container cndp -- udsTest -timeout-before-connect
		echo
		echo "***** Delete Pod *****"
		kubectl delete pod --grace-period 0 --ignore-not-found=true cndp-e2e-test &> /dev/null
		sleep 5
		echo
		echo "******************************************************************"
		echo "*          Run Pod: Timeout (after connect)                      *"
		echo "******************************************************************"
		echo "***** Expect Timeout Execution *****"
		kubectl create -f pod-1c1d.yaml
		sleep 10
		echo
		echo "***** UDS Test *****"
		echo
		kubectl exec -i cndp-e2e-test --container cndp -- udsTest -timeout-after-connect
		echo
		echo "***** Delete Pod *****"
		kubectl delete pod --grace-period 0 --ignore-not-found=true cndp-e2e-test &> /dev/null

	fi
}

display_help() {
	echo "Usage: $0 [option...]"
	echo
	echo "  -h, --help          Print Help (this message) and exit"
	echo "  -f, --full          Multiple runs with multiple containers and multiple devices"
	echo "  -d, --daemonset     Deploy the device plugin in a daemonset"
	echo
	exit 0
}

if [ -n "${1-}" ]
then
	while :; do
		case $1 in
			-h|--help)
				display_help
			;;
			-f|--full)
				full_run=true
			;;
			-d|--daemonset)
				daemonset=true
			;;
			-c|--ci)
				ci_run=true
			;;
			-?*)
				echo "Unknown argument $1"
				exit 1
			;;
			*) break
		esac
		shift
	done
fi

cleanup
build
run
trap cleanup EXIT
