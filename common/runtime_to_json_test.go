package common

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
)

var runtimeJsonConverterObj = RuntimeJsonConverter{}

/*
It Tests the Full-Flow From Runtime-Obj to Json
And Also Test For Combined-Cases (Struct-in-a-Struct, Slice-in-a-Struct, Struct-in-a-Slice) & so-on
*/
func TestConvert(t *testing.T) {
	inputFile := "tests/test-yamls/deployment.yaml"

	decoder := scheme.Codecs.UniversalDeserializer()
	data, err := getFileContents(inputFile)
	if err != nil {
		t.Errorf("Unable to Load File %s| Error %s", inputFile, err)
	}
	runtimeObject, gvk, err := decoder.Decode(data, nil, nil)
	if err != nil {
		t.Errorf("Unable to Decode the Yaml| %s", inputFile)
	}
	runtimeJsonConverterObj.Convert(runtimeObject, *gvk)

	resultFile := "temp/temp.json"
	expectedFile := "tests/expected-json/deployment.json"
	resultData, _ := getFileContents(resultFile)
	expectedData, _ := getFileContents(expectedFile)

	var result interface{}
	var expected interface{}
	json.Unmarshal([]byte(resultData), &result)
	json.Unmarshal([]byte(expectedData), &expected)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Result Doesn't Matches with Expected| Kindly Check |\n ResultFile %s \t| ExpectedFile %s\n", resultFile, expectedFile)
	}
	os.RemoveAll("temp")
}

/*
Tests For Base Cases in DFS Traversal (Float, Int, String, Bool)
*/
func TestRunDfsJsonOmitEmptyBaseCases(t *testing.T) {
	tests := []Tests{
		{"abc", "abc"},
		{"", nil},
		{5, "5"},
		{3.14, "3.14"},
		{0, nil},
		{true, true},
	}
	for _, test := range tests {
		result := runtimeJsonConverterObj.runDfsJsonOmitEmpty(test.input, 0)
		if result != test.expected {
			t.Errorf("Test DfsJson (Base Cases) Failed | EXpected %v | Got %v", test.expected, result)
		}
	}
}

/*
Tests For Complex Cases in DFS Traversal (Struct, Slices, Maps)
*/
func TestRunDfsJsonOmitEmptyComplexCases(t *testing.T) {
	tests := []Tests{
		{[]string{"abc", "def", ""}, []interface{}{"abc", "def", ""}},
		{[]byte("my-secret"), "bXktc2VjcmV0"}, //Base64 encoded version of my-secret// This is also a TODO task (to check if it is important or not)
		// {[]interface{}{0, "abc"}, []interface{}{"", "abc"}},// This is TODO Task
		{metav1.ObjectMeta{}, nil}, //Empty Struct Should Return Nil
		{
			input: map[string]interface{}{
				"key1": "abc",
				"key2": 6,
			},
			expected: map[string]interface{}{
				"key1": "abc",
				"key2": "6",
			},
		},
		{
			input: metav1.ObjectMeta{
				Name:       "tests",
				Generation: 2,
			},
			expected: map[string]interface{}{
				"Name": map[string]interface{}{
					"type": "string",
					"val":  "tests",
				},
				"Generation": map[string]interface{}{
					"type": "int64",
					"val":  "2",
				},
			},
		},
	}
	for _, test := range tests {
		result := runtimeJsonConverterObj.runDfsJsonOmitEmpty(test.input, 0)
		if !reflect.DeepEqual(test.expected, result) {
			t.Errorf("Test DfsJson (Complex Cases) Failed | Expected %v | Got %v", test.expected, result)
		}
	}
	_ = metav1.ConditionStatus("True")
}

/*
Tests For Special Cases in DFS Traversal (resource.Quantity, v1.Time)
*/
func TestRunDfsJsonOmitEmptySpecialCases(t *testing.T) {
	tests := []Tests{
		{resource.MustParse("0"), nil},
		{metav1.Time{}, nil},
		{
			input:    resource.MustParse("64Mi"),
			expected: map[string]string{"type": "int", "val": "resource.MustParse(\"64Mi\")"},
		},
		{
			input: metav1.Time{Time: time.Time.AddDate(time.Time{}, 2, 3, 0)},
			expected: map[string]interface{}{
				"Time": map[string]string{
					"type": "int",
					"val":  time.Time.AddDate(time.Time{}, 2, 3, 0).GoString(),
				},
			},
		},
	}
	for _, test := range tests {
		result := runtimeJsonConverterObj.runDfsJsonOmitEmpty(test.input, 0)
		if !reflect.DeepEqual(test.expected, result) {
			t.Errorf("Test DfsJson (Special Cases) Failed | Expected %v | Got %v", test.expected, result)
		}
	}
}
