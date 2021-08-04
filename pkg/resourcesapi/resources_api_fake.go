/*
 Copyright(c) 2021 Intel Corporation.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package resourcesapi

import (
	api "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
)

/*
FakeHandler interface extends the Handler interface to provide additional testing methods.
*/
type FakeHandler interface {
	Handler
	CreateFakePod(podName string, namespace string, resourceName string, deviceIds []string)
}

/*
fakeHandler implements the FakeHandler interface.
*/
type fakeHandler struct {
	podName      string
	namespace    string
	resourceName string
	deviceIds    []string
}

/*
NewFakeHandler returns an implementation of the FakeHandler interface.
*/
func NewFakeHandler() FakeHandler {
	return &fakeHandler{}
}

/*
GetPodResources returns a map of pods and associated devices.
In this FakeHandler, it returns a map containing just a single pod for testing against.
This pod does not come from the pod resources API, but instead is configurable through the
CreateFakePod function to give a predetermined response.
*/
func (f *fakeHandler) GetPodResources() (map[string]api.PodResources, error) {
	fakePod := api.PodResources{
		Name:      f.podName,
		Namespace: f.namespace,
		Containers: []*api.ContainerResources{
			{
				Name: "container-01",
				Devices: []*api.ContainerDevices{
					{
						ResourceName: f.resourceName,
						DeviceIds:    f.deviceIds,
					},
				},
			},
		},
	}

	podResourceMap := make(map[string]api.PodResources)
	podResourceMap[f.podName] = fakePod

	return podResourceMap, nil
}

/*
CreateFakePod allows us to configure our own fake pod and its associated devices.
This pods data is what is returned when GetPodResources is called.
Tweaking this pod allows us to test our code against different pods and scenarios.
*/
func (f *fakeHandler) CreateFakePod(podName string, namespace string, resourceName string, deviceIds []string) {
	f.podName = podName
	f.namespace = namespace
	f.resourceName = resourceName
	f.deviceIds = deviceIds
}
