// +build !linux

/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package kubelet

import "fmt"

// A Kubelet to flannel bridging helper.
type flannelHelper struct{}

// NewFlannelHelper creates a new flannel helper.
func NewFlannelHelper() FlannelHelper {
	return &flannelHelper{}
}

// Handshake waits for the flannel subnet file and installs a few IPTables
// rules, returning the pod CIDR allocated for this node.
func (f *flannelHelper) Handshake() (podCIDR string, err error) {
	return "", fmt.Errorf("not supported")
}
