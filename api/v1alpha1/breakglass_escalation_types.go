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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type (
	// BreakglassEscalationSpec defines the desired state of BreakglassEscalation.
	BreakglassEscalationSpec struct {
		// +required
		Cluster string `json:"cluster,omitempty"`
		// +required
		Username string `json:"username,omitempty"`
		// +required
		AllowedGroups []string `json:"allowedGroups,omitempty"`
		// +required
		EscalatedGroup string `json:"escalatedGroup,omitempty"`
		// +required
		Approvers BreakglassEscalationApprovers `json:"approvers,omitempty"`
	}

	// BreakglassEscalationApprovers
	BreakglassEscalationApprovers struct {
		Users  []string `json:"users,omitempty"`
		Groups []string `json:"groups,omitempty"`
	}

	// BreakglassEscalationStatus defines the observed state of BreakglassEscalation.
	BreakglassEscalationStatus struct{}
)

// +kubebuilder:resource:scope=Cluster
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:selectablefield:JSONPath=`.spec.cluster`
// +kubebuilder:selectablefield:JSONPath=`.spec.username`
// +kubebuilder:selectablefield:JSONPath=`.spec.escalatedGroup`

type BreakglassEscalation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   BreakglassEscalationSpec   `json:"spec"`
	Status BreakglassEscalationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BreakglassEscalationList contains a list of BreakglassEscalation.
type BreakglassEscalationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BreakglassEscalation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BreakglassEscalation{}, &BreakglassEscalationList{})
}
