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

package v1alpha1

import ()

// publish provides I/F for publishing output
type publisher interface {
	Setup() error
	Publish(name string, manifest []byte) error
}

// GetPublisher returns Publisher along w/ its Spec
func (o *Output) GetPublisher() publisher {
	switch {
	case o.Spec.GitHub != nil:
		return o.Spec.GitHub
	case o.Spec.Slack != nil:
		return o.Spec.Slack
	}
	return nil
}
