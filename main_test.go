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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/sirupsen/logrus"
)

func setLogLevelFatal() {
	ll, err := logrus.ParseLevel("fatal")
	if err != nil {
		ll = logrus.DebugLevel
	}
	logrus.SetLevel(ll)
}

func checkIfHelmInstalled() error {
	cmdStruct := exec.Command("helm", "version")
	stderr, _ := cmdStruct.StderrPipe() // Intialising a Pipe to read error stream
	if err := cmdStruct.Start(); err != nil {
		fmt.Println("Unable to Start Cmd Pipe to check Helm-Install")
		return err
	}

	scanner := bufio.NewScanner(stderr)
	helmCmdErr := ""
	for scanner.Scan() {
		helmCmdErr += scanner.Text()
	}
	if len(helmCmdErr) > 0 {
		fmt.Println("Error while checking the Helm Version|", helmCmdErr)
		return fmt.Errorf(helmCmdErr)
	}
	// fmt.Println("helm is installed")
	return nil
}

/*
Tests for Runtime-OBject Way of Handling KRM-Object
*/
func TestHandleSingleYamlDeployment(t *testing.T) {
	setLogLevelFatal()
	inputFilePath := "common/tests/test-yamls/deployment.yaml"
	runtimeObjList, gvkList, _, _ := handleSingleYaml(inputFilePath)
	// fmt.Println(runtimeObjList, gvkList, unstructObjList, unstructGvkList)
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
	setLogLevelFatal()
	inputFilePath := "common/tests/test-yamls/third-party-cr.yaml"
	_, _, unstructObjList, unstructGvkList := handleSingleYaml(inputFilePath)
	// fmt.Println(runtimeObjList, gvkList, unstructObjList, unstructGvkList)
	if len(unstructObjList) == 0 {
		t.Errorf("Unable to convert yaml to RuntimeObject")
	}
	if unstructGvkList[0].Kind != "ThirdPartyCR" {
		t.Errorf("Kind Detected is not what expected | Detected %s | Expected ThirdPartyCR", unstructGvkList[0].Kind)
	}
}

func TestMainFunc(t *testing.T) {
	setLogLevelFatal()
	err := checkIfHelmInstalled()
	if err != nil {
		fmt.Println("Helm Not Installed Detected| Aborting the current test")
		return
	}
	saveCmdArgs := os.Args
	os.Args = []string{"main.go", "common/tests/test-helmCharts/hello-world/", "abc"}
	os.Args = append(os.Args, saveCmdArgs...)
	main()
	os.Args = saveCmdArgs
	// According to the temp directory should have been deleted, If it is not Then The flow has encountered as error
	if _, err := os.Stat("temp/"); err == nil {
		_ = os.RemoveAll("temp")
		_ = os.Remove("outputs/generated_code.go")
		t.Errorf("Temp Directory still exists| Manually deleting | Failing this test")
	}

	// The Test is Considered PASSED if generated_code.go is created in the outputs/ folder
	if _, err := os.Stat("outputs/generated_code.go"); err != nil {
		t.Errorf("Generated_code.go File doesn't exist| Failing this test")
	}
	_ = os.Remove("outputs/generated_code.go")

}
