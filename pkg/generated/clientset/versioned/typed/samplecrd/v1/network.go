/*
Copyright The Kubernetes Authors.

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

package v1

import (
	context "context"
	samplecrdv1 "k8s-custom-crd-controller/pkg/apis/samplecrd/v1"
	scheme "k8s-custom-crd-controller/pkg/generated/clientset/versioned/scheme"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// NetworksGetter has a method to return a NetworkInterface.
// A group's client should implement this interface.
type NetworksGetter interface {
	Networks(namespace string) NetworkInterface
}

// NetworkInterface has methods to work with Network resources.
type NetworkInterface interface {
	Create(ctx context.Context, network *samplecrdv1.Network, opts metav1.CreateOptions) (*samplecrdv1.Network, error)
	Update(ctx context.Context, network *samplecrdv1.Network, opts metav1.UpdateOptions) (*samplecrdv1.Network, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, network *samplecrdv1.Network, opts metav1.UpdateOptions) (*samplecrdv1.Network, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*samplecrdv1.Network, error)
	List(ctx context.Context, opts metav1.ListOptions) (*samplecrdv1.NetworkList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *samplecrdv1.Network, err error)
	NetworkExpansion
}

// networks implements NetworkInterface
type networks struct {
	*gentype.ClientWithList[*samplecrdv1.Network, *samplecrdv1.NetworkList]
}

// newNetworks returns a Networks
func newNetworks(c *SamplecrdV1Client, namespace string) *networks {
	return &networks{
		gentype.NewClientWithList[*samplecrdv1.Network, *samplecrdv1.NetworkList](
			"networks",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *samplecrdv1.Network { return &samplecrdv1.Network{} },
			func() *samplecrdv1.NetworkList { return &samplecrdv1.NetworkList{} },
		),
	}
}
