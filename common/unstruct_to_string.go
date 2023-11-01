package common

import (
	"fmt"
	"log"
	"reflect"
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
		curSlice := v.Interface().([]interface{})
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
		return fmt.Sprintf("[]interface{}{\n%s\n%s}", outStr, repeat("\t", tabs))

	case reflect.Map:
		out := ""
		curMap := v.Interface().(map[string]interface{})
		for key, val := range curMap {
			// Run DFS over all the Values of Map, and capture the backtrack value
			backtackValues := obj.runDfsUnstruct(reflect.ValueOf(val), tabs+1)

			logrus.Debugf("%s\"%s\": %s", repeat("\t", tabs), key, backtackValues)
			out += fmt.Sprintf("%s\"%s\": %s,\n", repeat("\t", tabs), key, backtackValues)
		}
		if len(out) > 1 {
			out = out[:len(out)-1] //Removing the Last \n
		}
		return fmt.Sprintf("map[string]interface{}{\n%s\n%s}", out, repeat("\t", tabs))
	case reflect.String:
		data := v.String()

		if strings.Contains(data, "\n") {
			// New Lines Are now handled fmt.Sprint
			data = fmt.Sprintf("fmt.Sprint(`%s`)", data)
			return data
		}
		data = strings.ReplaceAll(data, "\"", "\\\"") // Need to preserve double quotes, therefore preserving by adding a backslash (" --> /")
		return "\"" + data + "\""                     // Sending with double quotes

	case reflect.Bool:
		return fmt.Sprint(v.Bool()) // Return the Bool value as String
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int64:
		return fmt.Sprint(v) // Return the Int, Float value as String
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
