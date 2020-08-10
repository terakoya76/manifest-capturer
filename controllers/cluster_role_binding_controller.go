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
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterRoleBindingController reconciles a Capturer object
type ClusterRoleBindingController struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=capturer.stable.example.com,resources=capturers,verbs=get;list
// +kubebuilder:rbac:groups=capturer.stable.example.com,resources=capturers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capturer.stable.example.com,resources=outputs,verbs=get;list
// +kubebuilder:rbac:groups=capturer.stable.example.com,resources=outputs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac,resources=clusterrolebindings,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac,resources=clusterrolebindings/status,verbs=get

func (r *ClusterRoleBindingController) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("clusterrolebindings", req.NamespacedName)

	var crb rbacv1.ClusterRoleBinding
	if err := r.Get(ctx, req.NamespacedName, &crb); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	resourceKind := "ClusterRoleBinding"
	retry, err := capture(ctx, r, resourceKind, &crb)
	if err != nil {
		log.Error(err, "failed to capture")

		if retry {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil

}

func (r *ClusterRoleBindingController) SetupWithManager(mgr ctrl.Manager) error {
	haveGeneration := false
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&rbacv1.ClusterRoleBinding{},
			builder.WithPredicates(Predicates(haveGeneration)),
		).
		Complete(r)
}
