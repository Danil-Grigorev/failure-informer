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
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NotifyPrefix Label prefix for current Notify resource
const NotifyPrefix = "%s-notify"

// NotifierSpec defines the desired state of Notifier
type NotifierSpec struct {
	Email   string   `json:"email"`
	Filters []string `json:"filters"`
}

// NotifierStatus defines the observed state of Notifier
type NotifierStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Notifier is the Schema for the notifiers API
type Notifier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotifierSpec   `json:"spec,omitempty"`
	Status NotifierStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NotifierList contains a list of Notifier
type NotifierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Notifier `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Notifier{}, &NotifierList{})
}

func (r Notifier) GetEmail() string {
	return r.Spec.Email
}

func (r Notifier) GetFilters() []string {
	return r.Spec.Filters
}

func (r Notifier) GetNotifyLabel() string {
	return fmt.Sprintf(NotifyPrefix, r.GetName())
}

func (r Notifier) FilterMatch(input string) (bool, error) {
	for _, filter := range r.GetFilters() {
		matched, err := regexp.MatchString(filter, input)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}

	return true, nil
}

func (r NotifierList) Matching(input string) ([]Notifier, error) {
	matchedNotifiers := []Notifier{}
	for _, notifier := range r.Items {
		match, err := notifier.FilterMatch(input)
		if err != nil {
			return nil, err
		}
		if match {
			matchedNotifiers = append(matchedNotifiers, notifier)
		}
	}
	return matchedNotifiers, nil
}
