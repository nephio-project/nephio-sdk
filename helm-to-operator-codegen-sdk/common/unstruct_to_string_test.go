/*
Copyright 2023 The Nephio Authors.

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

package common

import (
	"reflect"
	"testing"
)

var unstructStringConverterObj = UnstructStringConverter{}

/*
Composite Cases : Maps, Slices
*/
func TestRunDfsUnstructCompositeCases(t *testing.T) {
	tests := []Tests{
		{
			input: map[string]any{
				"Name": "ABC",
			},
			expected: "map[string]any{\n\"Name\": \"ABC\",\n}",
		},
		{
			input: map[string]any{
				"Replicas": 4,
			},
			expected: "map[string]any{\n\"Replicas\": 4,\n}",
		},
		{
			input: map[string]any{
				"Conditions": []any{true},
			},
			expected: "map[string]any{\n\"Conditions\": []any{\n\ttrue,\n\t},\n}",
		},
	}

	for _, test := range tests {
		result := unstructStringConverterObj.runDfsUnstruct(reflect.ValueOf(test.input), 0)
		expected := test.expected.(string)
		if expected != result {
			t.Errorf("RunDFSUnstructTest Failed| Input %v \nExpected %s \tGot %s", test.input, expected, result)
		}
	}
}

func TestConvertUnstruct(t *testing.T) {
	inputFilePath := "tests/test-yamls/deployment.yaml"
	data, err := GetFileContents(inputFilePath)
	if err != nil {
		t.Errorf("Unable to Open file %s for Unstruct-Convert| Error %v", inputFilePath, err)
	}
	unstructObj, _, err := unstructuredDecode(data)
	if err != nil {
		t.Errorf("Unable to convert KRM OBject to Unstruct During Unstruct-Convert| Error %v", err)
	}
	result := unstructStringConverterObj.Convert(*unstructObj)
	if result == "" {
		t.Errorf("Unable to generate goCode for Unstruct Object")
	}
}
