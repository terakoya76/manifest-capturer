/*


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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CapturerSpec defines the desired state of Capturer
type CapturerSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=bool

	NamespacedResource bool `json:"namespacedResource"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string

	ResourceKind string `json:"resourceKind"`

	// *kubebuilder:validation:Optional
	// +kubebuilder:validation:Format:=string

	ResourceNamespace string `json:"resourceNamespace,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string

	ResourceName string `json:"resourceName"`

	// +kubebuilder:validation:Required

	Outputs []string `json:"outputs"`
}

// CapturerStatus defines the observed state of Capturer
type CapturerStatus struct {
	// A list of pointers to currently running capturing object.
	// +optional
	Capturing []corev1.ObjectReference `json:"capturing,omitempty"`
}

// +kubebuilder:object:root=true

// Capturer is the Schema for the capturers API
type Capturer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CapturerSpec   `json:"spec,omitempty"`
	Status CapturerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CapturerList contains a list of Capturer
type CapturerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Capturer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Capturer{}, &CapturerList{})
}
