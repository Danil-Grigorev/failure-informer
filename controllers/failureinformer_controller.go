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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	notifierv1 "std/api/v1"
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

func (r *FailureInformerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	log := r.Log.WithValues("failureinformer", req.NamespacedName)

	log.Info("Entered reconcile with " + req.String())

	return ctrl.Result{}, nil
}

func (r *FailureInformerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&notifierv1.FailureInformer{}).
		For(&corev1.Pod{}).
		Complete(r)
}
