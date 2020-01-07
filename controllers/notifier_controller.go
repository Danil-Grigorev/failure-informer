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
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	emailv1 "std/api/v1"

	ctx "context"

	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
// +kubebuilder:rbac:groups="",resources=secret,verbs=get;watch
// +kubebuilder:rbac:groups="",resources=secret/status,verbs=get;watch

func (r *NotifierReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("notifier", req.NamespacedName)

	notifier := &emailv1.Notifier{}
	err := r.Get(ctx.TODO(), req.NamespacedName, notifier)
	if err != nil {
		log.Error(err, "Can't get notifier")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secret, err := r.getEmailSecret(notifier)
	if err != nil {
		log.Error(err, "Failed to get email Secret")
		return ctrl.Result{}, nil
	}

	if secret == nil {
		err = r.createEmailSecret(notifier)
		if err != nil {
			log.Error(err, "Failed to create initial email Secret")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true}, nil
	}

	events, err := r.getFilteredEvents(notifier)
	if err != nil {
		log.Error(err, "Failed to list Pod related Events")
		return ctrl.Result{}, nil
	}

	err = r.notify(events)
	if err != nil {
		log.Error(err, "Failed to notify event")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *NotifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&emailv1.Notifier{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Event{}).
		Complete(r)
}

func (r *NotifierReconciler) getEmailSecret(notifier *emailv1.Notifier) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	namespacedName := types.NamespacedName{
		Namespace: notifier.GetNamespace(),
		Name:      notifier.GetName(),
	}
	err := r.Get(ctx.TODO(), namespacedName, &secret)

	if k8serror.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &secret, nil
}

func (r *NotifierReconciler) createEmailSecret(notifier *emailv1.Notifier) error {
	secretMeta := metav1.ObjectMeta{
		Namespace:   notifier.GetNamespace(),
		Name:        notifier.GetName(),
		Annotations: notifier.GetAnnotations(),
	}
	secret := &corev1.Secret{
		ObjectMeta: secretMeta,
	}
	err := ctrl.SetControllerReference(notifier, secret, r.Scheme)
	if err != nil {
		return err
	}

	err = r.Create(ctx.TODO(), secret)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (r *NotifierReconciler) getFilteredEvents(notifier *emailv1.Notifier) ([]corev1.Event, error) {
	capturedEvents := &corev1.EventList{}
	err := r.List(
		ctx.TODO(),
		capturedEvents,
		client.InNamespace(notifier.GetNamespace()),
		client.MatchingLabels(map[string]string{NotifyLabel: "true"}))
	if err != nil {
		return nil, err
	}

	return capturedEvents.Items, nil
}

func (r *NotifierReconciler) notify(events []corev1.Event) error {
	for _, event := range events {
		r.Log.Info(fmt.Sprintf("Reason: %v, Message: %#v", event.Reason, event.Message))
		eventCopy := event.DeepCopy()
		eventCopy.SetLabels(nil)
		err := r.Update(ctx.TODO(), eventCopy)
		if err != nil {
			return err
		}
	}

	return nil
}
