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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SlackOutput defines the spec for integrating with GitHub
type SlackOutput struct {
	// +kubebuilder:validation:Format:=string

	WebhookUrl string `json:"webhookUrl"`
}

func (o *SlackOutput) Setup() error {
	return nil
}

func (o *SlackOutput) Publish(name string, manifest []byte) (err error) {
	url := o.WebhookUrl

	content := fmt.Sprintf(
		"A capture is reported by manifest-capturer %s\n\n```%s```",
		name,
		string(manifest),
	)

	var jsonStr = []byte(fmt.Sprintf(`{"text":"%s"}`, escapeString(content)))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json.RawMessage(jsonStr)))
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func escapeString(str string) string {
	return strings.Replace(str, `"`, `\"`, -1)
}
