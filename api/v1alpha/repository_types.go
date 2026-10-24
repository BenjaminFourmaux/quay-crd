/*
Copyright 2026.

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

package v1alpha

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RepositorySpec defines the desired state of Repository
type RepositorySpec struct {
	// Name of the repository
	// +required
	Name string `json:"name"`

	// Description of the repository
	// +optional
	Description string `json:"description,omitempty"`

	// Visibility of the repository
	// +optional
	// +kubebuilder:validation:Enum=public;private
	// +default="private"
	Visibility string `json:"visibility,omitempty"`

	// State of the repository
	// +optional
	// +kubebuilder:validation:Enum=normal;readonly;mirror
	// +default="normal"
	State string `json:"state,omitempty"`

	// Mirroring parameters
	// +optional
	Mirroring RepositoryMirroring `json:"mirroring,omitempty"`
}

type RepositoryMirroring struct {
	// RegistryLocation Address of the repository to mirror
	// +required
	RegistryLocation string `json:"registryLocation"`

	// Tags to mirror. Wildcard accepted
	// +required
	// +example=["latest","1.0.0", "3.3*","*"]
	Tags []string `json:"tags"`

	// StartDate to sync images
	// +optional
	StartDate string `json:"startDate,omitempty"`

	// SyncInterval to sync images. s for seconds, m for minutes, h for hours, d for days, w for weeks
	// +optional
	// +example="30m | 10s | 1d | 1w"
	SyncInterval string `json:"syncInterval,omitempty"`

	// RobotUser name of the robot account to use for sync actions
	// +optional
	RobotUser string `json:"robotUser,omitempty"`

	// Credentials for login in to the repository if not public
	// +optional
	Credentials RepositoryMirroringCredentials `json:"credentials,omitempty"`

	// VerifyTLS
	// +optional
	// +default=false
	VerifyTLS bool `json:"verifyTLS,omitempty"`

	// AcceptUnsignedImages
	// +optional
	// +default=false
	AcceptUnsignedImages bool `json:"acceptUnsignedImages,omitempty"`

	// HttpProxy
	// +optional
	HttpProxy string `json:"httpProxy,omitempty"`

	// HttpsProxy
	// +optional
	HttpsProxy string `json:"httpsProxy,omitempty"`

	// NoProxy
	// +optional
	NoProxy string `json:"noProxy,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.username) != has(self.usernameFrom)) && (has(self.password) != has(self.passwordFrom))",message="specify either a literal value or a valueFrom source for both username and password"
type RepositoryMirroringCredentials struct {
	// Username for the source registry
	// +optional
	Username string `json:"username,omitempty"`

	// Password for the source registry
	// +optional
	Password string `json:"password,omitempty"`

	// UsernameFrom references a Secret or ConfigMap key containing the username
	// +optional
	UsernameFrom *RepositoryCredentialSource `json:"usernameFrom,omitempty"`

	// PasswordFrom references a Secret or ConfigMap key containing the password
	// +optional
	PasswordFrom *RepositoryCredentialSource `json:"passwordFrom,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.secretKeyRef) && !has(self.configMapKeyRef)) || (!has(self.secretKeyRef) && has(self.configMapKeyRef))",message="specify either secretKeyRef or configMapKeyRef"
type RepositoryCredentialSource struct {
	// SecretKeyRef references a key in a Secret in the Repository namespace
	// +optional
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`

	// ConfigMapKeyRef references a key in a ConfigMap in the Repository namespace
	// +optional
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
}

// RepositoryStatus defines the observed state of Repository.
type RepositoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the Repository resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Repository is the Schema for the repositories API
type Repository struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Repository
	// +required
	Spec RepositorySpec `json:"spec"`

	// status defines the observed state of Repository
	// +optional
	Status RepositoryStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// RepositoryList contains a list of Repository
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Repository `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &Repository{}, &RepositoryList{})
		return nil
	})
}
