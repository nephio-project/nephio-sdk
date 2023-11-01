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
			input: map[string]interface{}{
				"Name": "ABC",
			},
			expected: "map[string]interface{}{\n\"Name\": \"ABC\",\n}",
		},
		{
			input: map[string]interface{}{
				"Replicas": 4,
			},
			expected: "map[string]interface{}{\n\"Replicas\": 4,\n}",
		},
		{
			input: map[string]interface{}{
				"Conditions": []interface{}{true},
			},
			expected: "map[string]interface{}{\n\"Conditions\": []interface{}{\n\ttrue,\n\t},\n}",
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
	data, err := getFileContents(inputFilePath)
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
