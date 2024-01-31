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

workdir="."
ciWorkdir="./../../.github/e2e"
run_dp="./../../bin/afxdp-dp"
full_run=false
daemonset=false
daemonsetGo=false
daemonsetUDS=false
soak=false
ci_run=false
pids=( )
container_tool=""


detect_container_engine() {
        echo "*****************************************************"
        echo "*          Checking Container Engine                *"
        echo "*****************************************************"
	if podman -v; then \
	  container_tool="podman"; \
	else \
	  container_tool="docker"; \
	fi

	echo "$container_tool recognised as container engine"
}

cleanup() {
	echo
	echo "*****************************************************"
	echo "*                     Cleanup                       *"
	echo "*****************************************************"
	echo "Delete Pod"
	kubectl delete pod --grace-period=0 --ignore-not-found=true afxdp-e2e-test &> /dev/null
	kubectl delete pods -l app=afxdp-e2e -n default --grace-period=0 --ignore-not-found=true &> /dev/null
	echo "Delete Test App"
	rm -f ./udsTest &> /dev/null
	echo "Delete CNI"
	rm -f /opt/cni/bin/afxdp &> /dev/null
	echo "Delete Network Attachment Definition"
	kubectl delete --ignore-not-found=true -f $workdir/nad.yaml
	echo "Delete Docker Image"
	$container_tool 2>/dev/null rmi afxdp-e2e-test || true
	echo "Stop Device Plugin on host (if running)"
	if [ ${#pids[@]} -eq 0 ]; then
		echo "No Device Plugin PID found on host"
	else
		echo "Found Device Plugin PID. Stopping..."
		(( ${#pids[@]} )) && kill "${pids[@]}"
	fi
	echo "Stop Daemonset Device Plugin (if running)"
	kubectl delete --ignore-not-found=true -f $workdir/daemonset.yml
}

build() {
	echo
	echo "*****************************************************"
	echo "*               Build and Install                   *"
	echo "*****************************************************"
	echo
	echo "***** CNI Install *****"
	cp ./../../bin/afxdp /opt/cni/bin/afxdp
	echo "***** Network Attachment Definition *****"
	kubectl create -f $workdir/nad.yaml
	echo "***** Test App *****"
	go build -tags netgo -o udsTest ./udsTest.go
	echo "***** Docker Image *****"
	$container_tool build -t afxdp-e2e-test -f Dockerfile .
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
			$container_tool tag afxdp-device-plugin "$DOCKER_REG"/test/afxdp-device-plugin-e2e:latest
			$container_tool push "$DOCKER_REG"/test/afxdp-device-plugin-e2e:latest
			echo "***** Deploying Device Plugin as daemonset *****"
			echo
			echo "Note that device plugin logs will not be printed to screen on a daemonset run"
			echo "Logs can be viewed separately in /var/log/afxdp-k8s-plugins/afxdp-dp-e2e.log"
			echo
			envsubst < $workdir/daemonset.yml | kubectl apply -f -
			echo "Pausing for 20 seconds to allow image pull on worker nodes"
			sleep 20
		else
			echo "***** Deploying Device Plugin as daemonset *****"
			echo
			echo "Note that device plugin logs will not be printed to screen on a daemonset run"
			echo "Logs can be viewed separately in /var/log/afxdp-k8s-plugins/afxdp-dp-e2e.log"
			echo
			kubectl create -f $workdir/daemonset.yml
		fi
	else
		echo "***** Starting Device Plugin as host binary *****"
		echo
		$run_dp & pids+=( "$!" ) #run the DP and save the PID
	fi
	sleep 10

	while :; do
		if [ "$ci_run" = true ]; then 
			run_ci_pods
		else
			run_local_pods
		fi
		if [ "$soak" = false ]; then break; fi
	done
}

run_local_pods() {
	echo
	echo "*****************************************************"
	echo "*          Run Pod: 1 container, 1 device           *"
	echo "*****************************************************"
	kubectl create -f $workdir/pod-1c1d.yaml
	sleep 10
	echo
	echo "***** Netdevs attached to pod (ip a) *****"
	echo
	kubectl exec -i afxdp-e2e-test -- ip a
	sleep 2
	echo
	echo "***** Netdevs attached to pod (ip l) *****"
	echo
	kubectl exec -i afxdp-e2e-test -- ip l
	sleep 2
	echo
	echo "***** Pod Env Vars *****"
	echo
	kubectl exec -i afxdp-e2e-test -- env
	echo
	if [ "$daemonsetUDS" = true ]; then
		echo "***** UDS Test *****"
		echo
		kubectl exec -i afxdp-e2e-test --container afxdp -- udsTest uds
	elif [ "$daemonsetGo" = true ]; then
		echo "***** GO Library Test *****" 
		echo
		kubectl exec -i afxdp-e2e-test --container afxdp -- udsTest golang
	fi
	echo "***** Delete Pod *****"
	kubectl delete pod --grace-period 0 --ignore-not-found=true afxdp-e2e-test &> /dev/null
	if [ "$full_run" = true ]; then
		sleep 5
		echo
		echo "*****************************************************"
		echo "*          Run Pod: 1 container, 2 device           *"
		echo "*****************************************************"
		kubectl create -f $workdir/pod-1c2d.yaml
		sleep 10
		echo
		echo "***** Netdevs attached to pod (ip a) *****"
		echo
		kubectl exec -i afxdp-e2e-test -- ip a
		sleep 2
		echo
		echo "***** Netdevs attached to pod (ip l) *****"
		echo
		kubectl exec -i afxdp-e2e-test -- ip l
		sleep 2
		echo
		echo "***** Pod Env Vars *****"
		echo
		kubectl exec -i afxdp-e2e-test -- env
		echo
		echo "***** UDS Test *****"
		echo
		kubectl exec -i afxdp-e2e-test -- udsTest
		echo
		echo "***** Delete Pod *****"
		kubectl delete pod --grace-period 0 --ignore-not-found=true afxdp-e2e-test &> /dev/null
		sleep 5
		echo
		echo "*****************************************************"
		echo "*       Run Pod: 2 containers, 1 device each        *"
		echo "*****************************************************"
		kubectl create -f $workdir/pod-2c2d.yaml
		sleep 10
		echo
		echo "***** Netdevs attached to pod (ip a) *****"
		echo
		kubectl exec -i afxdp-e2e-test -- ip a
		sleep 2
		echo
		echo "***** Netdevs attached to pod (ip l) *****"
		echo
		kubectl exec -i afxdp-e2e-test -- ip l
		sleep 2
		echo
		echo "***** Env vars container 1 *****"
		echo
		kubectl exec -i afxdp-e2e-test --container afxdp -- env
		echo
		echo "***** Env vars container 2 *****"
		echo
		kubectl exec -i afxdp-e2e-test --container afxdp2 -- env
		echo
		echo "***** UDS Test: Container 1 *****"
		echo
		kubectl exec -i afxdp-e2e-test --container afxdp -- udsTest
		echo
		echo "***** UDS Test: Container 2 *****"
		echo
		kubectl exec -i afxdp-e2e-test --container afxdp2 -- udsTest
		echo
		echo "***** Delete Pod *****"
		kubectl delete pod --grace-period 0 --ignore-not-found=true afxdp-e2e-test &> /dev/null
		sleep 5
		echo
		echo "*****************************************************"
		echo "*          Run Pod: Timeout (never connect)         *"
		echo "*****************************************************"
		echo "***** Expect Timeout Execution *****"
		kubectl create -f $workdir/pod-1c1d.yaml
		sleep 10
		echo
		echo "***** UDS Test *****"
		echo
		kubectl exec -i afxdp-e2e-test --container afxdp -- udsTest -timeout-before-connect
		echo
		echo "***** Delete Pod *****"
		kubectl delete pod --grace-period 0 --ignore-not-found=true afxdp-e2e-test &> /dev/null
		sleep 5
		echo
		echo "******************************************************************"
		echo "*          Run Pod: Timeout (after connect)                      *"
		echo "******************************************************************"
		echo "***** Expect Timeout Execution *****"
		kubectl create -f $workdir/pod-1c1d.yaml
		sleep 10
		echo
		echo "***** UDS Test *****"
		echo
		kubectl exec -i afxdp-e2e-test --container afxdp -- udsTest -timeout-after-connect
		echo
		echo "***** Delete Pod *****"
		kubectl delete pod --grace-period 0 --ignore-not-found=true afxdp-e2e-test &> /dev/null
	fi
}

run_ci_pods() {
	for podFilePath in "$ciWorkdir"/pod*.yaml;
	do
		runningPods=()
		podFile=$(basename "$podFilePath")

		OLDIFS=$IFS
		IFS=_.
		read -ra ARR <<< "$podFile"
		IFS=$OLDIFS

		containers=${ARR[1]}
		pods=${ARR[2]}

		echo "Spinning up $pods instances of $podFile, which has $containers container(s)"

		for (( i=1; i<=pods; i++ ))
		do
			kubectlOutput=$(kubectl create -f "$podFilePath")
			podName=${kubectlOutput%" created"}
			podName=${podName#"pod/"}
			runningPods+=("$podName")
			echo "$kubectlOutput"
		done

		wait_for_pods_to_start

		echo "*****************************************************"
		echo "*                Kubectl Get Pods                   *"
		echo "*****************************************************"
		kubectl get pods -o wide

		for pod in "${runningPods[@]}"
		do
			echo "*****************************************************"
			echo "*                Pod $pod                *"
			echo "*****************************************************"
			echo "***** Netdevs attached to pod *****"
			echo "ip a"
			kubectl exec -i "$pod" -- ip a
			echo
			echo "ip l"
			kubectl exec -i "$pod" -- ip l

			for (( j=1; j<=containers; j++ ))
			do
				echo "***** Env vars Container $j *****"
				kubectl exec -i "$pod" --container afxdp$j -- env
				echo "***** UDS Test Container $j *****"
				kubectl exec -i "$pod" --container afxdp$j -- cat /tmp/udsTest.txt
				echo
			done
		done
	done

	echo "*****************************************************"
	echo "*                Kubectl Get Pods                   *"
	echo "*****************************************************"
	kubectl get pods -o wide

	echo "*****************************************************"
	echo "*                  Delete Pods                      *"
	echo "*****************************************************"	
	kubectl delete pods -l app=afxdp-e2e -n default --grace-period=0
}

wait_for_pods_to_start() {
	counter=0
	while true
	do
		#starting_pods=( $( kubectl get pods | grep afxdp-e2e | awk '$3 != "Running" {print $1}' ) )
		mapfile -t starting_pods < <(kubectl get pods | grep afxdp-e2e | awk '$3 != "Running" {print $1}')

		if (( ${#starting_pods[@]} == 0 )); then
			echo "All pods have started"
			break
		else
			echo "Waiting for pods to start..."
			counter=$((counter+1))
		fi

		if (( counter > 60 )); then
			echo "Error: Pods took too long to start"

			for pod in "${starting_pods[@]}"
			do
				kubectl describe pod "$pod"
				echo -e "\n\n\n"
			done

			echo "Error: Pods took too long to start"
			cleanup
			exit 1
		fi

		sleep 10
	done
}

display_help() {
	echo "Usage: $0 [option...]"
	echo
	echo "  -h, --help          Print Help (this message) and exit"
	echo "  -f, --full          Multiple pods containers and devices. UDS timeout is tested"
	echo "  -d, --daemonset     Deploy the device plugin in a daemonset"
	echo "  -s, --soak          Continue to create and delete test pods until manually stopped"
	echo "  -c, --ci            Deploy as daemonset and deploy a large number of various test pods"
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
			-g|--golang)
				daemonsetGo=true
			;;
			-u|--uds)
				daemonsetUDS=true
			;;
			-c|--ci)
				ci_run=true
				daemonset=true
				workdir=$ciWorkdir
			;;
			-s|--soak)
				soak=true
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

detect_container_engine
cleanup
build
run
trap cleanup EXIT
