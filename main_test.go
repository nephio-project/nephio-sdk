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

package main

import (
	"fmt"
	"testing"
)

/*
Tests for Runtime-OBject Way of Handling KRM-Object
*/
func TestHandleSingleYamlDeployment(t *testing.T) {
	inputFilePath := "common/tests/test-yamls/deployment.yaml"
	runtimeObjList, gvkList, unstructObjList, unstructGvkList := handleSingleYaml(inputFilePath)
	fmt.Println(runtimeObjList, gvkList, unstructObjList, unstructGvkList)
	if len(runtimeObjList) == 0 {
		t.Errorf("Unable to convert yaml to RuntimeObject")
	}
	if gvkList[0].Kind != "Deployment" {
		t.Errorf("Kind Detected is not what expected | Detected %s | Expected Deployment", gvkList[0].Kind)
	}
}

/*
Tests for Unstructured Way of Handling KRM-Object
*/
func TestHandleSingleYamlCR(t *testing.T) {
	inputFilePath := "common/tests/test-yamls/third-party-cr.yaml"
	runtimeObjList, gvkList, unstructObjList, unstructGvkList := handleSingleYaml(inputFilePath)
	fmt.Println(runtimeObjList, gvkList, unstructObjList, unstructGvkList)
	if len(unstructObjList) == 0 {
		t.Errorf("Unable to convert yaml to RuntimeObject")
	}
	if unstructGvkList[0].Kind != "ThirdPartyCR" {
		t.Errorf("Kind Detected is not what expected | Detected %s | Expected ThirdPartyCR", unstructGvkList[0].Kind)
	}
}
