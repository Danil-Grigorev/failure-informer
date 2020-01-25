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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppScalerSpec defines the desired state of AppScaler
type AppScalerSpec struct {
	Replicas *int32   `json:"replicas"`
	Image    string   `json:"image"`
	Command  []string `json:"command"`
}

// AppScalerStatus defines the observed state of AppScaler
type AppScalerStatus struct {
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true

// AppScaler is the Schema for the appscalers API
type AppScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppScalerSpec   `json:"spec,omitempty"`
	Status AppScalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AppScalerList contains a list of AppScaler
type AppScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppScaler{}, &AppScalerList{})
}

func (r *AppScaler) ComposeReplicaSet() *v1beta1.ReplicaSet {
	objectMeta := metav1.ObjectMeta{
		Name:      r.GetName(),
		Namespace: r.GetNamespace(),
		Labels:    r.ComposeLabels(),
	}
	podContainers := []corev1.Container{corev1.Container{
		Name:    r.GetName(),
		Image:   r.Spec.Image,
		Command: r.Spec.Command,
	},
	}
	podSpec := corev1.PodSpec{
		Containers: podContainers,
	}
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: objectMeta,
		Spec:       podSpec,
	}
	replicaSetSpec := v1beta1.ReplicaSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: r.ComposeLabels(),
		},
		Replicas: r.Spec.Replicas,
		Template: podTemplate,
	}
	return &v1beta1.ReplicaSet{
		ObjectMeta: objectMeta,
		Spec:       replicaSetSpec,
	}
}

func (r *AppScaler) ComposeLabels() map[string]string {
	return map[string]string{"example": "true"}
}
