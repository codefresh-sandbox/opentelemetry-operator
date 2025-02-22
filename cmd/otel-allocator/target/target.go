// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package target

import (
	"fmt"
	"net/url"

	"github.com/prometheus/common/model"
)

// This package contains common structs and methods that relate to scrape targets.
type LinkJSON struct {
	Link string `json:"_link"`
}

type Item struct {
	JobName       string
	Link          LinkJSON
	TargetURL     string
	Label         model.LabelSet
	CollectorName string
}

func (t Item) Hash() string {
	return t.JobName + t.TargetURL + t.Label.Fingerprint().String()
}

func NewItem(jobName string, targetURL string, label model.LabelSet, collectorName string) *Item {
	return &Item{
		JobName:       jobName,
		Link:          LinkJSON{fmt.Sprintf("/jobs/%s/targets", url.QueryEscape(jobName))},
		TargetURL:     targetURL,
		Label:         label,
		CollectorName: collectorName,
	}
}
