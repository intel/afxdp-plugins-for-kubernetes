#!/bin/bash

pids=( )
run_dp="./../../bin/cndp-dp"

cleanup() {
	echo
	echo "*****************************************************"
	echo "*                     Cleanup                       *"
	echo "*****************************************************"
	echo "Delete Sample App"
	rm -f uds-client &> /dev/null
	echo "Delete CNI"
	rm -f /opt/cni/bin/cndp-e2e &> /dev/null
	echo "Delete Network Attachment Definition"
	kubectl delete network-attachment-definition cndp-e2e-test &> /dev/null
	echo "Delete Pod"
	kubectl delete pod cndp-e2e-test &> /dev/null
	echo "Delete Docker Image"
	docker rmi cndp-e2e-test &> /dev/null
	echo "Stop Device Plugin"
	(( ${#pids[@]} )) && kill "${pids[@]}" #if we have a saved DP PID, kill it
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
	echo "***** Sample App *****"
	go build -o uds-client ./main.go
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
	echo "*                 Run DP and Pod                    *"
	echo "*****************************************************"
	$run_dp & pids+=( "$!" ) #run the DP and save the PID
	sleep 10
	kubectl create -f pod.yaml
	sleep 10
	echo
	echo "*****************************************************"
	echo "*              Netdevs attached to pod              *"
	echo "*****************************************************"
	kubectl exec -i cndp-e2e-test ip a
	sleep 2
	echo
	echo "*****************************************************"
	echo "*                     UDS Test                      *"
	echo "*****************************************************"
	kubectl exec -i cndp-e2e-test /cndp/uds-client <<< $'end to end test\nexit\n'
	sleep 2
}

cleanup  
build
run
trap cleanup EXIT
