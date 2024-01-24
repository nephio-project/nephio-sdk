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
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type UnstructStringConverter struct {
}

/*
Runs Dfs Algorithm over the Unstructured Object Tree Like Structure
Input:

	v: Current Node of Tree you are at
	tabs: Depth of Current Node

Output:

	Go-Code String Representing the Unstructured Object
*/
func (obj *UnstructStringConverter) runDfsUnstruct(v reflect.Value, tabs int) string {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		log.Fatal("Assumption: Pointer Can Never come in Unstructed |Feels Wrong")
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		curSlice := v.Interface().([]any)
		var outStr = ""
		for _, sliceItem := range curSlice {
			// Run DFS over all the iterations of Slice, and capture the backtrack value
			backtackValues := obj.runDfsUnstruct(reflect.ValueOf(sliceItem), tabs+1)
			outStr += repeat("\t", tabs) + backtackValues + ",\n"
			logrus.Debugf("%s%s,\n", repeat("\t", tabs), backtackValues)
		}
		if len(outStr) > 1 {
			outStr = outStr[:len(outStr)-1] //Removing the Last \n
		}
		return fmt.Sprintf("[]any{\n%s\n%s}", outStr, repeat("\t", tabs))

	case reflect.Map:
		out := ""
		curMap := v.Interface().(map[string]any)
		for key, val := range curMap {
			// Run DFS over all the Values of Map, and capture the backtrack value
			backtackValues := obj.runDfsUnstruct(reflect.ValueOf(val), tabs+1)

			logrus.Debugf("%s\"%s\": %s", repeat("\t", tabs), key, backtackValues)
			out += fmt.Sprintf("%s\"%s\": %s,\n", repeat("\t", tabs), key, backtackValues)
		}
		if len(out) > 1 {
			out = out[:len(out)-1] //Removing the Last \n
		}
		return fmt.Sprintf("map[string]any{\n%s\n%s}", out, repeat("\t", tabs))
	case reflect.String:
		data := v.String()
		// Todo: Need much better handling to strings, Since Different combinations can lead to bad-buggy results
		// Below Additional Replace helps in building integrity of the "" string
		data = strings.ReplaceAll(data, "\\", "\\\\") // Replacing String containing \ with \\
		data = strings.ReplaceAll(data, "\"", "\\\"") // Replacing String containing " with \"
		if strings.Contains(data, "\n") {
			return handleMultiLineStrings(data)
		}
		return "\"" + data + "\"" // Sending with double quotes

	case reflect.Bool:
		return strconv.FormatBool(v.Bool()) // Return the Bool value as String
	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int64:
		return strconv.Itoa(int(v.Int())) // Return the Int value as String
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32) // Return the float32 value as string
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64) // Returns the float64 value as string
	default:
		logrus.Error("Current Type is Not Supported in Unstruct-To-String| ", v.Kind())

	}
	return ""
}

/*
Converts the Unstructured Object to a gocode (string) that can create the same Unstructured Object
*/
func (obj *UnstructStringConverter) Convert(unstructObj unstructured.Unstructured) string {
	outStr := obj.runDfsUnstruct(reflect.ValueOf(unstructObj.Object), 2)
	gocodeStr := fmt.Sprintf(
		`&unstructured.Unstructured{
	Object: %s,
	}`, outStr)
	logrus.Debug("\n ---------Your Generated Code -----------------\n")
	logrus.Debug(gocodeStr)
	return gocodeStr
}
