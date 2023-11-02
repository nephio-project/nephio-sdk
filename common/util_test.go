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
	"os"
	"testing"
)

func TestRepeat(t *testing.T) {
	result := repeat("a", 3)
	if result != "aaa" {
		t.Errorf("Util-tests | 'Repeat' test failed | Expected %s | Got %s", "aaa", result)
	}
}

func TestCreateDirIfDontExist(t *testing.T) {
	err := createDirIfDontExist("tests/test_createdir")
	if err != nil {
		t.Errorf("Util-tests | 'CreateDirIfDontExist' test failed | %s", err)
	}

	// Check If directory is created or not
	folderContent, _ := os.ReadDir("tests")
	testPassed := false
	for _, files := range folderContent {
		if files.Type().IsDir() {
			if files.Name() == "test_createdir" {
				testPassed = true
				break
			}
		}
	}
	if !testPassed {
		t.Errorf("Util-tests | 'CreateDirIfDontExist' test failed | Directory Not Found where expected")
	}
	os.RemoveAll("tests/test_createdir")
}

func TestRecursiveListYamls(t *testing.T) {
	result := RecursiveListYamls("tests")
	// testHelmChartFolder := "tests/test-helmCharts/hello-world/"
	// expected := []string{"tests/test-yamls/deployment.yaml", "tests/test-yamls/third-party-cr.yaml",
	// 	testHelmChartFolder + "Chart.yaml", testHelmChartFolder + "values.yaml",
	// 	testHelmChartFolder + "templates/deployment.yaml", testHelmChartFolder + "templates/service.yaml",
	// 	testHelmChartFolder + "templates/serviceaccount.yaml", testHelmChartFolder + "templates/third-party-cr.yaml",
	// }

	// if !reflect.DeepEqual(result, expected) {
	// 	t.Errorf("Util-tests | 'RecursiveListYamls' test failed | \n Expected %v \n Got %v", expected, result)
	// }
	if len(result) != 8 {
		t.Errorf("Util-tests | 'RecursiveListYamls' test failed | \n Expected Length %v \n Got %v", 8, result)
	}

}

func TestHandleMultiLineStrings(t *testing.T) {
	input := "abc\nd"
	result := handleMultiLineStrings(input)
	expected := "\"abc\\n\" + \n \"d\"" // Expected "abc\n" + \n "d" where Second \n Represents actual new line
	if result != expected {
		t.Errorf("Util-tests | 'HandleMultiLineStrings' test failed | \n Expected %v \n Got %v", expected, result)
	}
}
