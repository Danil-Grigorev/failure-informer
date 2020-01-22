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
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	emailv1 "std/api/v1"
)

// NotifierReconciler reconciles a Notifier object
type NotifierReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=email.notify.io,resources=notifiers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=email.notify.io,resources=notifiers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=event,verbs=get;list;watch

func (r *NotifierReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("notifier", req.NamespacedName)

	notifier := &emailv1.Notifier{}
	err := r.Get(ctx.TODO(), req.NamespacedName, notifier)
	if err != nil {
		log.Error(err, "Can't get notifier")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	events, err := r.getFilteredEvents(notifier)
	if err != nil {
		log.Error(err, "Failed to list Pod related Events")
		return ctrl.Result{Requeue: true}, nil
	}

	err = r.notify(notifier, events)
	if k8serror.IsConflict(err) {
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to notify event")
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *NotifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&emailv1.Notifier{}).
		Owns(&corev1.Event{}).
		Complete(r)
}

// Lists all events, which match the filter
func (r *NotifierReconciler) getFilteredEvents(notify *emailv1.Notifier) ([]corev1.Event, error) {
	labelFilter := map[string]string{}
	labelFilter[notify.GetNotifyLabel()] = "true"

	capturedEvents := &corev1.EventList{}

	err := r.List(
		ctx.TODO(),
		capturedEvents,
		client.InNamespace(notify.GetNamespace()),
		client.MatchingLabels(labelFilter))
	if err != nil {
		return nil, err
	}

	return capturedEvents.Items, nil
}

func (r *NotifierReconciler) notify(notifier *emailv1.Notifier, events []corev1.Event) error {
	for _, event := range events {
		r.Log.Info(fmt.Sprintf(`
		Event occured! Email sent: %v
		Reason: %v,
		Message: %#v,
		Pod: %v`,
			notifier.GetEmail(),
			event.Reason,
			event.Message,
			event.InvolvedObject.Name))

		eventCopy := event.DeepCopy()
		eventCopy.SetLabels(nil)
		err := r.Update(ctx.TODO(), eventCopy)
		if err != nil {
			return err
		}
	}

	return nil
}
