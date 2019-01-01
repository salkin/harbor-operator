package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HarborSpec defines the desired state of Harbor
type HarborSpec struct {
	Version  string `json:"version"`
	Config   `json:"config"`
	Registry `json:"registry"`
}

// HarborStatus defines the observed state of Harbor
type HarborStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Harbor is the Schema for the harbors API
// +k8s:openapi-gen=true
type Harbor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborSpec   `json:"spec,omitempty"`
	Status HarborStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HarborList contains a list of Harbor
type HarborList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Harbor `json:"items"`
}

type Config struct {
	ExtURL       string `json:"extURL"`
	LogLevel     string `json:"logLevel"`
	StorageClass string `json:"storageClass,omitempty"`
}

type Registry struct {
	Storage struct {
		Size         string `json:"size"`
		StorageClass string `json:"storageClass"`
	} `json:"storage"`
}

func init() {
	SchemeBuilder.Register(&Harbor{}, &HarborList{})
}
