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
	"helm_to_controller/packages/common"
	"os"
	"strings"

	"github.com/liyue201/gostl/ds/set"
	"github.com/liyue201/gostl/utils/comparator"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/kubectl/pkg/scheme"
)

var runtimeSupportKinds = []string{"Deployment", "Service", "Secret", "Role", "RoleBinding", "ClusterRoleBinding",
	"PersistentVolumeClaim", "StatefulSet", "ServiceAccount", "ClusterRole", "PriorityClass", "ConfigMap"}
var runtimeSupportKindSet = set.New[string](comparator.StringComparator, set.WithGoroutineSafe())

func init() {
	// Runtime Support Kind Set contains the set of all kinds whose runtime-support is handled by the script
	for _, val := range runtimeSupportKinds {
		runtimeSupportKindSet.Insert(val)
	}
}

/*
Convert the KRM Resources to *unstructured.Unstructured (map[string]any)
Returns the *unstructured.Unstructured Object, GroupVersionKind (gvk), error
*/
func unstructuredDecode(data []byte) (*unstructured.Unstructured, *schema.GroupVersionKind, error) {
	obj := &unstructured.Unstructured{}
	// Decode YAML into unstructured.Unstructured
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(data, nil, obj)
	if err != nil {
		return nil, nil, err
	}
	return obj, gvk, nil
}

/*
Input: Reads the yaml file from filepath
Output:

	runtimeObjList: List of runtime Objects Converted from the input yaml
	gvkList		: List of Group-Version-Kind for the runtime objects of runtimeObjList, mapped Index-wise
	unstructObjList: List of unstructured Objects Converted from the input yaml, whose Kind are not default to kubernetes| Third Party Kinds
	unstructGvkList: List of Group-Version-Kind for the unstructured objects of unstructObjList, mapped Index-wise
*/
func handleSingleYaml(inputFilepath string) (runtimeObjList []runtime.Object, gvkList []schema.GroupVersionKind, unstructObjList []unstructured.Unstructured, unstructGvkList []schema.GroupVersionKind) {
	data, err := common.GetFileContents(inputFilepath)
	if err != nil {
		logrus.Error("Error While Reading YAML file | ", inputFilepath, " \t |", err)
		return
	}
	// A Single yaml can contain muliple KRM reosurces, separated by ---, Therefore Spliting the yaml-file-content over "---" to get single  KRM Resource
	for _, doc := range strings.Split(string(data), "\n---") {
		if doc == "" {
			continue
		}
		// Parsing the KRM Resource to get the Kind which will decide to use either runtime-object-method or unstructured.Unstructured method
		unstructObject, gvk, err := unstructuredDecode([]byte(doc))
		if err != nil {
			logrus.Error("Unable to convert yaml to unstructured |", err)
			continue
		}
		resourceKind := gvk.Kind
		if runtimeSupportKindSet.Contains(resourceKind) {
			// Handle the current yaml with runtimeObject method
			decoder := scheme.Codecs.UniversalDeserializer()
			runtimeObject, gvk, err := decoder.Decode([]byte(doc), nil, nil)
			if err != nil {
				logrus.Error("Cant decode the section of yaml, by Runtime-Object \t |", err)
				continue
			}
			runtimeObjList = append(runtimeObjList, runtimeObject)
			gvkList = append(gvkList, *gvk)
		} else {
			logrus.Info("Kind | ", resourceKind, " Would Be Treated as Third Party Kind")
			unstructObjList = append(unstructObjList, *unstructObject)
			unstructGvkList = append(unstructGvkList, *gvk)
		}
	}
	return
}

func setLogLevel(loggingLvl string) {
	ll, err := logrus.ParseLevel(loggingLvl)
	if err != nil {
		ll = logrus.DebugLevel
	}
	logrus.SetLevel(ll)
}

func main() {
	curHelmChart := "inputs"
	cmdArgs := os.Args[1:]
	if len(cmdArgs) != 0 {
		curHelmChart = cmdArgs[0]
	}
	namespace := ""
	if len(cmdArgs) >= 2 {
		namespace = cmdArgs[1]
	}

	loggingLvl := "info"
	if len(cmdArgs) >= 3 {
		loggingLvl = cmdArgs[2]
	}
	setLogLevel(loggingLvl)

	var helmYamlConvertor = common.HelmYamlConvertor{Namespace: namespace, Chartpath: curHelmChart}
	err := helmYamlConvertor.ConvertHelmToYaml()
	if err != nil {
		logrus.Fatal("Unable to Convert Helm to Yamls| Error | ", err)
	}
	allYamlPaths := common.RecursiveListYamls("temp/templated")
	// Intialising Convertor Structs/Classes
	var jsonStringConverterObj = common.JsonStringConverter{}
	jsonStringConverterObj.Intialise()
	var goFileObj = common.GoFile{Namespace: namespace}
	goFileObj.Intialise(runtimeSupportKinds)
	var runtimeJsonConverterObj = common.RuntimeJsonConverter{}
	var unstructStringConverterObj = common.UnstructStringConverter{}

	// Loop over each Yaml File (recursively) and get their gocodes
	var gocodes = map[string][]string{}
	for _, yamlfile := range allYamlPaths {
		logrus.Info("CurFile --> | ", yamlfile)
		runtimeObjList, gvkList, unstructObjList, unstructGvkList := handleSingleYaml(yamlfile)
		for i := 0; i < len(runtimeObjList); i++ {
			logrus.Info(fmt.Sprintf(" Current KRM Resource| Kind : %s| YamlFilePath : %s", gvkList[i].Kind, yamlfile))
			err := runtimeJsonConverterObj.Convert(runtimeObjList[i], gvkList[i])
			if err != nil {
				logrus.Error("\t Converting Runtime to Json Failed (Skipping Current Resource)| Error : ", err)
				continue
			}

			logrus.Info("\t Converting Runtime to Json Completed")
			gocodeStr, err := jsonStringConverterObj.Convert(gvkList[i])
			if err != nil {
				logrus.Info("\t Converting Json to String Failed (Skipping Current Resource)| Error : ", err)
				continue
			}
			gocodes[gvkList[i].Kind] = append(gocodes[gvkList[i].Kind], gocodeStr)
			logrus.Info("\t Converting Json to String Completed ")
		}

		for i := 0; i < len(unstructObjList); i++ {
			gocode := unstructStringConverterObj.Convert(unstructObjList[i])
			gocodes[unstructGvkList[i].Kind] = append(gocodes[unstructGvkList[i].Kind], gocode)
			logrus.Info("\t Converting Unstructured to String Completed ")
		}
	}
	logrus.Info("----------------- Writing GO Code ---------------------------------")
	goFileObj.Generate(gocodes)
	goFileObj.WriteToFile()
	logrus.Info("----------------- Program Run Successful| Summary ---------------------------------")
	for resourceType, resourceList := range gocodes {
		logrus.Info(resourceType, "\t\t |", len(resourceList))
	}
	err = os.RemoveAll("temp")
	if err != nil {
		logrus.Warn("Failed to delete the Temp Directory| Error | ", err)
		logrus.Warn("Manual Delete Advised| temp")
	}

}
