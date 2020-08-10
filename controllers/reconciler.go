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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/yaml"

	capturerv1alpha1 "github.com/terakoya76/manifest-capturer/apis/capturer/v1alpha1"
)

var Predicates func(bool) predicate.Predicate = func(haveGeneration bool) predicate.Predicate {
	if haveGeneration {
		return predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		}
	}

	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

func capture(ctx context.Context, r client.Client, resourceKind string, obj metav1.Object) (bool, error) {
	retry := false

	c, err := findCapturer(ctx, r, resourceKind, obj)
	if err != nil {
		retry = true
		return retry, err
	}
	if c == nil {
		return retry, nil
	}

	manifest, err := extractManifest(resourceKind, obj)
	if err != nil {
		retry = true
		return retry, err
	}

	if err = publish(ctx, r, c, manifest); err != nil {
		retry = true
		return retry, err
	}

	return retry, nil
}

func findCapturer(ctx context.Context, r client.Client, resourceKind string, obj metav1.Object) (*capturerv1alpha1.Capturer, error) {
	caps := capturerv1alpha1.CapturerList{}
	if err := r.List(
		ctx,
		&caps,
	); err != nil {
		return nil, err
	}

	for _, c := range caps.Items {
		if _, err := json.Marshal(c); err != nil {
			return nil, err
		}

		if c.Spec.NamespacedResource {
			if c.Spec.ResourceKind == resourceKind &&
				c.Spec.ResourceNamespace == obj.GetNamespace() &&
				c.Spec.ResourceName == obj.GetName() {
				return &c, nil
			}
		} else {
			if c.Spec.ResourceKind == resourceKind &&
				c.Spec.ResourceName == obj.GetName() {
				return &c, nil
			}
		}
	}

	return nil, nil
}

func extractManifest(resourceKind string, resource interface{}) ([]byte, error) {
	switch resourceKind {
	case "ClusterRole":
		if v, ok := resource.(*rbacv1.ClusterRole); ok {
			vc := v.DeepCopy()
			vc.ObjectMeta.Reset()
			vc.SetName(v.GetName())
			vc.SetNamespace(v.GetNamespace())
			vc.SetLabels(v.GetLabels())

			manifest, err := yaml.Marshal(vc)
			if err != nil {
				return []byte{}, err
			}
			return manifest, nil
		}

	case "ClusterRoleBinding":
		if v, ok := resource.(*rbacv1.ClusterRoleBinding); ok {
			vc := v.DeepCopy()
			vc.ObjectMeta.Reset()
			vc.SetName(v.GetName())
			vc.SetNamespace(v.GetNamespace())
			vc.SetLabels(v.GetLabels())

			manifest, err := yaml.Marshal(vc)
			if err != nil {
				return []byte{}, err
			}
			return manifest, nil
		}

	case "ConfigMap":
		if v, ok := resource.(*corev1.ConfigMap); ok {
			vc := v.DeepCopy()
			vc.ObjectMeta.Reset()
			vc.SetName(v.GetName())
			vc.SetNamespace(v.GetNamespace())
			vc.SetLabels(v.GetLabels())

			manifest, err := yaml.Marshal(vc)
			if err != nil {
				return []byte{}, err
			}
			return manifest, nil
		}

	case "Deployment":
		if v, ok := resource.(*appsv1.Deployment); ok {
			vc := v.DeepCopy()
			vc.ObjectMeta.Reset()
			vc.Status.Reset()
			vc.SetName(v.GetName())
			vc.SetNamespace(v.GetNamespace())
			vc.SetLabels(v.GetLabels())

			manifest, err := yaml.Marshal(vc)
			if err != nil {
				return []byte{}, err
			}
			return manifest, nil
		}

	case "Secret":
		if v, ok := resource.(*corev1.Secret); ok {
			vc := v.DeepCopy()
			vc.ObjectMeta.Reset()
			vc.SetName(v.GetName())
			vc.SetNamespace(v.GetNamespace())
			vc.SetLabels(v.GetLabels())

			manifest, err := yaml.Marshal(vc)
			if err != nil {
				return []byte{}, err
			}
			return manifest, nil
		}

	case "Service":
		if v, ok := resource.(*corev1.Service); ok {
			vc := v.DeepCopy()
			vc.ObjectMeta.Reset()
			vc.SetName(v.GetName())
			vc.SetNamespace(v.GetNamespace())
			vc.SetLabels(v.GetLabels())

			manifest, err := yaml.Marshal(vc)
			if err != nil {
				return []byte{}, err
			}
			return manifest, nil
		}

	case "ServiceAccount":
		if v, ok := resource.(*corev1.ServiceAccount); ok {
			vc := v.DeepCopy()
			vc.ObjectMeta.Reset()
			vc.SetName(v.GetName())
			vc.SetNamespace(v.GetNamespace())
			vc.SetLabels(v.GetLabels())

			manifest, err := yaml.Marshal(vc)
			if err != nil {
				return []byte{}, err
			}
			return manifest, nil
		}

	default:
		return nil, fmt.Errorf("unsupported resource type %s", resourceKind)
	}

	return nil, fmt.Errorf("failed to cast resource %v into resourceKind %s", resource, resourceKind)
}

func publish(ctx context.Context, r client.Client, c *capturerv1alpha1.Capturer, m []byte) error {
	outputs := make(map[string]capturerv1alpha1.Output)
	for _, outputName := range c.Spec.Outputs {
		var output capturerv1alpha1.Output
		if err := r.Get(
			ctx,
			types.NamespacedName{
				Namespace: c.GetNamespace(),
				Name:      outputName,
			},
			&output,
		); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}

			return err
		}

		outputs[outputName] = output
	}

	for outputName, output := range outputs {
		if err := output.GetPublisher().Publish(outputName, m); err != nil {
			return err
		}
	}

	return nil
}
