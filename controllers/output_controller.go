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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	capturerv1alpha1 "github.com/terakoya76/manifest-capturer/apis/capturer/v1alpha1"
)

// OutputReconciler reconciles a Capturer object
type OutputReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=capturer.stable.example.com,resources=capturers,verbs=get;list
// +kubebuilder:rbac:groups=capturer.stable.example.com,resources=capturers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capturer.stable.example.com,resources=outputs,verbs=get;list
// +kubebuilder:rbac:groups=capturer.stable.example.com,resources=outputs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=,resources=configmaps/status,verbs=get

func (r *OutputReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("output", req.NamespacedName)

	var o capturerv1alpha1.Output
	if err := r.Get(ctx, req.NamespacedName, &o); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if err := o.GetPublisher().Setup(); err != nil {
		log.Error(err, "failed to setup Output")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *OutputReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var p predicate.Predicate = predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&capturerv1alpha1.Output{},
			builder.WithPredicates(p),
		).
		Complete(r)
}
