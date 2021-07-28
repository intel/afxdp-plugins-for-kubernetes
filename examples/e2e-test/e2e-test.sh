#!/usr/bin/env bash
set -e

pids=( )
run_dp="./../../bin/cndp-dp"
full_run=false
devices=()

cleanup() {
	echo
	echo "*****************************************************"
	echo "*                     Cleanup                       *"
	echo "*****************************************************"
	echo "Delete Pod"
	kubectl delete pod --ignore-not-found=true cndp-e2e-test &> /dev/null
	echo "Delete Sample Apps"
	rm -f uds-client-auto &> /dev/null
	rm -f uds-client-manual &> /dev/null
	echo "Delete CNI"
	rm -f /opt/cni/bin/cndp-e2e &> /dev/null
	echo "Delete Network Attachment Definition"
	kubectl delete network-attachment-definition --ignore-not-found=true cndp-e2e-test &> /dev/null
	echo "Delete Docker Image"
	docker 2>/dev/null rmi cndp-e2e-test | true #remove docker image, ignore "does not exist error"
	echo "Stop Device Plugin"
	if [ ${#pids[@]} -eq 0 ]; then
		echo "No Device Plugin PID found"
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
	echo
	echo "***** CNI Install *****"
	cp ../../bin/cndp /opt/cni/bin/cndp-e2e
	echo "***** Network Attachment Definition *****"
	kubectl create -f ./nad.yaml
	echo "***** Sample Apps *****"
	go build -o uds-client-auto ./autoTest/main.go
	go build -o uds-client-manual ./manualTest/main.go
	echo "***** Docker Image *****"
	docker build \
		--build-arg http_proxy=${http_proxy} \
		--build-arg HTTP_PROXY=${HTTP_PROXY} \
		--build-arg https_proxy=${https_proxy} \
		--build-arg HTTPS_PROXY=${HTTPS_PROXY} \
		--build-arg no_proxy=${no_proxy} \
		--build-arg NO_PROXY=${NO_PROXY} \
		-t cndp-e2e-test -f Dockerfile .
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
	echo "***** Env vars pod*****"
	echo
	kubectl exec -it cndp-e2e-test -- env
	echo
	echo "***** UDS Test *****"
	echo
	kubectl exec -i cndp-e2e-test --container cndp -- /cndp/uds-client-auto 
	echo "***** Delete Pod *****"
	kubectl delete pod cndp-e2e-test &> /dev/null

	if [ "$full_run" = true ]; then

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
		echo "***** Env vars pod*****"
	  echo
	  kubectl exec -it cndp-e2e-test -- env
	  echo
		echo "***** UDS Test *****"
		echo
		kubectl exec -i cndp-e2e-test --container cndp -- /cndp/uds-client-auto
	  echo
		echo "***** Delete Pod *****"
		kubectl delete pod cndp-e2e-test &> /dev/null

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
		kubectl exec -i cndp-e2e-test --container cndp -- /cndp/uds-client-auto
		echo
		echo "***** UDS Test: Container 2 *****"
		echo
		kubectl exec -i cndp-e2e-test --container cndp2 -- /cndp/uds-client-auto
		echo
		echo "***** Delete Pod *****"
		kubectl delete pod cndp-e2e-test &> /dev/null

	fi

}

display_help() {
	echo "Usage: $0 [option...]"
	echo
	echo "  -h, --help          Print Help (this message) and exit"
	echo "  -f, --full          Multiple runs with multiple containers and multiple devices"
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
			-?*)
				echo "Unknown argument $1"
				exit 1
			;;
			*) break
		esac
		shift
	done
fi

#remove logs from previous run
rm /var/log/cndp-cni-e2e.log
rm /var/log/cndp-dp-e2e.log

cleanup
build
run
trap cleanup EXIT
