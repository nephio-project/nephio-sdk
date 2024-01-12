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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RuntimeJsonConverter struct {
}

func (obj *RuntimeJsonConverter) refactorHelmLabels(labels map[string]interface{}) map[string]interface{} {
	/*
		According to helm-k8s docs:
		Standard-Labels = [app.kubernetes.io/name, helm.sh/chart, app.kubernetes.io/managed-by, app.kubernetes.io/instance, app.kubernetes.io/component, app.kubernetes.io/part-of]
		as of : https://helm.sh/docs/chart_best_practices/labels/
		According to me:
		Helm-Specific-Labels are: [helm.sh/chart, app.kubernetes.io/managed-by]
		ToDo: to check above intution
	*/
	helmSpecificLabels := []string{"helm.sh/chart", "app.kubernetes.io/managed-by"}
	for _, prohibitedLabels := range helmSpecificLabels {
		_, ok := labels[prohibitedLabels]
		if ok {
			delete(labels, prohibitedLabels)
		}
	}
	return labels
}

/*
Recursive Function (DFS Algorithm) to traverse the object structure and identify data-types of various attributes
If you see the Runtime-Object as a Hierachial Structure (Tree), then you say curObj would be the node of the tree/graph you are currently at
The DFS Algorithm would traverse all the nodes, but will not return empty fields
*/
func (obj *RuntimeJsonConverter) runDfsJsonOmitEmpty(curObj interface{}, tabs int) interface{} {

	// Handling Special Cases, When We can't move further because of Private Attributes
	if reflect.TypeOf(curObj).String() == "resource.Quantity" {
		res := curObj.(resource.Quantity)
		resourceVal := res.String()
		if resourceVal == "0" {
			return nil
		}
		// Hack: type is changed to int, because we don't want the value in double quote when converting it to string
		return map[string]string{"type": "int", "val": fmt.Sprintf("resource.MustParse(\"%s\")", resourceVal)}
	} else if reflect.TypeOf(curObj).String() == "v1.Time" {
		/*
			Since the attributes of v1.Time struct are private, Therefore we need to send back the value using GoString() method
		*/
		timeInter := curObj.(metav1.Time)
		timeVar := timeInter.Time
		defaultVal := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
		if timeVar == defaultVal { //If The LHS is the Default Value, then omit it
			return nil
		} else {
			var out = make(map[string]interface{})
			timeVal := timeVar.GoString()
			out["Time"] = map[string]string{"type": "int", "val": timeVal} // Hack: type is changed to int, because we don't want the value in double quote
			return out
		}
	}
	// Private Attributes Special Cases Handling  End

	objRef := reflect.ValueOf(curObj)
	if objRef.Kind() == reflect.Ptr {
		objRef = objRef.Elem() // Dereferencing the Pointer
	}

	switch objRef.Kind() {
	case reflect.Struct:
		var out = make(map[string]interface{})
		for i := 0; i < objRef.NumField(); i++ {
			var inter = make(map[string]interface{})
			if !objRef.Field(i).CanInterface() {
				logrus.Warn("Private Attributes are not visible to me ! Support Missing For || ", objRef.Type().Field(i).Name, objRef.Type().Field(i).Type)
				continue
			}
			// Run DFS over the attributes (Fields) of current Struct
			backtrackVal := obj.runDfsJsonOmitEmpty(objRef.Field(i).Interface(), tabs+1)
			if backtrackVal != nil {
				inter["type"] = objRef.Type().Field(i).Type.String() // Type of i'th Field
				inter["val"] = backtrackVal                          // Backtracked/Actual Value of i'th Field
				attributeName := objRef.Type().Field(i).Name
				if attributeName == "Labels" {
					inter["val"] = obj.refactorHelmLabels(backtrackVal.(map[string]interface{}))
				}
				out[attributeName] = inter // Save the (Type and Value of i'th Field) with key as i'th Field Name
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int64:
		data := strconv.Itoa(int(objRef.Int()))
		if data == "0" { // 0 is considered to be default value, Therefore, Omitting it
			return nil
		}
		return data
	case reflect.Float32:
		data := strconv.FormatFloat(objRef.Float(), 'f', -1, 32) // Converts the Obj Val to String Val
		if data == "0" {                                         // 0 is considered to be default value, Therefore, Omitting it
			return nil
		}
		return data
	case reflect.Float64:
		data := strconv.FormatFloat(objRef.Float(), 'f', -1, 64) // Converts the Obj Val to String Val
		if data == "0" {                                         // 0 is considered to be default value, Therefore, Omitting it
			return nil
		}
		return data
	case reflect.Bool:
		data := objRef.Bool()
		return data
	case reflect.String:
		data := objRef.String()
		if data == "" { // "" is considered to be default value, Therefore, Omitting it
			return nil
		}
		// Todo: Need much better handling to strings, Since Different combinations can lead to bad-buggy results
		// Below Additional Replace helps in building integrity of the "" string
		data = strings.ReplaceAll(data, "\\", "\\\\") // Replacing String containing \ with \\
		data = strings.ReplaceAll(data, "\"", "\\\"") // Replacing String containing " with \"
		return data
	case reflect.Slice:
		var out []interface{}
		if objRef.Len() == 0 {
			return nil
		}
		sliceElementType := objRef.Index(0).Type().String()
		// Special Cases:
		// Case 1: If The Slice represents a Byte, then its Element data-type would be uint8
		if sliceElementType == "uint8" {
			// Assuming that the byte has come from Kind: Secret, So, we need to encode the string to base64, before writing in code
			// Thought: You never write the actual value of secret in yaml, but the encoded versions of it, The same is happening below
			// Todo: Finding out where []byte is used other than Secret, and if it also represent encoded version or actual string
			byteVal := objRef.Interface()
			encodedByteVal := base64.StdEncoding.EncodeToString(byteVal.([]byte))
			return encodedByteVal
		}
		for i := 0; i < objRef.Len(); i++ {
			// Run DFS over the all the iterations of current slice and capture the backtrack value
			backtrackVal := obj.runDfsJsonOmitEmpty(objRef.Index(i).Interface(), tabs+1)
			if backtrackVal != nil {
				out = append(out, backtrackVal)
			} else {
				// Sometimes, "" in a slice, cannot be considered as nil value, For Examples for writing api-groups in Kind:Role, We write "" in a slice
				// Argument: If a person has written "" in a slice, then It might be written for a reason
				// Pending: Type Checking, Todo: if sliceElementType is int, then append 0, if bool, then append false, if string then, "" (Required Or Not)
				out = append(out, "")
			}

		}
		if len(out) == 0 {
			return nil
		}
		return out

	case reflect.Map:
		// Assumption : Key Value is Always String
		var out = make(map[string]interface{})
		switch objRef.Type().Key().Kind() {
		case reflect.String:
			for _, key := range objRef.MapKeys() {
				// Run DFS over all the Values of current Map and Capture the Output
				backtrackVal := obj.runDfsJsonOmitEmpty(objRef.MapIndex(key).Interface(), tabs+1)
				// Argument: Why would someone add "" value of a key, if it is not useful i.e (key1 : ""),
				// Pending: Type Checking, Todo: if val_type is int, then add 0, if bool, then add false, if string then, "" (Required Or Not)
				if backtrackVal != nil {
					out[key.String()] = backtrackVal
				}

			}
		default:
			logrus.Warn("Currently Map-keys with the following Kind ", objRef.Type().Key().Kind(), " Are not Supported")
		}
		if len(out) == 0 {
			return nil
		}
		return out

	case reflect.Invalid:
		return nil

	default:
		logrus.Fatal("Unsupported Type-Kind Found| Runtime-Json.Go|   ", objRef.Kind())
	}
	return "It shouldn't have reached here"
}

/*
Input: Runtime-Obj, Group-Version-Kind
Output: Writes Temp.json which represents the structure(Heirarchy) and corresponding data-types & values of the Runtime-Object
Example:

	{
		"ApiVersion": {
			"type": "string"
			"val" : "v1"
		},
		"Spec": {
			"type": "v1.DeploymentSpec"
			"val": {
				"Replicas": {
					"type" : "&int32",
					"val" : 5
				}
			}
		}
	} & so on
*/
func (obj *RuntimeJsonConverter) Convert(runtimeObj runtime.Object, gvk schema.GroupVersionKind) error {
	if gvk.Version != "v1" {
		logrus.Error("Currently Only Api-Version v1 is supported (Skipping)| Given Version " + gvk.Version)
		return fmt.Errorf("currently only Api-version v1 is supported| Given Version " + gvk.Version)
	}
	logrus.Debug("----------------------------------Your Runtime Object--------------------\n", runtimeObj)
	logrus.Debug("----------------------------------Your Runtime Object Ends--------------------\n")
	var objMap interface{}
	switch gvk.Kind {
	case "Service":
		curObj := runtimeObj.(*corev1.Service)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "Deployment":
		curObj := runtimeObj.(*appsv1.Deployment)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "ConfigMap":
		curObj := runtimeObj.(*corev1.ConfigMap)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "ServiceAccount":
		curObj := runtimeObj.(*corev1.ServiceAccount)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "PersistentVolumeClaim":
		curObj := runtimeObj.(*corev1.PersistentVolumeClaim)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "StatefulSet":
		curObj := runtimeObj.(*appsv1.StatefulSet)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "Secret":
		curObj := runtimeObj.(*corev1.Secret)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "PriorityClass":
		curObj := runtimeObj.(*schedulingv1.PriorityClass)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "Role":
		curObj := runtimeObj.(*rbacv1.Role)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "RoleBinding":
		curObj := runtimeObj.(*rbacv1.RoleBinding)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "ClusterRole":
		curObj := runtimeObj.(*rbacv1.ClusterRole)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	case "ClusterRoleBinding":
		curObj := runtimeObj.(*rbacv1.ClusterRoleBinding)
		objMap = obj.runDfsJsonOmitEmpty(curObj, 0)
	default:
		logrus.Warn("Kind Currently Not Supported  | ", gvk.Kind)
		return fmt.Errorf("kind Currently Not Supported  | %s", gvk.Kind)
	}

	logrus.Debug("----------------------------------Your JSON Map--------------------\n", objMap)
	logrus.Debug("----------------------------------Your JSON Map Ends--------------------\n")
	jsonString, jsonErr := json.MarshalIndent(objMap, "", "    ")
	if jsonErr != nil {
		return jsonErr
	}
	_ = createDirIfDontExist("temp")
	err := os.WriteFile("temp/temp.json", jsonString, 0600)
	return err
}
