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
	"bufio"
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

type HelmYamlConvertor struct {
	Namespace string
	Chartpath string
}

/*
Converts the Helm-Chart to Yaml Template in temp folder,
Runs the bash command "helm template <chartpath> --namespace <namespace> --output-dir temp/templated/"
Todo: Increase the functionality to handle remote helm charts, and support for using different values.yaml & so on
*/
func (obj *HelmYamlConvertor) ConvertHelmToYaml() error {
	logrus.Info(obj.Namespace, " ", obj.Chartpath)
	logrus.Info(" ----------------- Converting Helm to Yaml --------------------------")
	_ = createDirIfDontExist("temp")
	// logrus.Info(err)
	if obj.Namespace == "" {
		obj.Namespace = "default"
	}
	cmdStruct := exec.Command("helm", "template", obj.Chartpath, "--namespace", obj.Namespace, "--output-dir", "temp/templated/")
	stderr, _ := cmdStruct.StderrPipe() // Intialising a Pipe to read error stream
	if err := cmdStruct.Start(); err != nil {
		logrus.Error(err)
		return err
	}

	scanner := bufio.NewScanner(stderr)
	helmCmdErr := ""
	for scanner.Scan() {
		helmCmdErr += scanner.Text()
	}
	if len(helmCmdErr) > 0 {
		logrus.Error("Error while running the command| helm template " + obj.Chartpath + " --namespace " + obj.Namespace + " --output-dir temp/templated/ ")
		return fmt.Errorf(helmCmdErr)
	}
	return nil
}
