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
run_dp="./../../../../bin/afxdp-dp"

cleanup() {
	echo
	echo "*****************************************************"
	echo "*                     Cleanup                       *"
	echo "*****************************************************"
	echo "Delete Pod"
	kubectl delete pod --grace-period 0 --ignore-not-found=true cndp-e2e-test &> /dev/null
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
}

build() {
	echo
	echo "*****************************************************"
	echo "*               Build and Install                   *"
	echo "*****************************************************"
	echo "***** CNI Install *****"
	cp ./../../../../bin/afxdp /opt/cni/bin/afxdp-e2e
	echo "***** Network Attachment Definition *****"
	kubectl create -f ./nad.yaml
}

run() {
	echo
	echo "*****************************************************"
	echo "*              Run Device Plugin                    *"
	echo "*****************************************************"
		$run_dp & pids+=( "$!" ) #run the DP and save the PID
	sleep 10

	echo
	echo "*****************************************************"
	echo "*          Run Pod: 1 container, 1 device           *"
	echo "*****************************************************"
	echo "CNDP fuzz testing will be executed after pod is created..."
	kubectl create -f pod-1c1d.yaml
}

cleanup
build
run
trap cleanup EXIT
