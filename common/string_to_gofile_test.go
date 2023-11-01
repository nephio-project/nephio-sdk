package common

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

var goFileObj = GoFile{Namespace: "default"}

func TestGoFileIntialise(t *testing.T) {
	ll, _ := logrus.ParseLevel("fatal")
	logrus.SetLevel(ll)

	var runtimeSupportKinds = []string{"Deployment", "Service", "Secret", "Role", "RoleBinding", "ClusterRoleBinding",
		"PersistentVolumeClaim", "StatefulSet", "ServiceAccount", "ClusterRole", "PriorityClass", "ConfigMap"}

	goFileObj.Intialise(runtimeSupportKinds)
	if goFileObj.runtimeSupportKindSet.Size() == 0 {
		t.Error("Runtime Support Kind Set Failed to Intialise")
	}
}

func TestGetRunnableFunction(t *testing.T) {
	result := goFileObj.getRunnableFunction("Deployment", []string{"appsv1.Deployment{struct_attributes...}"})
	expectedLines := []string{"func GetDeployment() []appsv1.Deployment{", "deployment1 := appsv1.Deployment{struct_attributes...}",
		"return []appsv1.Deployment{deployment1, }"}
	for _, expected := range expectedLines {
		if !strings.Contains(result, expected) {
			t.Errorf("Current Line '%s' Not Found in Runnable Function| Actual Output : %s \n", expected, result)
			break
		}
	}

}

func TestGetMasterFxnCreateAll(t *testing.T) {
	result := goFileObj.getMasterFxn([]string{"GetDeployment"}, true)
	expectedContent := `
/*
// Before Uncommenting the following function, Make sure the data-type of r is same as of your Reconciler,
// Replace "YourKindReconciler" with the type of your Reconciler
func (r *YourKindReconciler)CreateAll(){
	var err error
	namespaceProvided := "default"

	for _, resource := range GetDeployment{
		if resource.GetNamespace() == ""{
			resource.SetNamespace(namespaceProvided)
		}
		err = r.Create(context.TODO(), resource)
		if err != nil {
			fmt.Println("Erorr During Creating resource of GetDeployment| Error --> |", err)
		}
	}

}
*/
`
	expectedLines := strings.Split(expectedContent, "\n")
	for _, expected := range expectedLines {
		if !strings.Contains(result, expected) {
			t.Errorf("Current Line '%s' Not Found in Runnable Function| Actual Output : %s \n", expected, result)
			break
		}
	}
}

func TestGetMasterFxnDeleteAll(t *testing.T) {
	result := goFileObj.getMasterFxn([]string{"GetDeployment"}, false)
	expectedContent := `
/*
// Before Uncommenting the following function, Make sure the data-type of r is same as of your Reconciler,
// Replace "YourKindReconciler" with the type of your Reconciler
func (r *YourKindReconciler)DeleteAll(){
	var err error
	namespaceProvided := "default"

	for _, resource := range GetDeployment{
		if resource.GetNamespace() == ""{
			resource.SetNamespace(namespaceProvided)
		}
		err = r.Delete(context.TODO(), resource)
		if err != nil {
			fmt.Println("Erorr During Deleting resource of GetDeployment| Error --> |", err)
		}
	}

}
*/
`
	expectedLines := strings.Split(expectedContent, "\n")
	for _, expected := range expectedLines {
		if !strings.Contains(result, expected) {
			t.Errorf("Current Line '%s' Not Found in Runnable Function| Actual Output : %s \n", expected, result)
			break
		}
	}
}

func TestGenerate(t *testing.T) {
	input := map[string][]string{
		"Deployment": {"appsv1.Deployment{struct_attributes...}"},
	}
	goFileObj.Generate(input)
	if goFileObj.FileContent == "" {
		t.Errorf("Generate GoCode Failed| Unable to Generate Go-File from Go-Code")
	}
}
