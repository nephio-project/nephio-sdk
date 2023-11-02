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

func TestConvertHelmToYaml(t *testing.T) {
	var helmYamlConvertor = HelmYamlConvertor{Namespace: "myns", Chartpath: "tests/test-helmCharts/hello-world/"}
	err := helmYamlConvertor.ConvertHelmToYaml()
	if err != nil {
		t.Errorf("Unable to convert helm-chart to yamls using helm template | Error %v", err)
	}
	os.RemoveAll("temp")
}
