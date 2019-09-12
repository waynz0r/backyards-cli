// Copyright © 2019 Banzai Cloud
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

package ts

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"emperror.dev/errors"
)

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

type parsedSubsets map[string]int

func (p parsedSubsets) String() string {
	parts := make([]string, 0)
	for subset, weight := range p {
		parts = append(parts, fmt.Sprintf("%s=%d", subset, weight))
	}

	sort.Strings(parts)

	return strings.Join(parts, ", ")
}

func (p parsedSubsets) Validate() error {
	sum := 0

	for _, weight := range p {
		sum += weight
	}

	if sum != 100 {
		return errors.New("sum of subset weights must be 100")
	}

	return nil
}

func parseSubsets(subsets []string) (parsedSubsets, error) {
	parsedSubsets := make(parsedSubsets)

	for _, subset := range subsets {
		parts := strings.Split(subset, "=")
		if len(parts) != 2 || !dns1123LabelRegexp.MatchString(parts[1]) {
			return nil, errors.Errorf("invalid subset: '%s': format must be <subset>=<weight>", subset)
		}

		weight, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, errors.Errorf("invalid subset: '%s': format must be <subset>=<weight>", subset)
		}
		parsedSubsets[parts[0]] = weight
	}

	err := parsedSubsets.Validate()
	if err != nil {
		return nil, err
	}

	return parsedSubsets, nil
}
