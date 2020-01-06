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
	"context"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	notifierv1 "std/api/v1"

	corev1 "k8s.io/api/core/v1"
	betav1extension "k8s.io/api/extensions/v1beta1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

// FailureInformerReconciler reconciles a FailureInformer object
type FailureInformerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=notifier.email.informer.io,resources=failureinformers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=notifier.email.informer.io,resources=failureinformers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pod,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pod/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secret,verbs=get;watch
// +kubebuilder:rbac:groups="",resources=secret/status,verbs=get;watch
// +kubebuilder:rbac:groups="extensions",resources=replicaset,verbs=get;list;watch
// +kubebuilder:rbac:groups="extensions",resources=replicaset/status,verbs=get;list;watch

func (r *FailureInformerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("failureinformer", req.NamespacedName)

	log.Info("Entered reconcile with " + req.String())

	failureInformer := &notifierv1.FailureInformer{}
	err := r.Get(context.TODO(), req.NamespacedName, failureInformer)
	if err != nil {
		log.Error(err, "Can't get failureInformer")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secret, err := r.getEmailSecret(failureInformer)
	if err != nil {
		log.Error(err, "Failed to get email secret")
		return ctrl.Result{}, nil
	}

	if secret == nil {
		err = r.createEmailSecret(failureInformer)
		if err != nil {
			log.Error(err, "Failed to create initial email secret")
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *FailureInformerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&notifierv1.FailureInformer{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Secret{}).
		Owns(&betav1extension.ReplicaSet{}).
		Complete(r)
}

func (r *FailureInformerReconciler) getEmailSecret(notifier *notifierv1.FailureInformer) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	namespacedName := types.NamespacedName{
		Namespace: notifier.ObjectMeta.Namespace,
		Name:      notifier.ObjectMeta.Name,
	}
	err := r.Get(context.TODO(), namespacedName, &secret)

	if k8serror.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &secret, nil
}

func (r *FailureInformerReconciler) createEmailSecret(notifier *notifierv1.FailureInformer) error {
	secretMeta := metav1.ObjectMeta{
		Namespace:   notifier.ObjectMeta.Namespace,
		Name:        notifier.ObjectMeta.Name,
		Annotations: notifier.GetAnnotations(),
	}
	secret := &corev1.Secret{
		ObjectMeta: secretMeta,
	}
	err := ctrl.SetControllerReference(notifier, secret, r.Scheme)
	if err != nil {
		return err
	}

	err = r.Create(context.TODO(), secret)
	if err != nil && !k8serror.IsAlreadyExists(err) {
		return err
	}

	return nil
}
