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
