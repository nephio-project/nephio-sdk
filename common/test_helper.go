package common

import (
	"bufio"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

type Tests struct {
	input    interface{}
	expected interface{}
}

/*
Reads the File from FilePath and returns the file-data
*/
func getFileContents(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	fp := bufio.NewReader(file)
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	return data, nil

}

/*
Convert the KRM Resources to *unstructured.Unstructured (map[string]interface{})
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

// func compare2Strings(a string, b string) {

// 	lenA, lenB := len(a), len(b)
// 	fmt.Println(lenA, lenB)
// 	minL := lenA
// 	if lenB < lenA {
// 		minL = lenB
// 	}

// 	for index := 0; index < minL; index++ {
// 		fmt.Printf("%d %c %c\n", index, a[index], b[index])
// 	}
// }
