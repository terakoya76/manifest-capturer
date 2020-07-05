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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capturerv1alpha1 "github.com/terakoya76/manifest-capturer/apis/capturer/v1alpha1"
)

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
