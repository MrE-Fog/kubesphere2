/*
Copyright 2020 The KubeSphere Authors.

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

package filters

import (
	"context"
	"errors"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

type authzFilter struct {
	next http.Handler
	authorizer.Authorizer
	serializer runtime.NegotiatedSerializer
}

// WithAuthorization passes all authorized requests on to handler, and returns forbidden error otherwise.
func WithAuthorization(next http.Handler, authorizers authorizer.Authorizer) http.Handler {
	if authorizers == nil {
		klog.Warningf("Authorization is disabled")
		return next
	}

	return &authzFilter{
		next:       next,
		Authorizer: authorizers,
		serializer: serializer.NewCodecFactory(runtime.NewScheme()).WithoutConversion(),
	}
}

func (a *authzFilter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	attributes, err := getAuthorizerAttributes(ctx)
	if err != nil {
		responsewriters.InternalError(w, req, err)
	}

	authorized, reason, err := a.Authorize(attributes)
	if authorized == authorizer.DecisionAllow {
		a.next.ServeHTTP(w, req)
		return
	}

	if err != nil {
		responsewriters.InternalError(w, req, err)
		return
	}

	klog.V(4).Infof("Forbidden: %#v, Reason: %q", req.RequestURI, reason)
	responsewriters.Forbidden(ctx, attributes, w, req, reason, a.serializer)
}

func getAuthorizerAttributes(ctx context.Context) (authorizer.Attributes, error) {
	attribs := authorizer.AttributesRecord{}

	user, ok := request.UserFrom(ctx)
	if ok {
		attribs.User = user
	}

	requestInfo, found := request.RequestInfoFrom(ctx)
	if !found {
		return nil, errors.New("no RequestInfo found in the context")
	}

	// Start with common attributes that apply to resource and non-resource requests
	attribs.ResourceScope = requestInfo.ResourceScope
	attribs.ResourceRequest = requestInfo.IsResourceRequest
	attribs.Path = requestInfo.Path
	attribs.Verb = requestInfo.Verb
	attribs.Cluster = requestInfo.Cluster
	attribs.Workspace = requestInfo.Workspace
	attribs.KubernetesRequest = requestInfo.IsKubernetesRequest
	attribs.APIGroup = requestInfo.APIGroup
	attribs.APIVersion = requestInfo.APIVersion
	attribs.Resource = requestInfo.Resource
	attribs.Subresource = requestInfo.Subresource
	attribs.Namespace = requestInfo.Namespace
	attribs.DevOps = requestInfo.DevOps
	attribs.Name = requestInfo.Name

	return &attribs, nil
}
