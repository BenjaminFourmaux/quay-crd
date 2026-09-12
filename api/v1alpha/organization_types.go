package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// OrganizationSpec defines the desired state of Organization
type OrganizationSpec struct {
	Email          *string `json:"email,omitempty"`
	InvoiceEmail   *string `json:"invoice_email,omitempty"`
	TagExpirationS *int    `json:"tag_expiration_s,omitempty"`
}

// OrganizationStatus defines the observed state of Organization.
type OrganizationStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Organization is the Schema for the organizations API
type Organization struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Organization
	// +required
	Spec OrganizationSpec `json:"spec"`

	// status defines the observed state of Organization
	// +optional
	Status OrganizationStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// OrganizationList contains a list of Organization
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Organization `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &Organization{}, &OrganizationList{})
		return nil
	})
}
