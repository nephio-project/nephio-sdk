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
	"fmt"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

var jsonStringConverterObj = JsonStringConverter{}

func TestIntialise(t *testing.T) {
	ll, _ := logrus.ParseLevel("fatal")
	logrus.SetLevel(ll)

	moduleStructMappingFile := "../config/struct_module_mapping.json"
	jsonStringConverterObj.setModStructMapping(moduleStructMappingFile)
	if len(jsonStringConverterObj.globalStructMapping) == 0 {
		t.Errorf("Intialise Failed| Unable to Populate Global-Struct-Mapping From Config-File %s", moduleStructMappingFile)
	}

	enumModuleMapping := "../config/enum_module_mapping.json"
	jsonStringConverterObj.setEnums(enumModuleMapping)
	if jsonStringConverterObj.globalEnumsSet.Size() == 0 {
		t.Errorf("Intialise Failed| Unable to Populate Enum-Module-Mapping From Config-File %s", enumModuleMapping)
	}
}

func TestCheckOpeningAndClosingBraces(t *testing.T) {
	tests := []Tests{
		{"{{}}{}", true},
		{"{{}}}", false},
	}

	for _, test := range tests {
		jsonStringConverterObj.generatedGoCode = test.input.(string)
		result := jsonStringConverterObj.checkOpeningAndClosingBraces()
		if result != test.expected.(bool) {
			t.Errorf("checkOpeningAndClosingBraces Failed| Expected %t | Got %t | Input %s", test.expected.(bool), result, test.input.(string))
		}
	}
}

/*
Tests For Base-Cases (int, bool, string, *int, *string, *bool)
*/
func TestFormatTypeValBaseCases(t *testing.T) {
	tests := []Tests{
		{[]string{"int32", "\"34\""}, "34"},
		{[]string{"string", "\"34\""}, "\"34\""},
		{[]string{"bool", "\"false\""}, "\"false\""},
		{[]string{"*int32", "\"34\""}, "int32Ptr(34)"},
		{[]string{"*string", "\"34\""}, "stringPtr(\"34\")"},
		{[]string{"*bool", "\"false\""}, "boolPtr(\"false\")"},
	}
	for _, test := range tests {
		testTyp, testVal := test.input.([]string)[0], test.input.([]string)[1]
		expected := test.expected.(string)
		result := jsonStringConverterObj.formatTypeVal(testTyp, testVal, 0)
		if result != expected {
			t.Errorf("FormatTypeVal Failed| Input : Type %s \t Value %s \nExpected %s \t Got %s\n", testTyp, testVal, expected, result)
		}

	}
}

/*
Tests For Composite-Literals (struct, enums, *struct)
*/
func TestFormatTypeValCompositeCases(t *testing.T) {
	tests := []Tests{
		{
			input: []string{"v1.ObjectMeta", "\tName : \"ABC\""},
			expected: `metav1.ObjectMeta{
	Name : "ABC"
}`},
		{
			input: []string{"*v1.ObjectMeta", "\tName : \"ABC\""},
			expected: `&metav1.ObjectMeta{
	Name : "ABC"
}`},
		{
			input:    []string{"v1.IncludeObjectPolicy", "\"True\""},
			expected: "metav1.IncludeObjectPolicy(\"True\")",
		},
	}

	for _, test := range tests {
		testTyp, testVal := test.input.([]string)[0], test.input.([]string)[1]
		expected := test.expected.(string)
		result := jsonStringConverterObj.formatTypeVal(testTyp, testVal, 0)
		if result != expected {
			// compare2Strings(result, expected)
			t.Errorf("FormatTypeVal Failed (Composite-Literal)| Input : Type %s \t Value %s \nExpected %s \t Got %s\n", testTyp, testVal, expected, result)
		}

	}
}

/*
Tests For Special-Type ([]byte, intstr.Type)
*/
func TestFormatTypeValSpecialCases(t *testing.T) {
	tests := []Tests{
		{[]string{"intstr.Type", "\"56\""}, "intstr.Type(56)"},
		{[]string{"[]byte", "\"my-secret\""}, "getDataForSecret(\"my-secret\")"},
	}

	for _, test := range tests {
		testTyp, testVal := test.input.([]string)[0], test.input.([]string)[1]
		expected := test.expected.(string)
		result := jsonStringConverterObj.formatTypeVal(testTyp, testVal, 0)
		if result != expected {
			// compare2Strings(result, expected)
			t.Errorf("FormatTypeVal Failed (Special Cases)| Input : Type %s \t Value %s \nExpected %s \t Got %s\n", testTyp, testVal, expected, result)
		}

	}
}

/*
Tests TraverseJson for Composite Cases (struct, slice, map)
*/
func TestTraverseJsonCompositeCases(t *testing.T) {
	tests := []Tests{
		{
			input: []string{"abc", "def"},
			expected: `
"abc",
"def",
`,
		},
		{
			input: map[string]interface{}{
				"Condition": map[string]interface{}{
					"type": "bool",
					"val":  true,
				},
			},
			expected: "Condition  : true, ",
		},
		{
			input: map[string]interface{}{
				"Labels": map[string]interface{}{
					"type": "map[string]string",
					"val": map[string]interface{}{
						"label1": "app1",
					},
				},
			},
			expected: "Labels : map[string]string{\n\t\"label1\" : \"app1\",\n},",
		},
	}

	for _, test := range tests {
		expected := test.expected.(string)
		result := jsonStringConverterObj.traverseJson(reflect.ValueOf(test.input), fmt.Sprint(reflect.TypeOf(test.input)), 0)
		if result != expected {
			// compare2Strings(result, expected)
			t.Errorf("TraverseJson Failed (Composite-Literal)| Input : Type %v \nExpected %s \t Got %s\n", test.input, expected, result)
		}

	}
}

func TestJsonToGoCode(t *testing.T) {
	jsonStringConverterObj.jsonToGoCode("tests/expected-json/deployment.json")
	if jsonStringConverterObj.generatedGoCode == "" {
		t.Errorf("JsonToGoCode Failed| Unable to Convert JSON To Go-Code")
	}
}
