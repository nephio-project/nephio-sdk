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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/liyue201/gostl/ds/set"
	"github.com/liyue201/gostl/ds/stack"
	"github.com/liyue201/gostl/utils/comparator"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type JsonStringConverter struct {
	globalStructMapping map[string]string // To be set by Calling Intialise (setModStructMapping)
	globalEnumsSet      set.Set[string]   // To be set by Calling Intialise (setEnums)
	generatedGoCode     string            // To be set by jsonToGoCode Function
}

func (obj *JsonStringConverter) checkOpeningAndClosingBraces() bool {
	s := stack.New[int]()
	lineCount := 0
	opCount, closeCount := 0, 0
	stringDetected := false
	for i := 0; i < len(obj.generatedGoCode); i++ {
		if obj.generatedGoCode[i] == '\n' {
			lineCount++
		}
		if obj.generatedGoCode[i] == '"' {
			// If Enters in double quotes (string) Don't Check For opening/Closing Braces
			stringDetected = !stringDetected
		}
		if !stringDetected {
			if obj.generatedGoCode[i] == '{' {
				s.Push('{')
				opCount++
			} else if obj.generatedGoCode[i] == '}' {
				closeCount++
				if s.Empty() {
					logrus.Error("Closing Brace has no Opening Brace| Linenumber ", lineCount)
					return false
				} else {
					s.Pop()
				}
			}
		}

	}

	logrus.Debugf("Test Summary| Opening Brace Count %d | Closing Brace Count %d| \n", opCount, closeCount)
	for !s.Empty() {
		logrus.Error("Extra Opening Braces Found at the End ")
		return false
	}

	return true
}

/*
Checks for Generated-Go-Code
Check 1: Opening and Closing Braces should be consistent (as you see in day to day like)
More Checks can also be added here
*/
func (obj *JsonStringConverter) checkGoCode() {
	obj.checkOpeningAndClosingBraces()
}

/*
Reads the module-structure mapping.json and populate globalStructMapping map
globalStructMapping Says: Give me any data-type (v1.Deployment) and I will tell you which module it belong (appsv1)
This is required, because runtime-Object type if you reflect, will say the datatype as v1.Service or v1.DeploymentSpec,
whereas in code, we want corev1.Service or appsv1.DeploymentSpec
*/
func (obj *JsonStringConverter) setModStructMapping(inputFilepath string) {

	out := map[string]string{}
	plan, _ := os.ReadFile(filepath.Clean(inputFilepath))
	var data map[string]interface{}
	err := json.Unmarshal(plan, &data)
	if err != nil {
		logrus.Fatal("Cannot unmarshal the json For Struct Mapping | ", inputFilepath, "  Error | ", err)
	}

	// Validating Struct Mapping, Every Mapping value should contain unique values: pending
	var validateSet = set.New[string](comparator.StringComparator, set.WithGoroutineSafe())
	for _, structsList := range data {
		for _, structName := range structsList.([]interface{}) {
			// ToDo: What to do when duplication is found, currently we are only logging
			if validateSet.Contains(structName.(string)) {
				logrus.Warn("Duplication Detected in Struct Mapping | For ", structName.(string))
			}
			validateSet.Insert(structName.(string))
		}

	}

	// Saving to Global Map, so that it could be used by "formatTypeVal" function
	for module, structs := range data {
		for _, structName := range structs.([]interface{}) {
			structNameStr := structName.(string)
			out[structNameStr] = module
		}
	}
	obj.globalStructMapping = out
}

/*
	 Enums are needed to handle differently than structs, therefore the below set tells which data-types are enum (non-composite),
		So, that it could be handled differently (Used by "formatTypeVal" function)
*/
func (obj *JsonStringConverter) setEnums(inputFilepath string) {
	fp, _ := os.ReadFile(filepath.Clean(inputFilepath))
	var data map[string]interface{}
	err := json.Unmarshal(fp, &data)
	if err != nil {
		logrus.Fatal("Cannot unmarshal the json For Enum-Mapping | ", inputFilepath, "  Error | ", err)
	}
	var tempSet = set.New[string](comparator.StringComparator, set.WithGoroutineSafe())
	// Saving to Global Enum Set, so that it could be used by "formatTypeVal" function
	for _, enums := range data {
		for _, val := range enums.([]interface{}) {
			enum := val.(string)
			if tempSet.Contains(enum) {
				// ToDo: What to do when duplication is found, currently we are only logging
				logrus.Warn("Duplication Detected in Enum Mapping | For ", enum)
			}
			tempSet.Insert(enum)
		}
	}
	obj.globalEnumsSet = *tempSet
}

/*
Based on different data-type, values are formated differently
Example	objType	objVal		Format_Val(Out)

	String		5			"5"
	Int32		5			5
	*int32		5			int32Ptr(5)
*/
func (obj *JsonStringConverter) formatTypeVal(objType string, objVal string, tabCount int) string {
	// Special Data-Types are Handled Here
	if objType == "intstr.Type" {
		/*intstr.Type need to be handled explictly: intstr.Type is a int enum*/
		objVal = objVal[1 : len(objVal)-1] // Removing the double quotes from the objVal (because we need int)
		return fmt.Sprintf("%s(%s)", objType, objVal)
	} else if objType == "[]uint8" || objType == "[]byte" {
		// Generally []uint8 is only used for secret
		return fmt.Sprintf("getDataForSecret(%s)", objVal)
	}
	// Special Data-Types area Ends

	pointerType := false
	if objType[0] == '*' {
		pointerType = true
	} else if objType[0] == '&' {
		log.Fatal(fmt.Errorf("& Types are not supported yet"))
	}

	if pointerType {
		switch objType[1:] {
		// You can find the defination of func boolPtr, int32Ptr, int64Ptr, intPtr, int16Ptr in the string_to_gocode.go
		case "int", "int16", "int32", "int64":
			return fmt.Sprintf("%sPtr(%s)", objType[1:], objVal[1:len(objVal)-1])
		case "bool":
			return fmt.Sprintf("boolPtr(%s)", objVal)
		case "string":
			return fmt.Sprintf("stringPtr(%s)", objVal)
		}
	}

	switch objType {
	case "int32", "int64", "int", "int16":
		return objVal[1 : len(objVal)-1] // Remove the double quotes and return
	case "bool":
		return objVal
	case "string":
		return objVal
	}

	// It will reach here If It is a Composite Literal i.e. Struct OR a Enum
	// Step-1: If type contains v1 --> Needs to change with corresponding modules using the globalStructMapping
	if pointerType {
		objType = "&" + objType[1:] //Converting pointer to address
	}
	re := regexp.MustCompile(`v1.`)
	index := re.FindStringIndex(objType)

	if index != nil {
		// v1 is present in objType
		startIndex := index[0]
		endIndex := index[1]
		curStruct := objType[endIndex:] // Converts &v1.DeploymentSpec --> DeploymentSpec and assigns to curStruct
		module := obj.globalStructMapping[curStruct]
		if module == "" {
			logrus.Error("Current Structure-Module Mapping is NOT KNOWN| Kindly add it in module_struct_mapping.json | ", objType)
		}
		objTypeWithModule := objType[:startIndex] + module + "." + curStruct // Converts &v1.DeploymentSpec --> &appsv1.DeploymentSpec
		/*
			The only difference between Enum and Structs is that:
			Enum are intialised using () where Structs are intialised using {}
			Therefore, Need to Handle Separatly
		*/
		if obj.globalEnumsSet.Contains(curStruct) && objType[:2] != "[]" { // List of enums([]enumtype) are also Intailised as Structs using {}
			return fmt.Sprintf("%s(%s)", objTypeWithModule, objVal) // Replacing {} with (), For Enums
		} else {
			return fmt.Sprintf("%s{\n%s\n%s}", objTypeWithModule, objVal, repeat("\t", tabCount))
		}
	}

	return fmt.Sprintf("%s{\n%s\n%s}", objType, objVal, repeat("\t", tabCount))
}

/*
Recursive Function (DFS Algorithm) to traverse json and convert to gocode
The DFS Algorithm would traverse all the nodes(represented by v) writes its corressponding gocode
*/
func (obj *JsonStringConverter) traverseJson(v reflect.Value, curObjType string, tabs int) string {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem() // Dereference the Pointer
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		inter_str := "\n"
		objType := curObjType[2:] // Removing the []
		for i := 0; i < v.Len(); i++ {
			// Run DFS Over each iterations of slice and capture the backtrack-values
			backtrackVal := obj.traverseJson(v.Index(i), objType, tabs+1)
			inter_str += repeat("\t", tabs) + obj.formatTypeVal(objType, backtrackVal, tabs)
			inter_str += ",\n" // Add a Comma plus a New Line after every value
		}
		/*
			Inter_str would like Something Like this,
			"backtrackVal1",
			"backtrackVal2",
			"backtrackVal3",

		*/

		return inter_str
	case reflect.Map:
		out := ""
		for _, key := range v.MapKeys() {
			// Here key represents the struct Attribute Name/ Field Name
			objMap, _ := v.MapIndex(key).Interface().(map[string]interface{})
			logrus.Debug(repeat("\t", tabs), key)
			objType := objMap["type"].(string) // objType represents the type of i'th attribute
			logrus.Debug(repeat("\t", tabs) + objType)
			objVal := objMap["val"] // objVal represents the value of i'th attribute
			if len(objType) > 5 && (objType[1:4] == "map" || objType[0:3] == "map") {
				// If objType/ Attribute type is Map. Then it is handled here
				mapStartIndex := 0
				if objType[1:4] == "map" {
					mapStartIndex = 1
				}
				// Assuming the Key in Map is always string, map[string] --> 11 Characters
				mapValuesType := objType[mapStartIndex+11:]
				curMap := objVal.(map[string]interface{})

				backTrackValues := ""
				for curKey, curVal := range curMap {
					logrus.Debug(repeat("\t", tabs), curKey)
					// Run DFS over the Values of the map that is contained by i'th attribute as its value
					backtrackVal := obj.traverseJson(reflect.ValueOf(curVal), objType, tabs+1)
					backTrackValues += fmt.Sprintf("%s\"%s\" : %s,\n", repeat("\t", tabs+1), curKey, obj.formatTypeVal(mapValuesType, backtrackVal, tabs))
				}
				backTrackValues = backTrackValues[:len(backTrackValues)-1] // Removing the Last Extra Comma
				out = out + fmt.Sprintf("%s%s : %s,\n", repeat("\t", tabs), key, obj.formatTypeVal(objType, backTrackValues, tabs))
				/*
					out would look like
					Attribute-1: map[string]some_datatype{
						"key1": "backtrackVal1",
						"key2": "backtrackVal2"
					}
				*/
			} else {
				// If objType/ Attribute type is Not Map, It could be String, Any other Struct, Int, Slice etc
				// Run DFS over the objVal which is the value of i'th attribute
				backtrackVal := obj.traverseJson(reflect.ValueOf(objVal), objType, tabs+1)
				// Special Case: If type is resourceList
				if curObjType == "v1.ResourceList" {
					// Need Extra Double-Quotes At Key (cpu, ephemoral-storage)
					out = out + fmt.Sprintf("%s\"%s\"  : %s, \n", repeat("\t", tabs), key, obj.formatTypeVal(objType, backtrackVal, tabs))
				} else {
					out = out + fmt.Sprintf("%s%s  : %s, \n", repeat("\t", tabs), key, obj.formatTypeVal(objType, backtrackVal, tabs))
				}
				/*
					out would look something Like:
					Replicas: 32,
					Containers: []corev1.Container{
						corev1.Container{
							Image: "hello-world" // These Values are a Result of Backtracking
						},

					}
				*/
			}
		}
		out = out[:len(out)-1] // Removing the last new line
		return out

	case reflect.String:
		logrus.Debug(repeat("\t", tabs), v.String())
		data := v.String()
		if strings.Contains(data, "\n") {
			return handleMultiLineStrings(data)
		}
		return "\"" + data + "\"" // Need to return the output with double quotes, " --> /"

	case reflect.Bool:
		logrus.Debug(repeat("\t", tabs), v.Bool())
		return fmt.Sprint(v.Bool()) // Return string version of bool

	default:
		logrus.Fatal("Unsupported Kind For Json-String DFS Traversal|  ", v.Kind())
	}
	return "\nOops, This should be Never Returned\n"
}

/*
Reads the temp.json created by runtime_to_json.go and traverse json(DFS Runner) and generates the go-code requried
*/
func (obj *JsonStringConverter) jsonToGoCode(inputFilepath string) {

	plan, _ := os.ReadFile(filepath.Clean(inputFilepath))
	var data map[string]interface{}
	err := json.Unmarshal(plan, &data)
	if err != nil {
		fmt.Println("Cannot unmarshal the json ", err)
	}
	logrus.Debug("Json Data", data)
	obj.generatedGoCode = obj.traverseJson(reflect.ValueOf(data), "", 2)

	logrus.Debug(" --------------Check-Your Go Code --------------------------")
	logrus.Debug(obj.generatedGoCode)
	logrus.Debug("Running GO-Code Checks")
	obj.checkGoCode()
}

/*
Intialises the Module-Struct Mapping and Enum-Struct Mapping (Used in Format-Val Function)
*/
func (obj *JsonStringConverter) Intialise() {
	// To Make Sure Intialise is only called once
	if len(obj.globalStructMapping) == 0 {
		obj.setModStructMapping("config/struct_module_mapping.json")
		obj.setEnums("config/enum_module_mapping.json")
	}
}

/*
Reads the temp.json created by runtime_to_json.go and Builds gocode string based on the contents of temps.json
*/
func (obj *JsonStringConverter) Convert(gvk schema.GroupVersionKind) (string, error) {
	if gvk.Version != "v1" {
		logrus.Error("Currently Only Api-Version v1 is supported")
		return "", fmt.Errorf("currently Only Api-Version v1 is supported")
	}

	module := obj.globalStructMapping[gvk.Kind]
	if module == "" {
		logrus.Warn("FATAL ERROR| Kind  " + gvk.Kind + "  Currently Not Supported")
		return "", fmt.Errorf("FATAL ERROR| Kind  " + gvk.Kind + "  Currently Not Supported")
	}

	objType := "&" + module + "." + gvk.Kind

	obj.jsonToGoCode("temp/temp.json")
	gocode := fmt.Sprintf("%s{\n%s\n\t}", objType, obj.generatedGoCode)
	return gocode, nil
}
