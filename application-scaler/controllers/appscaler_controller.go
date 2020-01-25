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
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/api/extensions/v1beta1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	samplev1beta1 "std/api/v1beta1"
)

// AppScalerReconciler reconciles a AppScaler object
type AppScalerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=sample.example.com,resources=appscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sample.example.com,resources=appscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=extensions,resources=replicasets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=extensions,resources=replicasets/status,verbs=get;update;patch

func (r *AppScalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error
	log := r.Log.WithValues("appscaler", req.NamespacedName)

	appScaler := &samplev1beta1.AppScaler{}
	err = r.Get(context.TODO(), req.NamespacedName, appScaler)
	if err != nil {
		log.Error(err, "Can't get application scaler")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err = r.upadateReplicaSet(appScaler)
	if k8serror.IsConflict(err) {
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Can't update ReplicaSet")
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *AppScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samplev1beta1.AppScaler{}).
		Owns(&v1beta1.ReplicaSet{}).
		Complete(r)
}

func (r *AppScalerReconciler) upadateReplicaSet(appScaler *samplev1beta1.AppScaler) error {
	replicaSet := appScaler.ComposeReplicaSet()

	err := ctrl.SetControllerReference(appScaler, replicaSet, r.Scheme)
	if err != nil {
		r.Log.Error(err, "Unable to set controller reference on replica set")
		return err
	}

	operation, err := ctrl.CreateOrUpdate(context.TODO(), r.Client, replicaSet, mutate(replicaSet, replicaSet.Spec))
	r.Log.Info(fmt.Sprintf("Performed '%s' on repicaSet", operation))

	return err
}

func mutate(rs *v1beta1.ReplicaSet, updatedSpec v1beta1.ReplicaSetSpec) controllerutil.MutateFn {
	return func() error {
		rs.Spec = updatedSpec
		return nil
	}
}
