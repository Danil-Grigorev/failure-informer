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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FailureInformerSpec defines the desired state of FailureInformer
type FailureInformerSpec struct {
	Replicas *int32 `json:"replicas"`
	Image    string `json:"image"`
	Email    string `json:"email"`
}

// FailureInformerStatus defines the observed state of FailureInformer
type FailureInformerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// FailureInformer is the Schema for the failureinformers API
type FailureInformer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FailureInformerSpec   `json:"spec,omitempty"`
	Status FailureInformerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FailureInformerList contains a list of FailureInformer
type FailureInformerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FailureInformer `json:"items"`
}

var FailureInformerAnnotations = map[string]string{
	"app": "notifier",
}

func init() {
	SchemeBuilder.Register(&FailureInformer{}, &FailureInformerList{})
}

func (r FailureInformer) CreatePodsTemplate() corev1.PodTemplateSpec {
	podTemplateMeta := metav1.ObjectMeta{
		GenerateName: r.GetName() + "-",
		Namespace:    r.GetNamespace(),
		Labels:       r.GetLabels(),
		Annotations:  r.GetAnnotations(),
	}
	podContainers := corev1.Container{
		Name:    r.GetName(),
		Image:   r.GetImage(),
		Command: []string{"sleep", "1"},
	}
	podTemplateSpec := corev1.PodSpec{
		Containers: []corev1.Container{podContainers},
	}
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: podTemplateMeta,
		Spec:       podTemplateSpec,
	}

	return podTemplate
}

func (r FailureInformer) GetAnnotations() map[string]string {
	return FailureInformerAnnotations
}

func (r FailureInformer) GetLabels() map[string]string {
	return r.GetAnnotations()
}

func (r FailureInformer) GetImage() string {
	return r.Spec.Image
}

func (r FailureInformer) GetReplicas() *int32 {
	return r.Spec.Replicas
}

func (r FailureInformer) GetEmail() string {
	return r.Spec.Email
}
