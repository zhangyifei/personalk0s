/*
Copyright 2022 eke authors

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

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	scheme "eke/pkg/apis/eke/clientset/scheme"
	v1beta1 "eke/pkg/apis/eke/v1beta1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rest "k8s.io/client-go/rest"
)

// ClusterConfigListsGetter has a method to return a ClusterConfigListInterface.
// A group's client should implement this interface.
type ClusterConfigListsGetter interface {
	ClusterConfigLists(namespace string) ClusterConfigListInterface
}

// ClusterConfigListInterface has methods to work with ClusterConfigList resources.
type ClusterConfigListInterface interface {
	Create(ctx context.Context, clusterConfigList *v1beta1.ClusterConfigList, opts v1.CreateOptions) (*v1beta1.ClusterConfigList, error)
	ClusterConfigListExpansion
}

// clusterConfigLists implements ClusterConfigListInterface
type clusterConfigLists struct {
	client rest.Interface
	ns     string
}

// newClusterConfigLists returns a ClusterConfigLists
func newClusterConfigLists(c *EkeV1beta1Client, namespace string) *clusterConfigLists {
	return &clusterConfigLists{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a clusterConfigList and creates it.  Returns the server's representation of the clusterConfigList, and an error, if there is any.
func (c *clusterConfigLists) Create(ctx context.Context, clusterConfigList *v1beta1.ClusterConfigList, opts v1.CreateOptions) (result *v1beta1.ClusterConfigList, err error) {
	result = &v1beta1.ClusterConfigList{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("clusterconfiglists").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(clusterConfigList).
		Do(ctx).
		Into(result)
	return
}
