package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Network describes a Network resource
type Network struct {
	// TypeMeta is the metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            NetworkStatus `json:"status"`

	// Spec is the custom resource spec
	Spec NetworkSpec `json:"spec"`
}

// FooStatus is the status for a Foo resource
type NetworkStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// NetworkSpec is the spec for a Network resource
type NetworkSpec struct {
	// Cidr and Gateway are example custom spec fields
	//
	// this is where you would put your custom resource data
	DeploymentName string `json:"deploymentName"`
	Replicas       *int32 `json:"replicas"`
	Cidr           string `json:"cidr"`
	Gateway        string `json:"gateway"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkList is a list of Network resources
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Network `json:"items"`
}
