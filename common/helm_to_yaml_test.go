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
