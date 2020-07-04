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
	"encoding/json"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/yaml"

	capturerv1alpha1 "github.com/terakoya76/manifest-capturer/apis/capturer/v1alpha1"
)

// ConfigMapReconciler reconciles a Capturer object
type ConfigMapReconciler struct {
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

func (r *ConfigMapReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("configmap", req.NamespacedName)

	var cm corev1.ConfigMap
	if err := r.Get(ctx, req.NamespacedName, &cm); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	c, err := r.findCapture(&cm)
	if err != nil {
		log.Error(err, "failed to find Capture")
		return ctrl.Result{}, err
	}
	if c == nil {
		log.Info("unable to find Capture")
		return ctrl.Result{}, nil
	}

	var output capturerv1alpha1.Output
	if err = r.Get(context.TODO(),
		types.NamespacedName{
			Namespace: req.Namespace,
			Name:      c.Spec.Output,
		},
		&output,
	); err != nil {
		if errors.IsNotFound(err) {
			log.Info("non exist output is specified")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to find Output")
		return ctrl.Result{}, err
	}

	manifest, err := r.getManifest(&cm)
	if err != nil {
		log.Error(err, "unable to fetch manifest")
	}

	log.Info(string(manifest))
	if err := output.GetPublisher().Publish(manifest); err != nil {
		log.Error(err, "unable to publish manifest")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ConfigMapReconciler) findCapture(cm *corev1.ConfigMap) (*capturerv1alpha1.Capturer, error) {
	caps := capturerv1alpha1.CapturerList{}
	if err := r.List(
		context.TODO(),
		&caps,
	); err != nil {
		return nil, err
	}

	for _, c := range caps.Items {
		if _, err := json.Marshal(c); err != nil {
			return nil, err
		}

		if c.Spec.ResourceKind == "ConfigMap" &&
			c.Spec.ResourceNamespace == cm.GetNamespace() &&
			c.Spec.ResourceName == cm.GetName() {
			return &c, nil
		}
	}

	return nil, nil
}

func (r *ConfigMapReconciler) getManifest(cm *corev1.ConfigMap) ([]byte, error) {
	cmc := cm.DeepCopy()
	cmc.SetManagedFields(nil)

	manifest, err := yaml.Marshal(cmc)
	if err != nil {
		return []byte{}, err
	}

	return manifest, nil
}

func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
			&corev1.ConfigMap{},
			builder.WithPredicates(p),
		).
		Complete(r)
}
