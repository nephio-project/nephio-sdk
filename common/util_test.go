package common

import (
	"os"
	"reflect"
	"testing"
)

func TestRepeat(t *testing.T) {
	result := repeat("a", 3)
	if result != "aaa" {
		t.Errorf("Util-tests | 'Repeat' test failed | Expected %s | Got %s", "aaa", result)
	}
}

func TestCreateDirIfDontExist(t *testing.T) {
	err := createDirIfDontExist("tests/test_createdir")
	if err != nil {
		t.Errorf("Util-tests | 'CreateDirIfDontExist' test failed | %s", err)
	}

	// Check If directory is created or not
	folderContent, _ := os.ReadDir("tests")
	testPassed := false
	for _, files := range folderContent {
		if files.Type().IsDir() {
			if files.Name() == "test_createdir" {
				testPassed = true
				break
			}
		}
	}
	if !testPassed {
		t.Errorf("Util-tests | 'CreateDirIfDontExist' test failed | Directory Not Found where expected")
	}
	os.RemoveAll("tests/test_createdir")
}

func TestRecursiveListYamls(t *testing.T) {
	result := RecursiveListYamls("tests")
	expected := []string{"tests/test-yamls/deployment.yaml", "tests/test-yamls/third-party-cr.yaml"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Util-tests | 'RecursiveListYamls' test failed | \n Expected %v \n Got %v", expected, result)
	}

}

func TestHandleMultiLineStrings(t *testing.T) {
	input := "abc\nd"
	result := handleMultiLineStrings(input)
	expected := "\"abc\\n\" + \n \"d\"" // Expected "abc\n" + \n "d" where Second \n Represents actual new line
	if result != expected {
		t.Errorf("Util-tests | 'HandleMultiLineStrings' test failed | \n Expected %v \n Got %v", expected, result)
	}
}
