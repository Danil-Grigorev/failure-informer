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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	emailv1 "std/api/v1"

	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	InvolvedObjectKind = "involvedobject.kind"
	Reason             = "reason"
	Kind               = "kind"
)

var FailureReasons = []string{
	"Failed",
	"Evicted",
	"FailedMount",
	"BackOff",
}

var NotifyLabel = "notify"

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

	notifierList := &emailv1.NotifierList{}
	err := r.Client.List(ctx.TODO(), notifierList, client.InNamespace(req.Namespace))
	if err != nil {
		log.Error(err, "Failed to list Notifiers")
		return ctrl.Result{}, nil
	}

	// No Notifier CRs found in the namespace
	if len(notifierList.Items) == 0 {
		return ctrl.Result{}, nil
	}

	event := &corev1.Event{}
	err = r.Get(ctx.TODO(), req.NamespacedName, event)
	if k8serror.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Could not get Event: "+req.String())
		return ctrl.Result{}, nil
	}

	if event.Type != "Warning" {
		return ctrl.Result{}, nil
	}

	failureEvent := false
	for _, failureReason := range FailureReasons {
		if event.Reason == failureReason {
			failureEvent = true
			break
		}
	}

	if !failureEvent {
		return ctrl.Result{}, nil
	}

	eventCopy := event.DeepCopy()

	eventCopy.SetLabels(map[string]string{NotifyLabel: "true"})

	for _, notifier := range notifierList.Items {
		err := ctrl.SetControllerReference(&notifier, eventCopy, r.Scheme)
		if err != nil {
			log.Error(err, "Failed to set Event referense to Notifier")
			return ctrl.Result{}, nil
		}
	}

	err = r.Update(ctx.TODO(), eventCopy)
	if err != nil {
		log.Error(err, "Error on updating Event with notify label")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *EventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Event{}).
		WithEventFilter(EventPredicate{}).
		Complete(r)
}
