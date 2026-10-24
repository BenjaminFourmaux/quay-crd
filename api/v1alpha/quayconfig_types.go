package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// QuayConfigSpec defines the desired state of QuayConfig
type QuayConfigSpec struct {
	// URL of the Quay API
	// +required
	URL string `json:"url"`

	// Credentials is a reference to a Kubernetes Secret that contains the Quay API credentials
	// +required
	Credentials SecretReference `json:"credentials"`
}

type SecretReference struct {
	// Name of the Kubernetes Secret
	// +required
	Name string `json:"name"`

	// Key of the secret data that contains the Quay API token
	// +required
	Key string `json:"key"`
}

type QuayConfigStatus struct {
	// Conditions represent the current state of the QuayConfig.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Ready indicates whether the Quay API can be reached.
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// QuayConfig is the Schema for the QuayConfig API
type QuayConfig struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of QuayConfig
	// +required
	Spec QuayConfigSpec `json:"spec"`

	// status defines the observed state of QuayConfig
	// +optional
	Status QuayConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// QuayConfigList contains a list of QuayConfig
type QuayConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []QuayConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &QuayConfig{}, &QuayConfigList{})
		return nil
	})
}
