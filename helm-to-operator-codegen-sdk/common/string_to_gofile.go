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
	"os"
	"strconv"
	"strings"

	"github.com/liyue201/gostl/ds/set"
	"github.com/liyue201/gostl/utils/comparator"
	"github.com/sirupsen/logrus"
)

type GoFile struct {
	Namespace             string
	FileContent           string
	runtimeSupportKindSet set.Set[string] // To be Set By Intialise
}

/*
Input:

	resourceType: The type of Resource: (Service, Deployment etc)
	resourceList: list of gocode of all the resources of the specified resource type

Output: Converts the input into a runnable gofunction
Example: resourceType = Service,  resourceList = ["Svc-1-Code", "Svc-2-Code"]
Returns:

	func Get_Service() []*corev1.Service{
		service_1 := svc-1-Code
		service_2 := svc-2-Code

		return []*corev1.Service{service_1, service_2,}
	}
*/
func (obj *GoFile) getRunnableFunction(resourceType string, resourceList []string) string {
	if len(resourceList) == 0 {
		return ""
	}
	fxnReturnType := ""
	firstResource := resourceList[0]
	for _, c := range firstResource {
		if c == '{' {
			break
		} else if c != '\t' {
			fxnReturnType += string(c)
		}
	}

	afterFxnReturnType, pointerType := strings.CutPrefix(fxnReturnType, "&")
	if pointerType {
		fxnReturnType = "*" + afterFxnReturnType
	}

	varList := ""
	varNamePrefix := fmt.Sprintf("%s%s", strings.ToLower(string(resourceType[0])), resourceType[1:]) // Service --> service
	createdVars := ""
	for i := 0; i < len(resourceList); i++ {
		curVarName := varNamePrefix + strconv.Itoa(i+1)
		varList += fmt.Sprintf(`
	%s := %s
	
		`, curVarName, resourceList[i])
		createdVars += curVarName + ", "
	}

	fxn := fmt.Sprintf(`
func Get%s() []%s{
	%s
	return []%s{%s}
}
	
	`, resourceType, fxnReturnType, varList, fxnReturnType, createdVars)

	logrus.Debug(fxn)
	return fxn
}

/*
It adds the imports as well as the helper fxns like int_ptr, string_ptr
Input:

	allFxn: Go-code for all the fxns (Get_Service(), Get_Deployment()) concatenated in a single string
	fxnCreated: List of all the fxnNames that allFxn contains (Used in getMasterFxn)
	debugging: For Testing (to be removed)

Output:

	A Go Package, containing all the functions, helper functions, required imports, The output of this function is what you see in the generated_code.go
*/
func (obj *GoFile) addFunctionsToGofile(allFxn string, fxnCreated []string, debugging bool) string {
	packageName := "main"
	if !debugging {
		packageName = "controller"
	}
	fileText := fmt.Sprintf(`
package %s

import (
	"context"
	"fmt"
	"time"
	"encoding/base64"
	
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
)

func deleteMeAfterDeletingUnusedImportedModules() {
	/*
		It is written to handle the error "Module Imported but not used",
		The user can delete the non-required modules from import and then delete this function also
	*/
	_ = time.Now()
	_ = &unstructured.Unstructured{}
	_ = corev1.Service{}
	_ = metav1.ObjectMeta{}
	_ = appsv1.Deployment{}
	_ = rbacv1.Role{}
	_ = schedulingv1.PriorityClass{}
	_ = intstr.FromInt(4)
	_, _ = resource.ParseQuantity("")
	_ = context.TODO()
	_ = fmt.Sprintf("")
	_ = ptr.To(32)
}

func int32Ptr(val int) *int32 {
	var a int32
	a = int32(val)
	return &a
}

func int64Ptr(val int) *int64 {
	var a int64
	a = int64(val)
	return &a
}

func intPtr(val int) *int {
	a := val
	return &a
}

func int16Ptr(val int) *int16 {
	var a int16
	a = int16(val)
	return &a
}

func boolPtr(val bool) *bool {
	a := val
	return &a
}

func stringPtr(val string) *string {
	a := val
	return &a
}

func getDataForSecret(encodedVal string) []byte {
	/*
		Concept: Based on my Understanding, corev1.Secret requires the actual data(not encoded) as secret-Data
		But in general terms, we put encoded values in secret-data, which make sense (why to write actual value in readable format)
		This function takes the encodedVal and decodes it and returns
	*/
	decodeVal, err := base64.StdEncoding.DecodeString(encodedVal)
	if err != nil {
		fmt.Println("Unable to decode the SecretVal ", encodedVal, " || This Secret Will Probably would give error during deployment| Kindly Check")
		return []byte(encodedVal)
	}
	return decodeVal
}

`, packageName) + obj.getMasterFxn(fxnCreated, true) + obj.getMasterFxn(fxnCreated, false) + allFxn
	mainfxn := `
	func main(){
		fmt.Println("Only for Debbugging purpose")
		fmt.Println(GetService())
		fmt.Println(GetDeployment())
	}
	`
	if debugging {
		return fileText + mainfxn
	} else {
		return fileText
	}

}

/*
Input:

	fxnCreated: List all functions that has been created so far, Example: [Get_Service(), Get_Deployment()]
	inCreatedState: Bool: true for Create_All, and false for Delete_All

Output:

	Output the Go-Code of the function that can either create or delete all the resources
*/
func (obj *GoFile) getMasterFxn(fxnCreated []string, inCreatedState bool) string {
	usage := "Delete"
	if inCreatedState {
		usage = "Create"
	}
	fxnStatement := ""
	if obj.Namespace == "" {
		for _, fxnName := range fxnCreated {
			fxnStatement += fmt.Sprintf(`
	for _, resource := range %s{
		err = r.%s(context.TODO(), resource)
		if err != nil {
			fmt.Println("Erorr During %sing resource of %s| Error --> |", err)
		}
	} 
			`, fxnName, usage, usage[:len(usage)-1], fxnName)
		}
	} else {
		for _, fxnName := range fxnCreated {
			fxnResourceType := fxnName[3 : len(fxnName)-2] // ResourceType from GetService(), GetDeployment()
			ifblock :=
				`if resource.ObjectMeta.Namespace == ""{
			resource.ObjectMeta.Namespace = namespaceProvided
		}`
			if !obj.runtimeSupportKindSet.Contains(fxnResourceType) {
				// If the resource type is unstructured.Unstructured
				ifblock =
					`if resource.GetNamespace() == ""{
			resource.SetNamespace(namespaceProvided)
		}`
			}

			fxnStatement += fmt.Sprintf(`
	for _, resource := range %s{
		%s
		err = r.%s(context.TODO(), resource)
		if err != nil {
			fmt.Println("Erorr During %sing resource of %s| Error --> |", err)
		}
	} 
			`, fxnName, ifblock, usage, usage[:len(usage)-1], fxnName)
		}
	}

	namespaceVarStatement := ""
	if obj.Namespace != "" {
		namespaceVarStatement = fmt.Sprintf("namespaceProvided := \"%s\"", obj.Namespace)
	}
	outFxn := fmt.Sprintf(`
/*
// Before Uncommenting the following function, Make sure the data-type of r is same as of your Reconciler,
// Replace "YourKindReconciler" with the type of your Reconciler
func (r *YourKindReconciler)%sAll(){
 	var err error
	%s
	%s
}
*/

	`, usage, namespaceVarStatement, fxnStatement)
	return outFxn
}

func (obj *GoFile) Intialise(runtimeSupportKinds []string) {
	var tempSet = set.New[string](comparator.StringComparator, set.WithGoroutineSafe())
	for _, val := range runtimeSupportKinds {
		tempSet.Insert(val)
	}
	obj.runtimeSupportKindSet = *tempSet
}

/*
Input: Map of Resource-Type as Key and the Value represents Go-Codes corresponding to the resource-type in slice
Example: "Service": ["GO-Code for Service-1", "GO-Code for Service-2"]

	"Deployment": ["GO-Code for Deployment-1", "GO-Code for Deployment-2", "GO-Code for Deployment-3"]

Output:

	Generates the Go-file String Content containing all the functions and libray imports, so the gocode can be deployed/ pluged in
*/
func (obj *GoFile) Generate(gocodes map[string][]string) {
	allFxn := ""
	functionsCreated := []string{}
	for resourceType, resourceList := range gocodes {
		allFxn += obj.getRunnableFunction(resourceType, resourceList)
		functionsCreated = append(functionsCreated, fmt.Sprintf("Get%s()", resourceType))
	}
	fileText := obj.addFunctionsToGofile(allFxn, functionsCreated, false)
	obj.FileContent = fileText
}

/*
Writes the "output_gocode/generated_code.go" file
*/
func (obj *GoFile) WriteToFile() {
	_ = createDirIfDontExist("outputs")
	err := os.WriteFile("outputs/generated_code.go", []byte(obj.FileContent), 0600)
	if err != nil {
		logrus.Fatal("Writing gocode to outputs/generated_code.go FAILED| Error --> | ", err)
	}
}
