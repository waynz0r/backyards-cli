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

package questionnaire

import (
	"reflect"
	"strconv"
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
)

func GetQuestionsFromStruct(obj interface{}) ([]*survey.Question, error) {
	qs := make([]*survey.Question, 0)

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	for k := 0; k < t.NumField(); k++ {

		f := t.Field(k)
		elem := v.Field(k)
		desc := f.Tag.Get("survey.question")
		if desc != "" {
			var def string
			switch elem.Kind() {
			case reflect.Int32:
				def = strconv.Itoa(int(elem.Interface().(int32)))
			case reflect.String:
				def = elem.Interface().(string)
			default:
				return nil, errors.Errorf("unsupported field type: %s", elem.Type())
			}

			qs = append(qs, &survey.Question{
				Name:     f.Name,
				Prompt:   &survey.Input{Message: desc, Default: def},
				Validate: validateFunc(f.Tag.Get("survey.validate")),
			})
		}
	}

	return qs, nil
}

func validateFunc(t string) survey.Validator {
	switch t {
	case "int":
		return func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				i, err := strconv.Atoi(s)
				if err != nil {
					return err
				}
				if i <= 0 {
					return errors.New("value must be greater than 0")
				}
			} else {
				return errors.New("invalid input type")
			}
			return nil
		}
	case "durationstring":
		return func(ans interface{}) error {
			if s, ok := ans.(string); ok {
				_, err := time.ParseDuration(s)
				return err
			}
			return errors.New("invalid input type")
		}
	default:
		return func(ans interface{}) error {
			return nil
		}
	}
}
