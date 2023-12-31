/*
Copyright 2020 KubeSphere Authors

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

package v1

import (
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"kubesphere.io/kubesphere/pkg/kapis/generic"
)

// there are no versions specified cause we want to proxy all versions of requests to backend service
// Deprecated: this api group would be removed in v4.0
var GroupVersion = schema.GroupVersion{Group: "kubeedge.kubesphere.io", Version: ""}

func AddToContainer(container *restful.Container, endpoint string) error {
	proxy, err := generic.NewGenericProxy(endpoint, GroupVersion.Group, GroupVersion.Version)
	if err != nil {
		return nil
	}

	proxy.SetProxyDesprecated()

	return proxy.AddToContainer(container)
}
