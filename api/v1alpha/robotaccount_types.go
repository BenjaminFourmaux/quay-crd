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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RobotAccountSpec defines the desired state of RobotAccount
type RobotAccountSpec struct {
	// Name of the RobotAccount (shortname)
	// +required
	Name string `json:"name"`

	// Description of the RobotAccount
	// +optional
	Description string `json:"description"`

	// Owner of Quay object
	// +required
	Owner RobotAccountOwner `json:"owner"`

	// Secret enables storage of Docker credentials in a Kubernetes Secret.
	// Its name defaults to a name derived from the full Quay robot account name.
	// +optional
	Secret *RobotAccountSecret `json:"secret,omitempty"`

	// TODO: select all repo per default ??
}

type RobotAccountOwner struct {
	// Kind of the Quay object
	// +required
	// +kubebuilder:validation:Enum=user;organization
	Kind string `json:"kind"`

	// Name of the kind object
	// +required
	Name string `json:"name"`
}

type RobotAccountSecret struct {
	// Name of the Secret. Defaults to the normalized Quay robot name with a -pull-secret suffix.
	// +optional
	Name string `json:"name,omitempty"`

	// Namespace of the Secret. Defaults to the RobotAccount namespace.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// RobotAccountStatus defines the observed state of RobotAccount.
type RobotAccountStatus struct {
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RobotAccount is the Schema for the robotaccounts API
type RobotAccount struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of RobotAccount
	// +required
	Spec RobotAccountSpec `json:"spec"`

	// status defines the observed state of RobotAccount
	// +optional
	Status RobotAccountStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// RobotAccountList contains a list of RobotAccount
type RobotAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []RobotAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &RobotAccount{}, &RobotAccountList{})
		return nil
	})
}
