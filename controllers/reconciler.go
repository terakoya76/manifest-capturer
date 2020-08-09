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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/yaml"

	capturerv1alpha1 "github.com/terakoya76/manifest-capturer/apis/capturer/v1alpha1"
)

var Predicates predicate.Predicate = predicate.Funcs{
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

func capture(ctx context.Context, r client.Client, resourceKind string, resource interface{}) (bool, error) {
	retry := false

	if obj, ok := resource.(metav1.Object); ok {
		c, err := findCapture(ctx, r, resourceKind, obj)
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
	}

	return retry, fmt.Errorf("failed to cast resourceKind %s resource %+v", resourceKind, resource)
}

func findCapture(ctx context.Context, r client.Client, resourceKind string, obj metav1.Object) (*capturerv1alpha1.Capturer, error) {
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

		if c.Spec.ResourceKind == resourceKind &&
			c.Spec.ResourceNamespace == obj.GetNamespace() &&
			c.Spec.ResourceName == obj.GetName() {
			return &c, nil
		}
	}

	return nil, nil
}

func extractManifest(resourceKind string, resource interface{}) ([]byte, error) {
	switch resourceKind {
	case "ConfigMap":
		if cm, ok := resource.(*corev1.ConfigMap); ok {
			cmc := cm.DeepCopy()
			cmc.ObjectMeta.Reset()
			cmc.SetName(cm.GetName())
			cmc.SetNamespace(cm.GetNamespace())
			cmc.SetLabels(cm.GetLabels())

			manifest, err := yaml.Marshal(cmc)
			if err != nil {
				return []byte{}, err
			}

			return manifest, nil
		}

	case "Deployment":
		if d, ok := resource.(*appsv1.Deployment); ok {
			dc := d.DeepCopy()
			dc.ObjectMeta.Reset()
			dc.Status.Reset()
			dc.SetName(d.GetName())
			dc.SetNamespace(d.GetNamespace())
			dc.SetLabels(d.GetLabels())

			manifest, err := yaml.Marshal(dc)
			if err != nil {
				return []byte{}, err
			}

			return manifest, nil
		}

	default:
		return nil, fmt.Errorf("unsupported resource type %s", resourceKind)
	}

	return nil, fmt.Errorf("failed to cast resourceKind %s resource %+v", resourceKind, resource)
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
