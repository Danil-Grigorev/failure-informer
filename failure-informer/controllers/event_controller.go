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

package controllers

import (
	ctx "context"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	emailv1 "std/api/v1"
)

const (
	InvolvedObjectKind = "involvedobject.kind"
	Reason             = "reason"
	Kind               = "kind"
)

var NotifyLabel = "%s-notify"

// EventReconciler reconciles a Event object
type EventReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events/status,verbs=get;update;patch

func (r *EventReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("event", req.NamespacedName)

	event := &corev1.Event{}
	err := r.Get(ctx.TODO(), req.NamespacedName, event)
	if k8serror.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Could not get Event: "+req.String())
		return ctrl.Result{Requeue: true}, nil
	}

	// Skip purely informational events
	if event.Type != "Warning" {
		return ctrl.Result{}, nil
	}

	notifiers, err := r.getMatchingNotifiers(event)
	if err != nil {
		log.Error(err, "Can't match notifiers for event")
		return ctrl.Result{Requeue: true}, nil
	}

	// No Notifier CRs found in the namespace
	if len(notifiers) == 0 {
		return ctrl.Result{}, nil
	}

	for _, notifier := range notifiers {
		err = r.requestNotify(event, &notifier)
		if k8serror.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Error on updating Event with notify label")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *EventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Event{}).
		WithEventFilter(EventPredicate{}).
		Complete(r)
}

func (r *EventReconciler) getMatchingNotifiers(event *corev1.Event) ([]emailv1.Notifier, error) {
	matchedNotifiers := []emailv1.Notifier{}
	notifierList := &emailv1.NotifierList{}
	err := r.Client.List(ctx.TODO(), notifierList, client.InNamespace(event.GetNamespace()))
	if err != nil {
		return matchedNotifiers, err
	}

	for _, notifier := range notifierList.Items {
		matched, err := regexp.MatchString(notifier.Spec.Filter, event.Reason)
		if err != nil {
			return matchedNotifiers, err
		}
		if matched {
			matchedNotifiers = append(matchedNotifiers, notifier)
		}
	}

	return matchedNotifiers, nil
}

func (r *EventReconciler) requestNotify(event *corev1.Event, notify *emailv1.Notifier) error {
	eventCopy := event.DeepCopy()

	setNotifyLabel(eventCopy, notify)

	err := ctrl.SetControllerReference(notify, eventCopy, r.Scheme)
	if err != nil {
		return errors.Wrap(err, "Failed to set Event referense to Notifier")
	}

	err = r.Update(ctx.TODO(), eventCopy)
	if err != nil {
		return errors.Wrap(err, "Error on updating Event with notify label")
	}

	return nil
}

func setNotifyLabel(e *corev1.Event, notify *emailv1.Notifier) {
	updatedLabels := make(map[string]string)
	for label, value := range e.GetLabels() {
		updatedLabels[label] = value
	}
	updatedLabels[notify.GetNotifyLabel()] = "true"

	e.SetLabels(updatedLabels)
}
