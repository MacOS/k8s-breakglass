/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterAccessReviewSpec defines the desired state of ClusterAccessReview.
type ClusterGroupAccessSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// +required
	Cluster string `json:"cluster,omitempty"`
	// +required
	Users []ClusterUserDefinition `json:"users,omitempty"`
	// +required
	Requests []GroupAccessRequest `json:"requests,omitempty"`
}

type ClusterUserDefinition struct {
	// +required
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

type GroupAccessRequest struct {
	// +required
	ForUser   ClusterUserDefinition `json:"for_user,omitempty"`
	Approvers []string              `json:"approvers,omitempty"`
}

// ClusterGroupAccessStatus defines the observed state of ClusterAccessReview.
type ClusterGroupAccessStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:resource:scope=Cluster
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:selectablefield:JSONPath=`.spec.cluster`

// ClusterGroupAccess is the Schema for the clusteraccessreviews API.
type ClusterGroupAccess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ClusterGroupAccessSpec   `json:"spec"`
	Status ClusterGroupAccessStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterGroupAccessList contains a list of ClusterGroup.
type ClusterGroupAccessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterGroupAccess `json:"items"`
}

type ClusterGroupRequest struct{}

func init() {
	SchemeBuilder.Register(&ClusterGroupAccess{}, &ClusterGroupAccessList{})
}
